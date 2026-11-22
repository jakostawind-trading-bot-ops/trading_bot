package transport_nats

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/commands"
	tracing "github.com/jakostawind-trading-bot-ops/trading_bot/internal/observability/tracing"
	"go.opentelemetry.io/otel"
)

// ControlCommandDecoder хранит версионные маршруты входящих команд.
type ControlCommandDecoder struct {
	Routes map[string]func(context.Context, []byte) (commands.BaseControlCommandI, error) // Декодеры по суффиксу subject.
	Logger *slog.Logger                                                                   // Логгер декодера из DI.
}

// Decode выбирает контракт по MessageType и передаёт Data в версионный декодер.
// Context связывает проверку JSON с обработкой сообщения. Содержимое Data в телеметрию не попадает.
func (Decoder ControlCommandDecoder) Decode(Context context.Context, MessageType string, Data []byte) (Command commands.BaseControlCommandI, DecodeError error) {
	Context, Span := otel.Tracer("trading_bot/internal/transport/nats").Start(Context, "ControlCommandDecoder.Decode")
	defer tracing.FinishSpan(Span, &DecodeError)
	Logger := Decoder.Logger.With("operation", "decode_control_command", "message_type", MessageType)
	Logger.DebugContext(Context, "Поиск маршрута декодирования", "payload_size", len(Data))
	DecodeCommand, Found := Decoder.Routes[MessageType]
	if !Found {
		return nil, fmt.Errorf("неизвестный тип команды %q", MessageType)
	}
	return DecodeCommand(Context, Data)
}
