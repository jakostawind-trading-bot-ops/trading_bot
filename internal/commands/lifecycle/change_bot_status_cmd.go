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

type ChangeBotStatusCommand struct {
	commands.BaseControlCommand
	RequestBotStatus string
}

type ChangeBotStatusCommandHandler struct {
	Logger            *slog.Logger
	ChangeBotStatusUC *usecases_lifecycle.ChangeBotStatusUC
}

// HandleControlCommand передаёт запрошенный статус в юзкейс с контекстом Context.
// Возвращает ошибку для пустой команды, неверного типа ControlCommand или сбоя Execute.
func (Handler ChangeBotStatusCommandHandler) HandleControlCommand(
	Context context.Context,
	ControlCommand commands.BaseControlCommandI,
) (HandlerError error) {
	Context, Span := otel.Tracer("trading_bot/internal/commands/lifecycle").Start(Context, "ChangeBotStatusCommandHandler.HandleControlCommand")
	defer observability_tracing.FinishSpan(Span, &HandlerError)

	Command, Valid := ControlCommand.(*ChangeBotStatusCommand)
	if !Valid || Command == nil {
		return fmt.Errorf("ожидалась непустая команда *ChangeBotStatusCommand, получено %T", ControlCommand)
	}
	Handler.Logger.DebugContext(Context, "Вызов ChangeBotStatusUC", "operation", "handle_change_bot_status_command", "requested_bot_status", Command.RequestBotStatus)
	return Handler.ChangeBotStatusUC.Execute(Context, Command.RequestBotStatus)
}
