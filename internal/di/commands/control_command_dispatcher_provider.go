package di_commands

import (
	"log/slog"
	"reflect"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/commands"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/commands/handlers"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/ports"
	usecases "github.com/jakostawind-trading-bot-ops/trading_bot/internal/usecases/lifecycle"
)

// ControlCommandDispatcherProvider регистрирует прикладные маршруты до запуска подписчика.
func ControlCommandDispatcherProvider(StartBot *usecases.StartBotUC, Logger *slog.Logger) ports.ControlCommandDispatcherPort {
	return commands.ControlCommandDispatcher{
		Logger: Logger.With("component", "control_command_dispatcher"),
		Handlers: map[reflect.Type]commands.ControlCommandHandlerI{
			reflect.TypeFor[*commands.StartBotCommand](): handlers.StartBotCommandHandler{
				StartBot: StartBot,
				Logger:   Logger.With("component", "start_bot_command_handler"),
			},
		},
	}
}
