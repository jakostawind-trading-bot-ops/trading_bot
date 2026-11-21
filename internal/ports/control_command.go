package ports

import (
	"context"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/commands"
)

// ControlCommandDecoderPort преобразует версионное сообщение во внутреннюю команду.
type ControlCommandDecoderPort interface {
	// Decode выбирает контракт по MessageType и проверяет JSON из Data.
	Decode(MessageType string, Data []byte) (commands.BaseControlCommandI, error)
}

// ControlCommandDispatcherPort передаёт команду прикладному обработчику.
type ControlCommandDispatcherPort interface {
	// Dispatch возвращает результат выполнения Command в контексте обработки.
	Dispatch(Context context.Context, Command commands.BaseControlCommandI) error
}

// ControlCommandSubscriberPort принимает команды одного экземпляра бота.
type ControlCommandSubscriberPort interface {
	// Run работает до отмены ReceiveContext. HandlerContext позволяет завершить
	// уже принятую команду в пределах срока остановки Fx.
	Run(ReceiveContext, HandlerContext context.Context)
}
