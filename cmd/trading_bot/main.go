package main

import (
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/di"
	di_states "github.com/jakostawind-trading-bot-ops/trading_bot/internal/di/states"
	di_nats "github.com/jakostawind-trading-bot-ops/trading_bot/internal/di/transport/nats"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/ports"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/states"
	publisher_routes "github.com/jakostawind-trading-bot-ops/trading_bot/internal/transport/nats/publisher_routes"
	"go.uber.org/fx"
)

// main собирает publisher и запускает lifecycle соединения через Fx.
// События публикуются потребителями порта; сам запуск приложения их не создаёт.
func main() {
	Settings := internal.LoadBootstrapSettings()
	App := fx.New(
		di.BootstrapSettingsProvider(&Settings),
		fx.Provide(di_nats.NatsClientProvider),
		fx.Provide(
			di_states.GeneralStateProvider,
			publisher_routes.EventPublisherRoutesRegistry,
			di_nats.NatsEventPublisherProvider,
		),
		// Создаём состояние бота и publisher с NatsClient при запуске.
		fx.Invoke(func(_ *states.GeneralState, _ ports.EventPublisherPort) {}),
	)
	App.Run()
}
