package events

import "github.com/jakostawind-trading-bot-ops/trading_bot/internal/lifecycle"

type BotStartedEvent struct {
	BaseBotEvent
	BotStatus lifecycle.LifecycleStatus
}