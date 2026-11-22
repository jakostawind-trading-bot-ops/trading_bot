package transport_nats_publisher_routes

import (
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/ports"
	v1 "github.com/jakostawind-trading-bot-ops/trading_bot/internal/transport/nats/publisher_routes/v1"
)

// EventPublisherRoutesRegistry возвращает маршруты, передаваемые Fx в publisher.
func EventPublisherRoutesRegistry() []ports.EventPublisherRoute {
	return []ports.EventPublisherRoute{
		v1.BotStartedEventV1PublisherRoute(),
		v1.BotRunnedEventV1PublisherRoute(),
		v1.BotRunnedErrorEventV1PublisherRoute(),
	}
}
