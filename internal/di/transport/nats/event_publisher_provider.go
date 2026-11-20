package di_transport_nats

import (
	"context"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/ports"
	transport "github.com/jakostawind-trading-bot-ops/trading_bot/internal/transport/nats"
	"github.com/nats-io/nats.go"
)

const NATS_BOT_EVENT_STREAM = "BOT_TO_CONTROL_EVENTS" // Существующий stream событий.

// NatsEventPublisherProvider связывает Routes из реестра и общий NatsClient.
// Возвращает порт publisher либо ошибку регистрации маршрутов.
// NatsClient запускается через Fx; провайдер подключения не ожидает.
//
// Схема:
//
//	Fx -> Routes + NatsClient -> publisher
//	PublishEvent -> фабрика -> WaitConnection(ctx) -> JetStream
func NatsEventPublisherProvider(
	Routes []ports.EventPublisherRoute,
	NatsClient *NatsClient,
) (ports.EventPublisherPort, error) {
	// Фабрика вызывается при публикации, когда lifecycle NatsClient уже запущен.
	GetJetStream := func(ctx context.Context) (nats.JetStreamContext, error) {
		Conn, err := NatsClient.WaitConnection(ctx)
		if err != nil {
			return nil, err
		}
		return Conn.JetStream()
	}
	Publisher, err := transport.NewNatsEventPublisher(
		Routes, GetJetStream, NATS_BOT_EVENT_STREAM,
	)
	if err != nil {
		return nil, err
	}
	return Publisher, nil
}
