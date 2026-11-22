package commands_lifecycle

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/commands"
	observability_tracing "github.com/jakostawind-trading-bot-ops/trading_bot/internal/observability/tracing"
	usecases_lifecycle "github.com/jakostawind-trading-bot-ops/trading_bot/internal/usecases/lifecycle"
	"go.opentelemetry.io/otel"
)

type RunBotCommand struct {
	commands.BaseControlCommand
}

type RunBotCommandHandler struct {
	Logger   *slog.Logger
	RunBotUC *usecases_lifecycle.RunBotUC
}

// HandleControlCommand вызывает RunBotUC с Context для базовой команды запуска.
// Возвращает ошибку для пустой команды, неверного типа ControlCommand или сбоя Execute.
func (Handler RunBotCommandHandler) HandleControlCommand(
	Context context.Context,
	ControlCommand commands.BaseControlCommandI,
) (HandlerError error) {
	Context, Span := otel.Tracer("trading_bot/internal/commands/lifecycle").Start(Context, "RunBotCommandHandler.HandleControlCommand")
	defer observability_tracing.FinishSpan(Span, &HandlerError)

	Command, Valid := ControlCommand.(*RunBotCommand)
	if !Valid || Command == nil {
		return fmt.Errorf("ожидалась непустая команда *RunBotCommand, получено %T", ControlCommand)
	}
	Handler.Logger.DebugContext(Context, "Вызов RunBotUC", "operation", "run_bot_command")
	return Handler.RunBotUC.Execute(Context)
}
