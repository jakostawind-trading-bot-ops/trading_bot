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
	commonv1 "github.com/jakostawind-trading-bot-ops/trading_bot/internal/transport/nats/contracts/bot/control/v1/common"
	lifecyclev1 "github.com/jakostawind-trading-bot-ops/trading_bot/internal/transport/nats/contracts/bot/control/v1/lifecycle"
	"go.opentelemetry.io/otel"
)

// BotStatusChangedEventV1PublisherRoute связывает смену статуса бота с JSON-контрактом V1.
func BotStatusChangedEventV1PublisherRoute() ports.EventPublisherRoute {
	return ports.EventPublisherRoute{
		EventType:   reflect.TypeFor[*events.BotStatusChangedEvent](),
		MessageType: "v1.lifecycle.bot_status_changed_event",
		Encode: func(Context context.Context, Event events.BaseBotEventI) (Data []byte, EncodeError error) {
			Context, Span := otel.Tracer("trading_bot/internal/transport/nats/publisher_routes/v1").Start(Context, "BotStatusChangedEventV1.Encode")
			defer tracing.FinishSpan(Span, &EncodeError)
			if ContextError := Context.Err(); ContextError != nil {
				return nil, ContextError
			}
			BotEvent, Matches := Event.(*events.BotStatusChangedEvent)
			if !Matches || BotEvent == nil {
				return nil, fmt.Errorf("ожидался ненулевой *events.BotStatusChangedEvent")
			}

			Contract := &lifecyclev1.BotStatusChangedEventContractV1{
				BaseBotEventContractV1: commonv1.BaseBotEventContractV1{
					BotID:     BotEvent.BotID,
					Payload:   BotEvent.Payload,
					Timestamp: BotEvent.Timestamp,
				},
				PreviousBotStatus: string(BotEvent.PreviousBotStatus),
				CurrentBotStatus:  string(BotEvent.CurrentBotStatus),
			}
			if ValidationError := Contract.Validate(); ValidationError != nil {
				return nil, fmt.Errorf("валидация BotStatusChangedEvent V1: %w", ValidationError)
			}
			Data, EncodeError = json.Marshal(Contract)
			if EncodeError != nil {
				return nil, logging.SafeError("ошибка кодирования BotStatusChangedEvent V1", EncodeError)
			}
			return Data, nil
		},
	}
}

// BotStatusChangedErrorEventV1PublisherRoute связывает ошибку смены статуса с JSON-контрактом V1.
func BotStatusChangedErrorEventV1PublisherRoute() ports.EventPublisherRoute {
	return ports.EventPublisherRoute{
		EventType:   reflect.TypeFor[*events.BotStatusChangedErrorEvent](),
		MessageType: "v1.lifecycle.bot_status_changed_error_event",
		Encode: func(Context context.Context, Event events.BaseBotEventI) (Data []byte, EncodeError error) {
			Context, Span := otel.Tracer("trading_bot/internal/transport/nats/publisher_routes/v1").Start(Context, "BotStatusChangedErrorEventV1.Encode")
			defer tracing.FinishSpan(Span, &EncodeError)
			if ContextError := Context.Err(); ContextError != nil {
				return nil, ContextError
			}
			BotEvent, Matches := Event.(*events.BotStatusChangedErrorEvent)
			if !Matches || BotEvent == nil {
				return nil, fmt.Errorf("ожидался ненулевой *events.BotStatusChangedErrorEvent")
			}

			Contract := &lifecyclev1.BotStatusChangedErrorEventContractV1{
				BaseBotEventContractV1: commonv1.BaseBotEventContractV1{
					BotID:     BotEvent.BotID,
					Payload:   BotEvent.Payload,
					Timestamp: BotEvent.Timestamp,
				},
				BotStatus:          string(BotEvent.BotStatus),
				RequestedBotStatus: string(BotEvent.RequestedBotStatus),
				ErrorBotStatus:     string(BotEvent.ErrorBotStatus),
			}
			if ValidationError := Contract.Validate(); ValidationError != nil {
				return nil, fmt.Errorf("валидация BotStatusChangedErrorEvent V1: %w", ValidationError)
			}
			Data, EncodeError = json.Marshal(Contract)
			if EncodeError != nil {
				return nil, logging.SafeError("ошибка кодирования BotStatusChangedErrorEvent V1", EncodeError)
			}
			return Data, nil
		},
	}
}
