package ports

import (
	"context"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/commands"
)

// ControlCommandDispatcherPort передаёт команду прикладному обработчику.
type ControlCommandDispatcherPort interface {
	// Dispatch возвращает результат выполнения Command в контексте обработки.
	Dispatch(Context context.Context, Command commands.BaseControlCommandI) error
}
