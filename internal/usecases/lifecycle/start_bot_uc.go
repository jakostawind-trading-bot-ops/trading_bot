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
	GeneralState       *states.GeneralState
	Lifecycle          lifecycle.Lifecycle
	EventPublisherPort ports.EventPublisherPort
}

// Execute публикует событие запуска. Возвращает ошибку публикации подписчику команд.
func (StartBotUC *StartBotUC) Execute(Context context.Context) error {
	Event := &events.BotStartedEvent{
		BaseBotEvent: events.BaseBotEvent{
			BotID: StartBotUC.GeneralState.BotID,
		},
		BotStatus: StartBotUC.GeneralState.GetBotStatus(),
	}
	Logger := slog.With("operation", "start_bot", "bot_id", Event.BotID)
	Logger.DebugContext(Context, "Публикация BotStartedEvent", "bot_status", Event.BotStatus)
	if PublishError := StartBotUC.EventPublisherPort.PublishEvent(Context, Event); PublishError != nil {
		Logger.ErrorContext(Context, "Не удалось опубликовать BotStartedEvent", "error", PublishError)
		return PublishError
	}

	return nil
}
