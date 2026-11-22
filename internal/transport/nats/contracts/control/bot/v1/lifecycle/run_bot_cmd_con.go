package lifecycle

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/commands"
	commands_lifecycle "github.com/jakostawind-trading-bot-ops/trading_bot/internal/commands/lifecycle"
	logging "github.com/jakostawind-trading-bot-ops/trading_bot/internal/observability/logging"
	tracing "github.com/jakostawind-trading-bot-ops/trading_bot/internal/observability/tracing"
	"go.opentelemetry.io/otel"
)

// RunBotCommandV1 содержит базовые поля команды перехода в Running.
type RunBotCommandV1 struct {
	BotID     string         `json:"bot_id"`    // Адресат, дополнительно сверяется с подпиской.
	Timestamp string         `json:"timestamp"` // Время формирования команды.
	Payload   map[string]any `json:"payload"`   // Дополнительные параметры Control Plane.
}

// DecodeRunBotCommand проверяет обязательные поля Data и возвращает команду
// для диспетчера. В ошибку не включается содержимое сообщения.
//
//	JSON -> контракт V1 -> проверка полей -> RunBotCommand
func DecodeRunBotCommand(Context context.Context, Data []byte) (Command commands.BaseControlCommandI, ContractError error) {
	_, Span := otel.Tracer("trading_bot/internal/transport/nats/contracts/control/bot/v1/lifecycle").Start(Context, "RunBotCommandV1.Decode")
	defer tracing.FinishSpan(Span, &ContractError)
	var Contract RunBotCommandV1
	if !json.Valid(Data) {
		return nil, fmt.Errorf("некорректный JSON команды RunBot")
	}
	Decoder := json.NewDecoder(bytes.NewReader(Data))
	Decoder.DisallowUnknownFields()
	if DecodeError := Decoder.Decode(&Contract); DecodeError != nil {
		return nil, logging.SafeError("некорректный JSON команды запуска", DecodeError)
	}
	if strings.TrimSpace(Contract.BotID) == "" || strings.TrimSpace(Contract.Timestamp) == "" {
		return nil, fmt.Errorf("для команды запуска обязательны bot_id и timestamp")
	}
	return &commands_lifecycle.RunBotCommand{BaseControlCommand: commands.BaseControlCommand{
		BotID:     Contract.BotID,
		Timestamp: Contract.Timestamp,
		Payload:   Contract.Payload,
	}}, nil
}
