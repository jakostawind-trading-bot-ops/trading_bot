package transport_nats_publisher_routes_v1

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/events"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/ports"
	transport_nats_contracts_bot_control_v1_common "github.com/jakostawind-trading-bot-ops/trading_bot/internal/transport/nats/contracts/bot/control/v1/common"
	transport_nats_contracts_bot_control_v1_lifecycle "github.com/jakostawind-trading-bot-ops/trading_bot/internal/transport/nats/contracts/bot/control/v1/lifecycle"
)

func BotStartedEventV1PublisherRoute() ports.EventPublisherRoute {
	return ports.EventPublisherRoute{
		EventType:   reflect.TypeFor[*events.BotStartedEvent](),
		MessageType: "v1.lifecycle.bot_started_event",
		Encode: func(ctx context.Context, Event events.BaseBotEventI) ([]byte, error) {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			BotEvent, ok := Event.(*events.BotStartedEvent)
			if !ok || BotEvent == nil {
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
			if err := Contract.Validate(); err != nil {
				return nil, fmt.Errorf("валидация BotStartedEvent V1: %w", err)
			}
			Data, err := json.Marshal(Contract)
			if err != nil {
				return nil, fmt.Errorf("кодирование BotStartedEvent V1: %w", err)
			}
			return Data, nil
		},
	}
}
