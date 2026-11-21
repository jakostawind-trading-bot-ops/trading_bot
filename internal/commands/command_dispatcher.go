package commands

import (
	"context"
	"fmt"
	"reflect"
)

// ControlCommandDispatcher хранит зарегистрированные в DI обработчики по Go-типу команды.
type ControlCommandDispatcher map[reflect.Type]ControlCommandHandlerI

// Dispatch находит обработчик Command и возвращает ошибку выполнения вызывающему подписчику.
//
//	Command -> Go-тип -> обработчик -> результат
func (Dispatcher ControlCommandDispatcher) Dispatch(Context context.Context, Command BaseControlCommandI) error {
	Handler, Found := Dispatcher[reflect.TypeOf(Command)]
	if !Found {
		return fmt.Errorf("обработчик команды %T не зарегистрирован", Command)
	}
	return Handler.HandleControlCommand(Context, Command)
}
