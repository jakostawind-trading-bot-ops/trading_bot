package di_commands

import (
	"reflect"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/commands"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/commands/handlers"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/ports"
	usecases "github.com/jakostawind-trading-bot-ops/trading_bot/internal/usecases/lifecycle"
)

// ControlCommandDispatcherProvider регистрирует прикладные маршруты до запуска подписчика.
func ControlCommandDispatcherProvider(StartBot *usecases.StartBotUC) ports.ControlCommandDispatcherPort {
	return commands.ControlCommandDispatcher{
		reflect.TypeFor[*commands.StartBotCommand](): handlers.StartBotCommandHandler{StartBot: StartBot},
	}
}
