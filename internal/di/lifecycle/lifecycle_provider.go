package di_lifecycle

import (
	"log/slog"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/lifecycle"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/states"
)

// LifecycleProvider создаёт Lifecycle, который читает и меняет статус бота через generalState.
func LifecycleProvider(GeneralState *states.GeneralState, Logger *slog.Logger) *lifecycle.Lifecycle {
	return &lifecycle.Lifecycle{GeneralState: GeneralState, Logger: Logger.With("component", "lifecycle", "bot_id", GeneralState.BotID)}
}
