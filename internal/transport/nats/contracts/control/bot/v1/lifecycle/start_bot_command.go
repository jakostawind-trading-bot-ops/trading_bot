package lifecycle

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	telemetry "github.com/jakostawind-trading-bot-ops/trading_bot/internal"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/commands"
	"go.opentelemetry.io/otel"
)

// StartBotCommandV1 описывает входной JSON команды запуска.
type StartBotCommandV1 struct {
	BotID     string         `json:"bot_id"`    // Адресат, дополнительно сверяется с подпиской.
	Timestamp string         `json:"timestamp"` // Время формирования команды.
	Payload   map[string]any `json:"payload"`   // Дополнительные параметры Control Plane.
}

// DecodeStartBotCommand проверяет обязательные поля Data и возвращает команду
// для диспетчера. В ошибку не включается содержимое сообщения.
//
//	JSON -> контракт V1 -> проверка полей -> StartBotCommand
func DecodeStartBotCommand(Context context.Context, Data []byte) (Command commands.BaseControlCommandI, ContractError error) {
	_, Span := otel.Tracer("trading_bot/internal/transport/nats/contracts/control/bot/v1/lifecycle").Start(Context, "StartBotCommandV1.Decode")
	defer telemetry.FinishSpan(Span, &ContractError)
	var Contract StartBotCommandV1
	if DecodeError := json.Unmarshal(Data, &Contract); DecodeError != nil {
		return nil, telemetry.SafeError("некорректный JSON команды запуска", DecodeError)
	}
	if strings.TrimSpace(Contract.BotID) == "" || strings.TrimSpace(Contract.Timestamp) == "" {
		return nil, fmt.Errorf("для команды запуска обязательны bot_id и timestamp")
	}
	return &commands.StartBotCommand{BaseControlCommand: commands.BaseControlCommand{
		BotID: Contract.BotID, Timestamp: Contract.Timestamp, Payload: Contract.Payload,
	}}, nil
}
