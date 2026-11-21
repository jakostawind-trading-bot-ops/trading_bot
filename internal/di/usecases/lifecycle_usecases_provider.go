package di_usecases

import (
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/lifecycle"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/ports"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/states"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/usecases/lifecycle"
)

// StartBotUCProvider создаёт сценарий запуска бота с общим состоянием и издателем событий.
func StartBotUCProvider(
	GeneralState *states.GeneralState,
	Lifecycle *lifecycle.Lifecycle,
	EventPublisherPort ports.EventPublisherPort,
) *usecases_lifecycle.StartBotUC {
	return &usecases_lifecycle.StartBotUC{
		GeneralState:       GeneralState,
		Lifecycle:          *Lifecycle,
		EventPublisherPort: EventPublisherPort,
	}
}
