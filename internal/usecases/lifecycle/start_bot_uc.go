package usecases_lifecycle

import (
	"context"
	"log/slog"

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

// Execute публикует событие запуска и записывает ошибку публикации в лог.
func (StartBotUC *StartBotUC) Execute(ctx context.Context) error {
	Event := &events.BotStartedEvent{
		BaseBotEvent: events.BaseBotEvent{
			BotID: StartBotUC.GeneralState.BotID,
		},
		BotStatus: StartBotUC.GeneralState.GetBotStatus(),
	}
	Logger := slog.With("operation", "start_bot", "bot_id", Event.BotID)
	Logger.DebugContext(ctx, "Публикация BotStartedEvent", "bot_status", Event.BotStatus)
	if err := StartBotUC.EventPublisherPort.PublishEvent(ctx, Event); err != nil {
		Logger.ErrorContext(ctx, "Не удалось опубликовать BotStartedEvent", "error", err)
	}

	return nil
}
