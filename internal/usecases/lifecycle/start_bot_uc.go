package usecases_lifecycle

import (
	"context"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/events"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/lifecycle"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/ports"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/states"
)

type StartBotUC struct {
	GeneralState *states.GeneralState
	Lifecycle lifecycle.Lifecycle
	EventPublisherPort ports.EventPublisherPort
}

func (StartBotUC *StartBotUC) Execute(ctx context.Context) error {

	Event := &events.BotStartedEvent{}
	StartBotUC.EventPublisherPort.PublishEvent(ctx, Event)

	return nil
}