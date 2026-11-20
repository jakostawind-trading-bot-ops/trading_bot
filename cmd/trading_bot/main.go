package main

import (
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/di"
	di_lifecycle "github.com/jakostawind-trading-bot-ops/trading_bot/internal/di/lifecycle"
	di_states "github.com/jakostawind-trading-bot-ops/trading_bot/internal/di/states"
	di_nats "github.com/jakostawind-trading-bot-ops/trading_bot/internal/di/transport/nats"
	di_usecases "github.com/jakostawind-trading-bot-ops/trading_bot/internal/di/usecases"
	publisher_routes "github.com/jakostawind-trading-bot-ops/trading_bot/internal/transport/nats/publisher_routes"
	usecases_lifecycle "github.com/jakostawind-trading-bot-ops/trading_bot/internal/usecases/lifecycle"
	"go.uber.org/fx"
)

// main собирает зависимости бота и вызывает StartBotUC.Execute при запуске Fx.
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
		fx.Invoke(func(lc fx.Lifecycle, startBot *usecases_lifecycle.StartBotUC) {
			lc.Append(fx.Hook{OnStart: startBot.Execute})
		}),
	)
	App.Run()
}
