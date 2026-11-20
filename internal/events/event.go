package events

import "time"

type BaseBotEvent struct {
	BotID     string
	Payload   map[string]any
	Timestamp string
}

type BaseBotEventI interface {
	IsBaseBotEvent() *BaseBotEvent
}

func (BaseBotEvent *BaseBotEvent) IsBaseBotEvent() *BaseBotEvent {
	return BaseBotEvent
}

func (BaseBotEvent *BaseBotEvent) SetTimestamp() string {
	TimeStamp := time.Now().UTC().Format(time.RFC3339)
	return TimeStamp
}
