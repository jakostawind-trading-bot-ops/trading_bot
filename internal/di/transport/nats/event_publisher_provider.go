package di_transport_nats

import (
	"context"
	"log/slog"

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
	Logger *slog.Logger,
) (ports.EventPublisherPort, error) {
	// Фабрика вызывается при публикации, когда lifecycle NatsClient уже запущен.
	GetJetStream := func(Context context.Context) (nats.JetStreamContext, error) {
		Connection, WaitError := NatsClient.WaitConnection(Context)
		if WaitError != nil {
			return nil, WaitError
		}
		return Connection.JetStream()
	}
	return transport.NewNatsEventPublisher(
		Routes, GetJetStream, NATS_BOT_EVENT_STREAM,
		Logger.With("component", "nats_event_publisher", "stream", NATS_BOT_EVENT_STREAM),
	)
}
