package transport_nats

import (
	"fmt"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/commands"
)

// ControlCommandDecoder хранит функции декодирования по версионному суффиксу subject.
type ControlCommandDecoder map[string]func([]byte) (commands.BaseControlCommandI, error)

// Decode выбирает контракт по MessageType и передаёт ему исходный JSON Data.
func (Decoder ControlCommandDecoder) Decode(MessageType string, Data []byte) (commands.BaseControlCommandI, error) {
	DecodeCommand, Found := Decoder[MessageType]
	if !Found {
		return nil, fmt.Errorf("неизвестный тип команды %q", MessageType)
	}
	return DecodeCommand(Data)
}
