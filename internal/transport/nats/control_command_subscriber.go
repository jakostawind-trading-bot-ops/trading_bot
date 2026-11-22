package transport_nats

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/commands"
	tracing "github.com/jakostawind-trading-bot-ops/trading_bot/internal/observability/tracing"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/ports"
	"github.com/nats-io/nats.go/jetstream"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

const ControlCommandStream = "CONTROL_TO_BOT_COMMANDS" // Stream из Terraform.

// ControlCommandSubscriber читает команды одного BotID последовательно.
// Один вызов Run владеет итератором. Общим соединением управляет NatsClient.
//
//	JetStream -> декодер -> диспетчер -> Ack
type ControlCommandSubscriber struct {
	Logger     *slog.Logger                                       // Постоянные поля компонента, stream и адресата.
	BotID      string                                             // ID из GeneralState, используемый также издателем событий.
	Connect    func(context.Context) (jetstream.JetStream, error) // Текущее общее соединение.
	Decoder    ports.ControlCommandDecoderPort                    // Реестр транспортных контрактов.
	Dispatcher ports.ControlCommandDispatcherPort                 // Реестр прикладных обработчиков.
}

// NewControlCommandSubscriber проверяет BotID для subject и имени consumer.
// Connect, Decoder, Dispatcher и Logger задаются при сборке приложения до запуска Run.
func NewControlCommandSubscriber(
	BotID string,
	Connect func(context.Context) (jetstream.JetStream, error),
	Decoder ports.ControlCommandDecoderPort,
	Dispatcher ports.ControlCommandDispatcherPort,
	Logger *slog.Logger,
) (*ControlCommandSubscriber, error) {
	if !validSubjectToken(BotID) || strings.ContainsAny(BotID, `/\`) {
		return nil, fmt.Errorf("недопустимый BotID для подписки команд")
	}
	if Connect == nil || Decoder == nil || Dispatcher == nil || Logger == nil {
		return nil, fmt.Errorf("нужны соединение, логгер, декодер и диспетчер команд")
	}
	return &ControlCommandSubscriber{BotID: BotID, Connect: Connect, Decoder: Decoder, Dispatcher: Dispatcher, Logger: Logger}, nil
}

// Run открывает consumer после подключения, при сбое повторяет попытку.
// ReceiveContext останавливает приём, HandlerContext ограничивает обработку текущей команды.
//
//	подключение -> consumer -> чтение -> сбой -> пауза -> подключение
func (Subscriber *ControlCommandSubscriber) Run(ReceiveContext, HandlerContext context.Context) {
	Logger := Subscriber.Logger.With("operation", "receive_control_commands")
	Logger.DebugContext(ReceiveContext, "Запуск приёма команд")
	defer Logger.InfoContext(ReceiveContext, "Приём команд остановлен")
	for ReceiveContext.Err() == nil {
		ReceiveError := Subscriber.receive(ReceiveContext, HandlerContext)
		if ReceiveContext.Err() != nil {
			return
		}
		Logger.WarnContext(ReceiveContext, "Повтор подключения подписчика команд", "error", ReceiveError, "retry_after", 2*time.Second)
		select {
		case <-ReceiveContext.Done():
			return
		case <-time.After(2 * time.Second):
		}
	}
}

// receive открывает или находит consumer с точным фильтром BotID. Случайный ID
// живёт один запуск бота. Срок очистки consumer превышает 30 дней хранения команд,
// чтобы обычное переподключение могло продолжить чтение с подтверждённой позиции.
//
//	соединение -> consumer -> Next -> выполнение -> Next
func (Subscriber *ControlCommandSubscriber) receive(ReceiveContext, HandlerContext context.Context) error {
	// Спан открытия завершаем до ожидания первого сообщения.
	JetStream, Consumer, OpenError := Subscriber.openConsumer(ReceiveContext)
	if OpenError != nil {
		return OpenError
	}

	ConsumerName := "bot_command_" + Subscriber.BotID
	// После штатной остановки этот процесс уже не использует свой случайный BotID.
	// При разрыве соединения consumer нужен для продолжения с подтверждённой позиции.
	defer func() {
		if ReceiveContext.Err() == nil || HandlerContext.Err() != nil {
			return
		}
		CleanupContext, CancelCleanup := context.WithTimeout(HandlerContext, 2*time.Second)
		defer CancelCleanup()
		CleanupContext, Span := otel.Tracer("trading_bot/internal/transport/nats").Start(CleanupContext, "ControlCommandSubscriber.DeleteConsumer")
		defer Span.End()
		if CleanupError := JetStream.DeleteConsumer(CleanupContext, ControlCommandStream, ConsumerName); CleanupError != nil {
			Span.RecordError(CleanupError)
			Span.SetStatus(codes.Error, CleanupError.Error())
			Subscriber.Logger.WarnContext(CleanupContext, "Не удалось удалить consumer завершённого экземпляра", "operation", "delete_consumer", "consumer", ConsumerName, "error", CleanupError)
		}
	}()

	Messages, IteratorError := Consumer.Messages(jetstream.PullMaxMessages(1), jetstream.WithMessagesErrOnMissingHeartbeat(true))
	if IteratorError != nil {
		return IteratorError
	}
	defer Messages.Stop()

	// Обрабатываем сообщения в независимом HandlerContext.
	for ReceiveContext.Err() == nil {
		Message, ReadError := Messages.Next(jetstream.NextContext(ReceiveContext))
		if ReadError != nil {
			return ReadError
		}
		Subscriber.handleMessage(HandlerContext, Message)
	}
	return ReceiveContext.Err()
}

// openConsumer получает соединение и открывает персональный consumer в пределах ReceiveContext.
func (Subscriber *ControlCommandSubscriber) openConsumer(ReceiveContext context.Context) (Connection jetstream.JetStream, OpenedConsumer jetstream.Consumer, OpenError error) {
	ReceiveContext, Span := otel.Tracer("trading_bot/internal/transport/nats").Start(ReceiveContext, "ControlCommandSubscriber.OpenConsumer")
	defer tracing.FinishSpan(Span, &OpenError)
	ConsumerName := "bot_command_" + Subscriber.BotID
	Logger := Subscriber.Logger.With("operation", "open_consumer", "consumer", ConsumerName)
	Logger.DebugContext(ReceiveContext, "Ожидание соединения для подписки")
	JetStream, ConnectError := Subscriber.Connect(ReceiveContext)
	if ConnectError != nil {
		return nil, nil, ConnectError
	}

	// CreateConsumer повторно открывает идентичную конфигурацию и отклоняет конфликт.
	RequestContext, CancelRequest := context.WithTimeout(ReceiveContext, 3*time.Second)
	Logger.DebugContext(ReceiveContext, "Открытие consumer команд")
	Consumer, ConsumerError := JetStream.CreateConsumer(RequestContext, ControlCommandStream, jetstream.ConsumerConfig{
		Durable:           ConsumerName,
		FilterSubject:     Subscriber.subjectPrefix() + ">",
		AckPolicy:         jetstream.AckExplicitPolicy,
		DeliverPolicy:     jetstream.DeliverAllPolicy,
		AckWait:           30 * time.Second,
		MaxAckPending:     1,
		InactiveThreshold: 31 * 24 * time.Hour,
	})
	CancelRequest()
	if ConsumerError != nil {
		return nil, nil, ConsumerError
	}
	Logger.InfoContext(ReceiveContext, "Consumer команд открыт")
	return JetStream, Consumer, nil
}

// subjectPrefix возвращает префикс персональной подписки с конечной точкой.
func (Subscriber *ControlCommandSubscriber) subjectPrefix() string {
	return "control.to.bot." + Subscriber.BotID + ".command."
}

// handleMessage проверяет адресата, вызывает диспетчер и подтверждает успешное выполнение.
// При ошибке обработчика JetStream может повторить доставку по AckWait.
//
//	subject и JSON -> ошибка контракта -> Term
//	валидная команда -> Dispatch -> успех -> Ack
func (Subscriber *ControlCommandSubscriber) handleMessage(Context context.Context, Message jetstream.Msg) {
	Context, Span := otel.Tracer("trading_bot/internal/transport/nats").Start(Context, "ControlCommandSubscriber.HandleMessage")
	var HandlingError error
	defer tracing.FinishSpan(Span, &HandlingError)
	Started := time.Now()
	Logger := Subscriber.Logger.With("operation", "handle_control_command", "subject", Message.Subject())
	Logger.DebugContext(Context, "Получена команда", "payload_size", len(Message.Data()))
	Command, DecodeError := Subscriber.decodeMessage(Context, Message)
	if DecodeError != nil {
		HandlingError = DecodeError
		Logger.WarnContext(Context, "Команда отклонена", "error", DecodeError)
		if TermError := Message.Term(); TermError != nil {
			Span.RecordError(DecodeError)
			HandlingError = TermError
			Logger.ErrorContext(Context, "Ошибка Term команды", "error", TermError)
		}
		return
	}
	Logger = Logger.With("command_type", fmt.Sprintf("%T", Command))
	Logger.DebugContext(Context, "Передача команды диспетчеру")
	if DispatchError := Subscriber.Dispatcher.Dispatch(Context, Command); DispatchError != nil {
		HandlingError = DispatchError
		if errors.Is(DispatchError, context.Canceled) {
			Logger.DebugContext(Context, "Обработка команды отменена", "error", DispatchError)
		} else {
			Logger.ErrorContext(Context, "Ошибка выполнения команды, ожидается повторная доставка", "error", DispatchError)
		}
		return
	}
	Logger.DebugContext(Context, "Отправка Ack команды")
	if AckError := Message.Ack(); AckError != nil {
		HandlingError = AckError
		Logger.ErrorContext(Context, "Ошибка Ack команды", "error", AckError)
		return
	}
	Logger.InfoContext(Context, "Команда обработана, Ack отправлен", "duration_ms", time.Since(Started).Milliseconds())
}

// decodeMessage проверяет subject и после декодирования сверяет адресата из JSON.
func (Subscriber *ControlCommandSubscriber) decodeMessage(Context context.Context, Message jetstream.Msg) (Decoded commands.BaseControlCommandI, ValidationError error) {
	Context, Span := otel.Tracer("trading_bot/internal/transport/nats").Start(Context, "ControlCommandSubscriber.DecodeMessage")
	defer tracing.FinishSpan(Span, &ValidationError)
	MessageType, Matches := strings.CutPrefix(Message.Subject(), Subscriber.subjectPrefix())
	if !Matches || MessageType == "" {
		return nil, fmt.Errorf("неверный subject команды")
	}
	Command, DecodeError := Subscriber.Decoder.Decode(Context, MessageType, Message.Data())
	if DecodeError != nil {
		return nil, DecodeError
	}
	if Command.IsBaseControlCommand().BotID != Subscriber.BotID {
		return nil, fmt.Errorf("bot_id команды не совпадает с адресатом")
	}
	return Command, nil
}
