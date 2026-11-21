package handlers

import (
	"context"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/commands"
	usecases "github.com/jakostawind-trading-bot-ops/trading_bot/internal/usecases/lifecycle"
)

// StartBotCommandHandler выполняет сценарий запуска, полученный из DI.
type StartBotCommandHandler struct {
	StartBot *usecases.StartBotUC // Общий сценарий публикации события запуска.
}

// HandleControlCommand передаёт Context сценарию и сохраняет его ошибку для JetStream.
// Маршрут привязан к *commands.StartBotCommand при сборке диспетчера.
func (Handler StartBotCommandHandler) HandleControlCommand(Context context.Context, _ commands.BaseControlCommandI) error {
	return Handler.StartBot.Execute(Context)
}
