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

// BotRunnedEventV1PublisherRoute связывает переход в Running с JSON-контрактом V1.
func BotRunnedEventV1PublisherRoute() ports.EventPublisherRoute {
	return ports.EventPublisherRoute{
		EventType:   reflect.TypeFor[*events.BotRunnedEvent](),
		MessageType: "v1.lifecycle.bot_runned_event",
		Encode: func(Context context.Context, Event events.BaseBotEventI) (Data []byte, EncodeError error) {
			Context, Span := otel.Tracer("trading_bot/internal/transport/nats/publisher_routes/v1").Start(Context, "BotRunnedEventV1.Encode")
			defer tracing.FinishSpan(Span, &EncodeError)
			if ContextError := Context.Err(); ContextError != nil {
				return nil, ContextError
			}
			BotEvent, Matches := Event.(*events.BotRunnedEvent)
			if !Matches || BotEvent == nil {
				return nil, fmt.Errorf("ожидался ненулевой *events.BotRunnedEvent")
			}

			Contract := &lifecyclev1.BotRunnedEventContractV1{
				BaseBotEventContractV1: commonv1.BaseBotEventContractV1{
					BotID:     BotEvent.BotID,
					Payload:   BotEvent.Payload,
					Timestamp: BotEvent.Timestamp,
				},
				PreviousBotStatus: string(BotEvent.PreviousBotStatus),
				CurrentBotStatus:  string(BotEvent.CurrentBotStatus),
			}
			if ValidationError := Contract.Validate(); ValidationError != nil {
				return nil, fmt.Errorf("валидация BotRunnedEvent V1: %w", ValidationError)
			}
			Data, EncodeError = json.Marshal(Contract)
			if EncodeError != nil {
				return nil, logging.SafeError("ошибка кодирования BotRunnedEvent V1", EncodeError)
			}
			return Data, nil
		},
	}
}

// BotRunnedErrorEventV1PublisherRoute связывает отказ в запуске с JSON-контрактом V1.
func BotRunnedErrorEventV1PublisherRoute() ports.EventPublisherRoute {
	return ports.EventPublisherRoute{
		EventType:   reflect.TypeFor[*events.BotRunnedErrorEvent](),
		MessageType: "v1.lifecycle.bot_runned_error_event",
		Encode: func(Context context.Context, Event events.BaseBotEventI) (Data []byte, EncodeError error) {
			Context, Span := otel.Tracer("trading_bot/internal/transport/nats/publisher_routes/v1").Start(Context, "BotRunnedErrorEventV1.Encode")
			defer tracing.FinishSpan(Span, &EncodeError)
			if ContextError := Context.Err(); ContextError != nil {
				return nil, ContextError
			}
			BotEvent, Matches := Event.(*events.BotRunnedErrorEvent)
			if !Matches || BotEvent == nil {
				return nil, fmt.Errorf("ожидался ненулевой *events.BotRunnedErrorEvent")
			}

			Contract := &lifecyclev1.BotRunnedErrorEventContractV1{
				BaseBotEventContractV1: commonv1.BaseBotEventContractV1{
					BotID:     BotEvent.BotID,
					Payload:   BotEvent.Payload,
					Timestamp: BotEvent.Timestamp,
				},
				BotStatus: string(BotEvent.BotStatus),
				Error:     BotEvent.Error,
			}
			if ValidationError := Contract.Validate(); ValidationError != nil {
				return nil, fmt.Errorf("валидация BotRunnedErrorEvent V1: %w", ValidationError)
			}
			Data, EncodeError = json.Marshal(Contract)
			if EncodeError != nil {
				return nil, logging.SafeError("ошибка кодирования BotRunnedErrorEvent V1", EncodeError)
			}
			return Data, nil
		},
	}
}
