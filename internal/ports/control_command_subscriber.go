package ports

import (
	"context"
)

// ControlCommandSubscriberPort принимает команды одного экземпляра бота.
type ControlCommandSubscriberPort interface {
	// Run работает до отмены ReceiveContext. HandlerContext позволяет завершить
	// уже принятую команду в пределах срока остановки Fx.
	Run(ReceiveContext, HandlerContext context.Context)
}
