package handlers

import (
	"context"
	"log/slog"

	telemetry "github.com/jakostawind-trading-bot-ops/trading_bot/internal"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/commands"
	usecases "github.com/jakostawind-trading-bot-ops/trading_bot/internal/usecases/lifecycle"
	"go.opentelemetry.io/otel"
)

// StartBotCommandHandler выполняет сценарий запуска, полученный из DI.
type StartBotCommandHandler struct {
	Logger   *slog.Logger         // Логгер обработчика из DI.
	StartBot *usecases.StartBotUC // Общий сценарий публикации события запуска.
}

// HandleControlCommand передаёт Context сценарию и сохраняет его ошибку для JetStream.
// Маршрут привязан к *commands.StartBotCommand при сборке диспетчера.
func (Handler StartBotCommandHandler) HandleControlCommand(Context context.Context, _ commands.BaseControlCommandI) (HandlerError error) {
	Context, Span := otel.Tracer("trading_bot/internal/commands/handlers").Start(Context, "StartBotCommandHandler.HandleControlCommand")
	defer telemetry.FinishSpan(Span, &HandlerError)
	Handler.Logger.DebugContext(Context, "Вызов сценария запуска бота", "operation", "handle_start_bot_command")
	return Handler.StartBot.Execute(Context)
}
