package main

import (
	"context"
	"log/slog"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/di"
	di_commands "github.com/jakostawind-trading-bot-ops/trading_bot/internal/di/commands"
	di_lifecycle "github.com/jakostawind-trading-bot-ops/trading_bot/internal/di/lifecycle"
	di_states "github.com/jakostawind-trading-bot-ops/trading_bot/internal/di/states"
	di_nats "github.com/jakostawind-trading-bot-ops/trading_bot/internal/di/transport/nats"
	di_usecases "github.com/jakostawind-trading-bot-ops/trading_bot/internal/di/usecases"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/ports"
	publisher_routes "github.com/jakostawind-trading-bot-ops/trading_bot/internal/transport/nats/publisher_routes"
	usecases_lifecycle "github.com/jakostawind-trading-bot-ops/trading_bot/internal/usecases/lifecycle"
	"go.opentelemetry.io/otel"
	"go.uber.org/fx"
)

// main собирает зависимости, публикует событие запуска и включает приём команд через Fx.
func main() {
	Settings := internal.LoadBootstrapSettings()
	App := fx.New(
		di.BootstrapSettingsProvider(&Settings),
		fx.Provide(di.LoggerProvider, di_nats.NatsClientProvider),
		// Потребители стартуют после SDK и останавливаются до выгрузки телеметрии.
		fx.Invoke(func(_ *slog.Logger) {}, di.RegisterTracing),
		fx.Provide(
			di_states.GeneralStateProvider,
			di_lifecycle.LifecycleProvider,
			publisher_routes.EventPublisherRoutesRegistry,
			di_nats.NatsEventPublisherProvider,
		),

		// USECASES
		fx.Provide(
			di_usecases.StartBotUCProvider,
		),
		fx.Provide(
			di_commands.ControlCommandDispatcherProvider,
			di_nats.ControlCommandDecoderProvider,
			di_nats.ControlCommandSubscriberProvider,
		),
		fx.Invoke(func(Lifecycle fx.Lifecycle, StartBot *usecases_lifecycle.StartBotUC, Logger *slog.Logger) {
			Lifecycle.Append(fx.Hook{OnStart: func(Context context.Context) error {
				Context, Span := otel.Tracer("trading_bot/cmd/trading_bot").Start(Context, "TradingBot.Start")
				StartupError := StartBot.Execute(Context)
				defer internal.FinishSpan(Span, &StartupError)
				StartupLogger := Logger.With("component", "bootstrap", "operation", "start_bot", "bot_id", StartBot.GeneralState.BotID)
				if StartupError != nil {
					StartupLogger.ErrorContext(Context, "Не удалось опубликовать событие запуска, бот продолжает работу", "error", StartupError)
				} else {
					StartupLogger.InfoContext(Context, "Событие запуска бота опубликовано")
				}
				return nil
			}})
		}),
		fx.Invoke(func(_ ports.ControlCommandSubscriberPort) {}),
	)
	App.Run()
}
