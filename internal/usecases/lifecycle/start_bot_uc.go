package usecases_lifecycle

import (
	"context"
	"log/slog"

	telemetry "github.com/jakostawind-trading-bot-ops/trading_bot/internal"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/events"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/lifecycle"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/ports"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/states"
	"go.opentelemetry.io/otel"
)

// StartBotUC строит событие запуска по текущему состоянию и публикует его в JetStream.
type StartBotUC struct {
	Logger             *slog.Logger             // Логгер сценария с идентификатором бота.
	GeneralState       *states.GeneralState     // Состояние бота в памяти, включая BotID.
	Lifecycle          lifecycle.Lifecycle      // Управление статусом этого экземпляра.
	EventPublisherPort ports.EventPublisherPort // Издатель, возвращающий результат публикации.
}

// Execute публикует событие запуска. Возвращает ошибку публикации подписчику команд.
func (StartBotUC *StartBotUC) Execute(Context context.Context) (PublishError error) {
	Context, Span := otel.Tracer("trading_bot/internal/usecases/lifecycle").Start(Context, "StartBotUC.Execute")
	defer telemetry.FinishSpan(Span, &PublishError)
	Logger := StartBotUC.Logger.With("operation", "start_bot")
	Logger.DebugContext(Context, "Чтение текущего статуса для события запуска")
	Event := &events.BotStartedEvent{
		BaseBotEvent: events.BaseBotEvent{
			BotID: StartBotUC.GeneralState.BotID,
		},
		BotStatus: StartBotUC.GeneralState.GetBotStatus(),
	}
	Logger.DebugContext(Context, "Публикация BotStartedEvent", "bot_status", Event.BotStatus)
	return StartBotUC.EventPublisherPort.PublishEvent(Context, Event)
}
