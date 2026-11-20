package transport_nats

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unicode"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/events"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/ports"
	"github.com/nats-io/nats.go"
)

const publishTimeout = 3 * time.Second // Общий предел ожидания соединения и PubAck.

// JetStreamFactory получает актуальный клиент; ctx отменяет ожидание подключения.
// Возвращает JetStreamContext либо ошибку; соединением владеет NatsClient.
type JetStreamFactory func(ctx context.Context) (nats.JetStreamContext, error)

// NatsEventPublisher кодирует события и публикует их с ожиданием PubAck.
// Создаётся конструктором; таблица маршрутов после создания не меняется.
type NatsEventPublisher struct {
	routes       map[reflect.Type]ports.EventPublisherRoute // Маршруты по конкретному Go-типу.
	getJetStream JetStreamFactory                           // Отложенный доступ к текущему соединению.
	streamName   string                                     // Ожидаемый существующий stream; publisher его не создаёт.
}

var _ ports.EventPublisherPort = (*NatsEventPublisher)(nil)

// NewNatsEventPublisher проверяет Routes и собирает publisher без сетевых вызовов.
// GetJetStream — ненулевая фабрика; StreamName — имя существующего stream.
// Возвращает publisher либо ошибку зависимости, маршрута или повторного типа.
//
// Схема:
//
//	зависимости -> проверка маршрутов -> неизменяемая таблица -> publisher
func NewNatsEventPublisher(
	Routes []ports.EventPublisherRoute,
	GetJetStream JetStreamFactory,
	StreamName string,
) (*NatsEventPublisher, error) {
	if GetJetStream == nil || strings.TrimSpace(StreamName) == "" {
		return nil, fmt.Errorf("publisher: нужны JetStream factory и имя stream")
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
		routes:       RouteMap,
		getJetStream: GetJetStream,
		streamName:   StreamName,
	}, nil
}

// PublishEvent кодирует Event и ждёт подтверждения сохранения в JetStream.
// ctx должен быть ненулевым; ограничивает ожидание соединения и PubAck.
// Event нельзя изменять до возврата метода.
// Возвращает ошибку события, маршрута, кодирования или транспорта.
// Соединение и stream не создаёт самостоятельно; состояние бота не изменяет.
//
// Схема:
//
//	Event -> маршрут -> JSON и subject -> JetStream -> Publish -> PubAck
func (Publisher *NatsEventPublisher) PublishEvent(
	ctx context.Context,
	Event events.BaseBotEventI,
) error {
	// Проверяем событие до вызова его методов.
	if err := ctx.Err(); err != nil {
		return err
	}
	Value := reflect.ValueOf(Event)
	if !Value.IsValid() || (Value.Kind() == reflect.Pointer && Value.IsNil()) {
		return fmt.Errorf("publisher: событие отсутствует")
	}
	Route, Exists := Publisher.routes[Value.Type()]
	if !Exists {
		return fmt.Errorf("publisher: маршрут для %T не зарегистрирован", Event)
	}

	// Кодирование выполняется до ожидания транспорта.
	Data, err := Route.Encode(ctx, Event)
	if err != nil {
		return fmt.Errorf("publisher: %s: %w", Route.MessageType, err)
	}
	Base := Event.IsBaseBotEvent()
	if Base == nil || !validSubjectToken(Base.BotID) {
		return fmt.Errorf("publisher: BotID должен быть одним непустым токеном subject")
	}
	Subject := "bot." + Base.BotID + ".to.control.event." + Route.MessageType

	// Каждый вызов получает актуальное соединение после возможного переподключения.
	PublishCtx, Cancel := context.WithTimeout(ctx, publishTimeout)
	defer Cancel()
	JetStream, err := Publisher.getJetStream(PublishCtx)
	if err != nil {
		return fmt.Errorf("publisher: получение JetStream: %w", err)
	}
	if JetStream == nil {
		return fmt.Errorf("publisher: фабрика вернула пустой JetStream")
	}
	_, err = JetStream.Publish(
		Subject, Data,
		nats.Context(PublishCtx),
		nats.ExpectStream(Publisher.streamName),
	)
	if err != nil {
		return fmt.Errorf("publisher: публикация %s: %w", Subject, err)
	}
	return nil
}

// validSubjectToken проверяет Token: запрещены пустота, разделитель,
// wildcard, пробельные и управляющие символы. Возвращает допустимость токена.
func validSubjectToken(Token string) bool {
	return Token != "" && !strings.ContainsAny(Token, ".*>") &&
		strings.IndexFunc(Token, func(r rune) bool {
			return unicode.IsSpace(r) || unicode.IsControl(r)
		}) == -1
}
