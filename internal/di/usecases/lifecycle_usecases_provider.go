package di_usecases

import (
	"log/slog"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/lifecycle"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/ports"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/states"
	usecases_lifecycle "github.com/jakostawind-trading-bot-ops/trading_bot/internal/usecases/lifecycle"
)

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

func RunBotUCProvider(
	Logger *slog.Logger,
	GeneralState *states.GeneralState,
	Lifecycle *lifecycle.Lifecycle,
	EventPublisherPort ports.EventPublisherPort,
) *usecases_lifecycle.RunBotUC {
	return &usecases_lifecycle.RunBotUC{
		Logger:             Logger.With("component", "run_bot_use_case", "bot_id", GeneralState.BotID),
		GeneralState:       GeneralState,
		Lifecycle:          *Lifecycle,
		EventPublisherPort: EventPublisherPort,
	}
}
