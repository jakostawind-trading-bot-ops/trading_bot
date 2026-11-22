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

type controlCommandSequenceKey struct{}

// WithControlCommandSequence связывает Context с номером исходного сообщения в stream.
func WithControlCommandSequence(Context context.Context, Sequence uint64) context.Context {
	return context.WithValue(Context, controlCommandSequenceKey{}, Sequence)
}

// ControlCommandSequence возвращает общий для всех доставок номер сообщения или ноль при прямом вызове.
func ControlCommandSequence(Context context.Context) uint64 {
	Sequence, _ := Context.Value(controlCommandSequenceKey{}).(uint64)
	return Sequence
}
