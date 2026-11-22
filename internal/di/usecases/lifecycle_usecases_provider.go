package di_usecases

import (
	"log/slog"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/lifecycle"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/ports"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/states"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/usecases/lifecycle"
)

// StartBotUCProvider создаёт сценарий запуска бота с общим состоянием и издателем событий.
func StartBotUCProvider(
	Logger *slog.Logger,
	GeneralState *states.GeneralState,
	Lifecycle *lifecycle.Lifecycle,
	EventPublisherPort ports.EventPublisherPort,
) *usecases_lifecycle.StartBotUC {
	return &usecases_lifecycle.StartBotUC{
		Logger:             Logger.With("component", "start_bot_use_case", "bot_id", GeneralState.BotID),
		GeneralState:       GeneralState,
		Lifecycle:          *Lifecycle,
		EventPublisherPort: EventPublisherPort,
	}
}
