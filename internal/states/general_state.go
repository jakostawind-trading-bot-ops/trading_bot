package states

import (
	"crypto/rand"
	"math/big"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/lifecycle"
)

// GeneralState хранит общее состояние запущенного экземпляра бота.
type GeneralState struct {
	BotID     string
	BotStatus lifecycle.LifecycleStatus
}

// GenerateBotID возвращает случайный идентификатор из 16 букв A-Z.
func GenerateBotID() (string, error) {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var id [16]byte
	limit := big.NewInt(int64(len(alphabet)))

	for i := range id {
		index, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", err
		}
		id[i] = alphabet[index.Int64()]
	}
	return string(id[:]), nil
}

func (GeneralState *GeneralState) GetBotStatus() lifecycle.LifecycleStatus {
	return GeneralState.BotStatus
}

func (GeneralState *GeneralState) SetBotStatus(Status lifecycle.LifecycleStatus) {
	GeneralState.BotStatus = Status
}
