package ports

import (
	"context"

	events_bot "github.com/jakostawind-trading-bot-ops/trading_bot/internal/events"
)

type EventPublisherPort interface {
	PublishEvent(ctx context.Context, Event events_bot.BaseBotEventI)
}