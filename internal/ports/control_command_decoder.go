package ports

import (
	"context"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/commands"
)

type ControlCommandDecoderRoute struct {
	MessageType string // Ключ маршрута, включающий версию и тип транспортного сообщения.
	// Decode — обязательная функция обработки RawMessage в формате своего контракта.
	// На каждом вызове она разбирает и проверяет отдельное сообщение, затем
	// возвращает внутреннюю команду либо ошибку разбора, валидации или преобразования.
	Decode func(Context context.Context, RawMessage []byte) (commands.BaseControlCommandI, error)
}

// ControlCommandDecoderPort преобразует версионное сообщение во внутреннюю команду.
type ControlCommandDecoderPort interface {
	// Decode выбирает контракт по MessageType и проверяет JSON из Data.
	Decode(Context context.Context, MessageType string, Data []byte) (commands.BaseControlCommandI, error)
}
