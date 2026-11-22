package ports

import (
	"context"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/commands"
)

// ControlCommandDecoderPort преобразует версионное сообщение во внутреннюю команду.
type ControlCommandDecoderPort interface {
	// Decode выбирает контракт по MessageType и проверяет JSON из Data.
	Decode(Context context.Context, MessageType string, Data []byte) (commands.BaseControlCommandI, error)
}
