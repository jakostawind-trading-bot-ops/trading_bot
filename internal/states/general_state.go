package states

import (
	"crypto/rand"
	"math/big"
)

// GeneralState хранит общее состояние запущенного экземпляра бота.
type GeneralState struct {
	BotID string
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