package usecases_lifecycle

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/events"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/lifecycle"
	tracing "github.com/jakostawind-trading-bot-ops/trading_bot/internal/observability/tracing"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/ports"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/states"
	"go.opentelemetry.io/otel"
)

type ChangeBotStatusUC struct {
	Logger             *slog.Logger
	GeneralState       *states.GeneralState     // Состояние бота в памяти, включая BotID.
	Lifecycle          lifecycle.Lifecycle      // Управление статусом этого экземпляра.
	EventPublisherPort ports.EventPublisherPort // Издатель, возвращающий результат публикации.
}

// Execute переводит бота из Started в Running по запросу RequestedBotStatus.
// Передаёт Context в ChangeStatus и в публикацию события об успехе или отказе.
// Возвращает ошибку перехода, публикации или обе ошибки.
func (ChangeBotStatusUC *ChangeBotStatusUC) Execute(
	Context context.Context,
	RequestedBotStatus string,
) (ExecutionError error) {
	Context, Span := otel.Tracer("trading_bot/internal/usecases/lifecycle").Start(Context, "ChangeBotStatusUC.Execute")
	defer tracing.FinishSpan(Span, &ExecutionError)

	RequestedStatus := lifecycle.LifecycleStatus(RequestedBotStatus)
	PreviousBotStatus := ChangeBotStatusUC.GeneralState.GetBotStatus()
	BaseEvent := events.BaseBotEvent{BotID: ChangeBotStatusUC.GeneralState.BotID}
	var StatusError error
	if RequestedStatus == lifecycle.RUNNING {
		StatusError = ChangeBotStatusUC.Lifecycle.ChangeStatus(Context, RequestedStatus, lifecycle.STARTED)
	} else {
		StatusError = fmt.Errorf("неподдерживаемый целевой статус %q", RequestedStatus)
	}
	if StatusError == nil {
		Event := &events.BotStatusChangedEvent{
			BaseBotEvent:      BaseEvent,
			PreviousBotStatus: PreviousBotStatus,
			CurrentBotStatus:  RequestedStatus,
		}
		return ChangeBotStatusUC.EventPublisherPort.PublishEvent(Context, Event)
	}

	Event := &events.BotStatusChangedErrorEvent{
		BaseBotEvent:       BaseEvent,
		BotStatus:          ChangeBotStatusUC.GeneralState.GetBotStatus(),
		RequestedBotStatus: RequestedStatus,
	}
	ChangeBotStatusUC.Logger.WarnContext(Context, "Смена статуса отклонена", "operation", "change_bot_status_uc",
		"bot_status", Event.BotStatus, "requested_bot_status", RequestedStatus, "error", StatusError)
	PublishError := ChangeBotStatusUC.EventPublisherPort.PublishEvent(Context, Event)
	if PublishError != nil {
		return errors.Join(StatusError, PublishError)
	}
	return StatusError
}
