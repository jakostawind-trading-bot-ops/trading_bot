package di_lifecycle

import (
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/lifecycle"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/states"
)

// LifecycleProvider создаёт Lifecycle, который читает и меняет статус бота через generalState.
func LifecycleProvider(GeneralState *states.GeneralState) *lifecycle.Lifecycle {
	return &lifecycle.Lifecycle{GeneralState: GeneralState}
}
