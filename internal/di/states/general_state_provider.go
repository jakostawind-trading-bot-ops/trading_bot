package di_states

import (
	"log/slog"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/lifecycle"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/states"
)

// GeneralStateProvider создаёт состояние бота с новым идентификатором.
func GeneralStateProvider(Logger *slog.Logger) (*states.GeneralState, error) {
	BotID, GenerateError := states.GenerateBotID()
	if GenerateError != nil {
		Logger.Error("Не удалось создать идентификатор бота", "component", "general_state", "operation", "initialize_state", "error", GenerateError)
		return nil, GenerateError
	}
	Logger.Debug("Создано состояние экземпляра", "component", "general_state", "operation", "initialize_state", "bot_id", BotID)
	return &states.GeneralState{BotID: BotID, BotStatus: lifecycle.STARTED}, nil
}
