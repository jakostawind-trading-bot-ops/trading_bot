package di_commands

import (
	"log/slog"
	"reflect"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/commands"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/commands/handlers"
	commands_lifecycle "github.com/jakostawind-trading-bot-ops/trading_bot/internal/commands/lifecycle"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/ports"
	usecases "github.com/jakostawind-trading-bot-ops/trading_bot/internal/usecases/lifecycle"
)

// ControlCommandDispatcherProvider регистрирует прикладные маршруты до запуска подписчика.
func ControlCommandDispatcherProvider(StartBot *usecases.StartBotUC, RunBot *usecases.RunBotUC, Logger *slog.Logger) ports.ControlCommandDispatcherPort {
	return commands.ControlCommandDispatcher{
		Logger: Logger.With("component", "control_command_dispatcher"),
		Handlers: map[reflect.Type]commands.ControlCommandHandlerI{
			reflect.TypeFor[*commands.StartBotCommand](): handlers.StartBotCommandHandler{
				StartBot: StartBot,
				Logger:   Logger.With("component", "start_bot_command_handler"),
			},
			reflect.TypeFor[*commands_lifecycle.RunBotCommand](): commands_lifecycle.RunBotCommandHandler{
				RunBotUC: RunBot,
				Logger:   Logger.With("component", "run_bot_command_handler"),
			},
		},
	}
}
