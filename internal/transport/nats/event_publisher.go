package transport_nats

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"time"
	"unicode"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/events"
	tracing "github.com/jakostawind-trading-bot-ops/trading_bot/internal/observability/tracing"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/ports"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
)

const PublishTimeout = 3 * time.Second // Общий предел ожидания соединения и PubAck.

// JetStreamFactory получает актуальный клиент; Context отменяет ожидание подключения.
// Возвращает JetStreamContext либо ошибку; соединением владеет NatsClient.
type JetStreamFactory func(Context context.Context) (nats.JetStreamContext, error)

// NatsEventPublisher кодирует события и публикует их с ожиданием PubAck.
// Создаётся конструктором; таблица маршрутов после создания не меняется.
type NatsEventPublisher struct {
	Logger       *slog.Logger                               // Логгер компонента и stream из DI.
	Routes       map[reflect.Type]ports.EventPublisherRoute // Маршруты по конкретному Go-типу.
	GetJetStream JetStreamFactory                           // Отложенный доступ к текущему соединению.
	StreamName   string                                     // Ожидаемый существующий stream; publisher его не создаёт.
}

var _ ports.EventPublisherPort = (*NatsEventPublisher)(nil)

// NewNatsEventPublisher проверяет Routes и собирает publisher без сетевых вызовов.
// GetJetStream — ненулевая фабрика, StreamName — имя существующего stream, Logger — логгер компонента.
// Возвращает publisher либо ошибку зависимости, маршрута или повторного типа.
//
// Схема:
//
//	зависимости -> проверка маршрутов -> неизменяемая таблица -> publisher
func NewNatsEventPublisher(
	Routes []ports.EventPublisherRoute,
	GetJetStream JetStreamFactory,
	StreamName string,
	Logger *slog.Logger,
) (*NatsEventPublisher, error) {
	if Logger == nil || GetJetStream == nil || strings.TrimSpace(StreamName) == "" {
		return nil, fmt.Errorf("publisher: нужны логгер, JetStream factory и имя stream")
	}

	// Ошибки регистрации выявляются при сборке приложения.
	RouteMap := make(map[reflect.Type]ports.EventPublisherRoute, len(Routes))
	for _, Route := range Routes {
		if Route.EventType == nil ||
			!Route.EventType.Implements(reflect.TypeFor[events.BaseBotEventI]()) ||
			Route.Encode == nil {
			return nil, fmt.Errorf("publisher: некорректный маршрут %q", Route.MessageType)
		}
		for _, Token := range strings.Split(Route.MessageType, ".") {
			if !validSubjectToken(Token) {
				return nil, fmt.Errorf("publisher: некорректный тип сообщения %q", Route.MessageType)
			}
		}
		if _, Exists := RouteMap[Route.EventType]; Exists {
			return nil, fmt.Errorf("publisher: повторный маршрут для %v", Route.EventType)
		}
		RouteMap[Route.EventType] = Route
	}

	return &NatsEventPublisher{
		Logger:       Logger,
		Routes:       RouteMap,
		GetJetStream: GetJetStream,
		StreamName:   StreamName,
	}, nil
}

// PublishEvent кодирует Event и ждёт подтверждения сохранения в JetStream.
// Context должен быть ненулевым; ограничивает ожидание соединения и PubAck.
// Заполняет отсутствующие Timestamp и Payload в Event.
// Event нельзя изменять или публиковать параллельно до возврата метода.
// Возвращает ошибку события, маршрута, кодирования или транспорта.
// Соединение и stream не создаёт самостоятельно; состояние бота не изменяет.
//
// Схема:
//
//	Event -> маршрут -> JSON и subject -> JetStream -> Publish -> PubAck
func (Publisher *NatsEventPublisher) PublishEvent(
	Context context.Context,
	Event events.BaseBotEventI,
) (PublishError error) {
	Context, Span := otel.Tracer("trading_bot/internal/transport/nats").Start(Context, "NatsEventPublisher.PublishEvent")
	defer tracing.FinishSpan(Span, &PublishError)
	Logger := Publisher.Logger.With("operation", "publish_bot_event", "event_type", fmt.Sprintf("%T", Event))
	Logger.DebugContext(Context, "Поиск маршрута публикации")
	// Проверяем событие до вызова его методов.
	if ContextError := Context.Err(); ContextError != nil {
		return ContextError
	}
	Value := reflect.ValueOf(Event)
	if !Value.IsValid() || (Value.Kind() == reflect.Pointer && Value.IsNil()) {
		return fmt.Errorf("publisher: событие отсутствует")
	}
	Route, Exists := Publisher.Routes[Value.Type()]
	if !Exists {
		return fmt.Errorf("publisher: маршрут для %T не зарегистрирован", Event)
	}

	Base := Event.IsBaseBotEvent()
	if Base == nil || !validSubjectToken(Base.BotID) {
		return fmt.Errorf("publisher: BotID должен быть одним непустым токеном subject")
	}
	if strings.TrimSpace(Base.Timestamp) == "" {
		Base.Timestamp = Base.SetTimestamp()
	}
	if Base.Payload == nil {
		Base.Payload = map[string]any{}
	}

	Logger = Logger.With("bot_id", Base.BotID, "message_type", Route.MessageType)
	Logger.DebugContext(Context, "Кодирование события")
	// Кодирование выполняется до ожидания транспорта.
	Data, EncodeError := Route.Encode(Context, Event)
	if EncodeError != nil {
		return fmt.Errorf("publisher: %s: %w", Route.MessageType, EncodeError)
	}
	Subject := "bot." + Base.BotID + ".to.control.event." + Route.MessageType

	// Каждый вызов получает актуальное соединение после возможного переподключения.
	PublishContext, Cancel := context.WithTimeout(Context, PublishTimeout)
	defer Cancel()
	Logger.DebugContext(PublishContext, "Ожидание JetStream", "subject", Subject)
	JetStream, ConnectionError := Publisher.GetJetStream(PublishContext)
	if ConnectionError != nil {
		return fmt.Errorf("publisher: получение JetStream: %w", ConnectionError)
	}
	if JetStream == nil {
		return fmt.Errorf("publisher: фабрика вернула пустой JetStream")
	}
	Logger.DebugContext(PublishContext, "Отправка события в JetStream", "subject", Subject, "payload_size", len(Data))
	Ack, PublishError := JetStream.Publish(
		Subject, Data,
		nats.Context(PublishContext),
		nats.ExpectStream(Publisher.StreamName),
	)
	if PublishError != nil {
		return fmt.Errorf("publisher: публикация %s: %w", Subject, PublishError)
	}
	if Ack != nil {
		Logger.DebugContext(PublishContext, "JetStream подтвердил публикацию",
			"subject", Subject, "ack_stream", Ack.Stream, "sequence", Ack.Sequence)
	}
	return nil
}

// validSubjectToken проверяет Token: запрещены пустота, разделитель,
// wildcard, пробельные и управляющие символы. Возвращает допустимость токена.
func validSubjectToken(Token string) bool {
	return Token != "" && !strings.ContainsAny(Token, ".*>") &&
		strings.IndexFunc(Token, func(Character rune) bool {
			return unicode.IsSpace(Character) || unicode.IsControl(Character)
		}) == -1
}
