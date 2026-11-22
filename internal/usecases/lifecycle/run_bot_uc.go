package usecases_lifecycle

import (
	"context"
	"log/slog"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/events"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/lifecycle"
	tracing "github.com/jakostawind-trading-bot-ops/trading_bot/internal/observability/tracing"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/ports"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/states"
	"go.opentelemetry.io/otel"
)

// RunBotUC обрабатывает последовательные команды одного экземпляра бота.
type RunBotUC struct {
	Logger             *slog.Logger             // Логгер сценария запуска.
	GeneralState       *states.GeneralState     // Состояние бота в памяти, включая BotID.
	Lifecycle          lifecycle.Lifecycle      // Управление статусом этого экземпляра.
	EventPublisherPort ports.EventPublisherPort // Издатель, возвращающий результат публикации.
	LastSequence       uint64                   // Номер последней команды RunBot в stream.
	PendingEvent       events.BaseBotEventI     // Результат до подтверждения публикации.
}

// Execute выполняет RunBot и публикует результат с Context. Отказ в запуске
// считается обработанным после публикации BotRunnedErrorEvent. При ошибке издателя
// подписчик повторит доставку той же команды. Он обрабатывает один запрос за раз.
func (RunBotUC *RunBotUC) Execute(Context context.Context) (ExecutionError error) {
	Context, Span := otel.Tracer("trading_bot/internal/usecases/lifecycle").Start(Context, "RunBotUC.Execute")
	defer tracing.FinishSpan(Span, &ExecutionError)

	Sequence := ports.ControlCommandSequence(Context)
	if Sequence == 0 {
		return RunBotUC.EventPublisherPort.PublishEvent(Context, RunBotUC.run(Context))
	}
	if Sequence > RunBotUC.LastSequence {
		RunBotUC.PendingEvent = RunBotUC.run(Context)
		RunBotUC.LastSequence = Sequence
	}
	if Sequence == RunBotUC.LastSequence && RunBotUC.PendingEvent != nil {
		ExecutionError = RunBotUC.EventPublisherPort.PublishEvent(Context, RunBotUC.PendingEvent)
		if ExecutionError == nil {
			RunBotUC.PendingEvent = nil
		}
	}
	return ExecutionError
}

// run переводит Started в Running и формирует событие по результату перехода.
func (RunBotUC *RunBotUC) run(Context context.Context) events.BaseBotEventI {
	BaseEvent := events.BaseBotEvent{BotID: RunBotUC.GeneralState.BotID}
	PreviousStatus := RunBotUC.GeneralState.GetBotStatus()
	if StatusError := RunBotUC.Lifecycle.ChangeStatus(Context, lifecycle.RUNNING, lifecycle.STARTED); StatusError != nil {
		RunBotUC.Logger.WarnContext(Context, "Запуск отклонён", "operation", "run_bot", "bot_status", PreviousStatus, "error", StatusError)
		return &events.BotRunnedErrorEvent{BaseBotEvent: BaseEvent, BotStatus: PreviousStatus, Error: StatusError.Error()}
	}
	return &events.BotRunnedEvent{BaseBotEvent: BaseEvent, PreviousBotStatus: PreviousStatus, CurrentBotStatus: lifecycle.RUNNING}
}
