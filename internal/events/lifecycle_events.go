package events

import "github.com/jakostawind-trading-bot-ops/trading_bot/internal/lifecycle"

type BotStartedEvent struct {
	BaseBotEvent
	BotStatus lifecycle.LifecycleStatus
}

type BotRunnedEvent struct {
	BaseBotEvent
	PreviousBotStatus lifecycle.LifecycleStatus
	CurrentBotStatus  lifecycle.LifecycleStatus
}

type BotRunnedErrorEvent struct {
	BaseBotEvent
	BotStatus lifecycle.LifecycleStatus
	Error     string
}
