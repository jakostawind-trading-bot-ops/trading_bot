package commands

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"

	telemetry "github.com/jakostawind-trading-bot-ops/trading_bot/internal"
	"go.opentelemetry.io/otel"
)

// ControlCommandDispatcher выбирает обработчик из таблицы, собранной до запуска подписчика.
type ControlCommandDispatcher struct {
	Handlers map[reflect.Type]ControlCommandHandlerI // Маршруты по прикладному типу команды.
	Logger   *slog.Logger                            // Логгер с именем компонента из DI.
}

// Dispatch передаёт Command выбранному обработчику в дочернем спане Context.
// Ошибка маршрутизации или выполнения возвращается подписчику для решения об Ack.
func (Dispatcher ControlCommandDispatcher) Dispatch(Context context.Context, Command BaseControlCommandI) (DispatchError error) {
	Context, Span := otel.Tracer("trading_bot/internal/commands").Start(Context, "ControlCommandDispatcher.Dispatch")
	defer telemetry.FinishSpan(Span, &DispatchError)
	CommandType := reflect.TypeOf(Command)
	Logger := Dispatcher.Logger.With("operation", "dispatch_control_command", "command_type", fmt.Sprintf("%T", Command))
	Logger.DebugContext(Context, "Поиск обработчика команды")
	Handler, Found := Dispatcher.Handlers[CommandType]
	if !Found {
		return fmt.Errorf("обработчик команды %T не зарегистрирован", Command)
	}
	Logger.DebugContext(Context, "Передача команды обработчику")
	return Handler.HandleControlCommand(Context, Command)
}
