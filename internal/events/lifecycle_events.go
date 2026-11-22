package events

import "github.com/jakostawind-trading-bot-ops/trading_bot/internal/lifecycle"

type BotStartedEvent struct {
	BaseBotEvent
	BotStatus lifecycle.LifecycleStatus
}

type BotStatusChangedEvent struct {
	BaseBotEvent
	PreviousBotStatus lifecycle.LifecycleStatus
	CurrentBotStatus lifecycle.LifecycleStatus
}

type BotStatusChangedErrorEvent struct {
	BaseBotEvent
	BotStatus lifecycle.LifecycleStatus
	RequestedBotStatus lifecycle.LifecycleStatus
	ErrorBotStatus lifecycle.LifecycleStatus
}