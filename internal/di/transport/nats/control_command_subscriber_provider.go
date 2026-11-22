package di_transport_nats

import (
	"context"
	"log/slog"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/commands"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/ports"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/states"
	transport "github.com/jakostawind-trading-bot-ops/trading_bot/internal/transport/nats"
	contracts "github.com/jakostawind-trading-bot-ops/trading_bot/internal/transport/nats/contracts/control/bot/v1/lifecycle"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/fx"
)

// ControlCommandDecoderProvider связывает версионный суффикс subject с декодером JSON.
func ControlCommandDecoderProvider(Logger *slog.Logger) ports.ControlCommandDecoderPort {
	return transport.ControlCommandDecoder{
		Logger: Logger.With("component", "control_command_decoder"),
		Routes: map[string]func(context.Context, []byte) (commands.BaseControlCommandI, error){
			"v1.lifecycle.start_bot_command": contracts.DecodeStartBotCommand,
		},
	}
}

// ControlCommandSubscriberProvider подключает подписчик к общему клиенту и lifecycle Fx.
// GeneralState задаёт адресата, Decoder и Dispatcher обрабатывают принятые команды.
//
//	OnStart -> Run в фоне
//	OnStop -> отмена приёма -> ожидание текущей команды в пределах StopContext
func ControlCommandSubscriberProvider(
	Lifecycle fx.Lifecycle,
	Client *NatsClient,
	Logger *slog.Logger,
	GeneralState *states.GeneralState,
	Decoder ports.ControlCommandDecoderPort,
	Dispatcher ports.ControlCommandDispatcherPort,
) (ports.ControlCommandSubscriberPort, error) {
	Connect := func(Context context.Context) (jetstream.JetStream, error) {
		Connection, ConnectionError := Client.WaitConnection(Context)
		if ConnectionError != nil {
			return nil, ConnectionError
		}
		return jetstream.New(Connection)
	}
	SubscriberLogger := Logger.With("component", "control_command_subscriber", "bot_id", GeneralState.BotID, "stream", transport.ControlCommandStream)
	Subscriber, CreationError := transport.NewControlCommandSubscriber(GeneralState.BotID, Connect, Decoder, Dispatcher, SubscriberLogger)
	if CreationError != nil {
		return nil, CreationError
	}

	// Контекст хука старта Fx завершается сразу после запуска приложения.
	var StopReceiving, StopHandling context.CancelFunc
	Done := make(chan struct{})
	Lifecycle.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			ReceiveContext, CancelReceive := context.WithCancel(context.Background())
			HandlerContext, CancelHandler := context.WithCancel(context.Background())
			StopReceiving, StopHandling = CancelReceive, CancelHandler
			go func() {
				defer close(Done)
				Subscriber.Run(ReceiveContext, HandlerContext)
			}()
			return nil
		},
		OnStop: func(StopContext context.Context) error {
			StopReceiving()
			defer StopHandling()
			select {
			case <-Done:
				return nil
			case <-StopContext.Done():
				return StopContext.Err()
			}
		},
	})
	return Subscriber, nil
}
