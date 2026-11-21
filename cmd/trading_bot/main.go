package main

import (
	"context"

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
	"go.uber.org/fx"
)

// main собирает зависимости, публикует событие запуска и включает приём команд через Fx.
func main() {
	Settings := internal.LoadBootstrapSettings()
	App := fx.New(
		di.BootstrapSettingsProvider(&Settings),
		fx.Provide(di_nats.NatsClientProvider),
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
		fx.Invoke(func(Lifecycle fx.Lifecycle, StartBot *usecases_lifecycle.StartBotUC) {
			Lifecycle.Append(fx.Hook{OnStart: func(Context context.Context) error {
				// Ошибка публикации записана в лог. При bootstrap бот продолжает запуск.
				_ = StartBot.Execute(Context)
				return nil
			}})
		}),
		fx.Invoke(func(_ ports.ControlCommandSubscriberPort) {}),
	)
	App.Run()
}
