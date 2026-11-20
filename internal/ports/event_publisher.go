package ports

import (
	"context"
	"reflect"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/events"
)

// EventPublisherRoute связывает внутреннее событие с кодированием сообщения.
type EventPublisherRoute struct {
	EventType   reflect.Type // Конкретный Go-тип события, включая указатель.
	MessageType string       // Версия и тип сообщения: v1.lifecycle.bot_started_event.
	// Encode проверяет event, собирает и валидирует контракт, возвращает JSON.
	// ctx позволяет отклонить отменённый вызов; Marshal выполняется синхронно.
	Encode func(ctx context.Context, event events.BaseBotEventI) ([]byte, error)
}

type EventPublisherPort interface {
	PublishEvent(ctx context.Context, Event events.BaseBotEventI) error
}
