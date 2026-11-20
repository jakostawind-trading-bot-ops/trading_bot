package di_states

import "github.com/jakostawind-trading-bot-ops/trading_bot/internal/states"

// GeneralStateProvider создаёт состояние бота с новым идентификатором.
func GeneralStateProvider() (*states.GeneralState, error) {
	botID, err := states.GenerateBotID()
	if err != nil {
		return nil, err
	}
	return &states.GeneralState{BotID: botID}, nil
}
