package transport_nats_publisher_routes_v1

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/events"
	logging "github.com/jakostawind-trading-bot-ops/trading_bot/internal/observability/logging"
	tracing "github.com/jakostawind-trading-bot-ops/trading_bot/internal/observability/tracing"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/ports"
	transport_nats_contracts_bot_control_v1_common "github.com/jakostawind-trading-bot-ops/trading_bot/internal/transport/nats/contracts/bot/control/v1/common"
	transport_nats_contracts_bot_control_v1_lifecycle "github.com/jakostawind-trading-bot-ops/trading_bot/internal/transport/nats/contracts/bot/control/v1/lifecycle"
	"go.opentelemetry.io/otel"
)

// BotStartedEventV1PublisherRoute связывает событие запуска с JSON-контрактом V1.
func BotStartedEventV1PublisherRoute() ports.EventPublisherRoute {
	return ports.EventPublisherRoute{
		EventType:   reflect.TypeFor[*events.BotStartedEvent](),
		MessageType: "v1.lifecycle.bot_started_event",
		Encode: func(Context context.Context, Event events.BaseBotEventI) (Data []byte, EncodeError error) {
			Context, Span := otel.Tracer("trading_bot/internal/transport/nats/publisher_routes/v1").Start(Context, "BotStartedEventV1.Encode")
			defer tracing.FinishSpan(Span, &EncodeError)
			if ContextError := Context.Err(); ContextError != nil {
				return nil, ContextError
			}
			BotEvent, Matches := Event.(*events.BotStartedEvent)
			if !Matches || BotEvent == nil {
				return nil, fmt.Errorf("ожидался ненулевой *events.BotStartedEvent")
			}

			// Явно переносим данные во внешний контракт.
			Contract := &transport_nats_contracts_bot_control_v1_lifecycle.BotStartedEventContractV1{
				BaseBotEventContractV1: transport_nats_contracts_bot_control_v1_common.BaseBotEventContractV1{
					BotID:     BotEvent.BotID,
					Payload:   BotEvent.Payload,
					Timestamp: BotEvent.Timestamp,
				},
			}
			if ValidationError := Contract.Validate(); ValidationError != nil {
				return nil, fmt.Errorf("валидация BotStartedEvent V1: %w", ValidationError)
			}
			Data, EncodeError = json.Marshal(Contract)
			if EncodeError != nil {
				return nil, logging.SafeError("ошибка кодирования BotStartedEvent V1", EncodeError)
			}
			return Data, nil
		},
	}
}
