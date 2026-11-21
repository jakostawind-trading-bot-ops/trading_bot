package transport_nats

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/commands"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/ports"
	"github.com/nats-io/nats.go/jetstream"
)

const ControlCommandStream = "CONTROL_TO_BOT_COMMANDS" // Stream из Terraform.

// ControlCommandSubscriber читает команды одного BotID последовательно.
// Один вызов Run владеет итератором. Общим соединением управляет NatsClient.
//
//	JetStream -> декодер -> диспетчер -> Ack
type ControlCommandSubscriber struct {
	BotID      string                                             // ID из GeneralState, используемый также издателем событий.
	Connect    func(context.Context) (jetstream.JetStream, error) // Текущее общее соединение.
	Decoder    ports.ControlCommandDecoderPort                    // Реестр транспортных контрактов.
	Dispatcher ports.ControlCommandDispatcherPort                 // Реестр прикладных обработчиков.
}

// NewControlCommandSubscriber проверяет BotID для subject и имени consumer.
// Connect, Decoder и Dispatcher задаются при сборке приложения до запуска Run.
func NewControlCommandSubscriber(
	BotID string,
	Connect func(context.Context) (jetstream.JetStream, error),
	Decoder ports.ControlCommandDecoderPort,
	Dispatcher ports.ControlCommandDispatcherPort,
) (*ControlCommandSubscriber, error) {
	if !validSubjectToken(BotID) || strings.ContainsAny(BotID, `/\`) {
		return nil, fmt.Errorf("недопустимый BotID для подписки команд")
	}
	if Connect == nil || Decoder == nil || Dispatcher == nil {
		return nil, fmt.Errorf("нужны соединение, декодер и диспетчер команд")
	}
	return &ControlCommandSubscriber{BotID: BotID, Connect: Connect, Decoder: Decoder, Dispatcher: Dispatcher}, nil
}

// Run открывает consumer после подключения, при сбое повторяет попытку.
// ReceiveContext останавливает приём, HandlerContext ограничивает обработку текущей команды.
//
//	подключение -> consumer -> чтение -> сбой -> пауза -> подключение
func (Subscriber *ControlCommandSubscriber) Run(ReceiveContext, HandlerContext context.Context) {
	for ReceiveContext.Err() == nil {
		ReceiveError := Subscriber.receive(ReceiveContext, HandlerContext)
		if ReceiveContext.Err() != nil {
			return
		}
		slog.Warn("Повтор подключения подписчика команд", "bot_id", Subscriber.BotID, "error", ReceiveError)
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
	JetStream, ConnectError := Subscriber.Connect(ReceiveContext)
	if ConnectError != nil {
		return ConnectError
	}

	// CreateConsumer повторно открывает идентичную конфигурацию и отклоняет конфликт.
	RequestContext, CancelRequest := context.WithTimeout(ReceiveContext, 3*time.Second)
	ConsumerName := "bot_command_" + Subscriber.BotID
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
		return ConsumerError
	}
	// После штатной остановки этот процесс уже не использует свой случайный BotID.
	// При разрыве соединения consumer нужен для продолжения с подтверждённой позиции.
	defer func() {
		if ReceiveContext.Err() == nil || HandlerContext.Err() != nil {
			return
		}
		CleanupContext, CancelCleanup := context.WithTimeout(HandlerContext, 2*time.Second)
		defer CancelCleanup()
		if CleanupError := JetStream.DeleteConsumer(CleanupContext, ControlCommandStream, ConsumerName); CleanupError != nil {
			slog.Warn("Не удалось удалить consumer завершённого экземпляра", "consumer", ConsumerName, "error", CleanupError)
		}
	}()

	Messages, IteratorError := Consumer.Messages(jetstream.PullMaxMessages(1), jetstream.WithMessagesErrOnMissingHeartbeat(true))
	if IteratorError != nil {
		return IteratorError
	}
	defer Messages.Stop()

	// Одна команда в работе: следующие команды остаются на стороне брокера.
	for ReceiveContext.Err() == nil {
		Message, ReadError := Messages.Next(jetstream.NextContext(ReceiveContext))
		if ReadError != nil {
			return ReadError
		}
		Subscriber.handleMessage(HandlerContext, Message)
	}
	return ReceiveContext.Err()
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
	Logger := slog.With("subject", Message.Subject(), "bot_id", Subscriber.BotID)
	Command, DecodeError := Subscriber.decodeMessage(Message)
	if DecodeError != nil {
		Logger.WarnContext(Context, "Команда отклонена", "error", DecodeError)
		if TermError := Message.Term(); TermError != nil {
			Logger.ErrorContext(Context, "Ошибка Term команды", "error", TermError)
		}
		return
	}
	if DispatchError := Subscriber.Dispatcher.Dispatch(Context, Command); DispatchError != nil {
		Logger.WarnContext(Context, "Ошибка выполнения команды, ожидается повторная доставка", "error", DispatchError)
		return
	}
	if AckError := Message.Ack(); AckError != nil {
		Logger.ErrorContext(Context, "Ошибка Ack команды", "error", AckError)
	}
}

// decodeMessage проверяет subject и после декодирования сверяет адресата из JSON.
func (Subscriber *ControlCommandSubscriber) decodeMessage(Message jetstream.Msg) (commands.BaseControlCommandI, error) {
	MessageType, Matches := strings.CutPrefix(Message.Subject(), Subscriber.subjectPrefix())
	if !Matches || MessageType == "" {
		return nil, fmt.Errorf("неверный subject команды")
	}
	Command, DecodeError := Subscriber.Decoder.Decode(MessageType, Message.Data())
	if DecodeError != nil {
		return nil, DecodeError
	}
	if Command.IsBaseControlCommand().BotID != Subscriber.BotID {
		return nil, fmt.Errorf("bot_id команды не совпадает с адресатом")
	}
	return Command, nil
}
