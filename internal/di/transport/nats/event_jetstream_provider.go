package di_transport_nats

import (
	"context"
	"time"

	tracing "github.com/jakostawind-trading-bot-ops/trading_bot/internal/observability/tracing"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
)

// BotEventJetStreamFactory получает JetStream после подключения NATS.
// Context отменяет ожидание соединения. Результат — JetStreamContext либо ошибка
// ожидания/создания. Запросы JetStream ограничены таймаутом в три секунды.
type BotEventJetStreamFactory func(Context context.Context) (nats.JetStreamContext, error)

// NatsBotEventJetStreamProvider возвращает функцию отложенного получения
// JetStream. NatsClient — владелец соединения, создаваемый общим провайдером.
// Конструктор не ждёт подключения; возвращённую функцию вызывают после OnStart.
//
// Схема:
//
//	Вызов фабрики -> [WaitConnection(ctx)] -> [Conn.JetStream] -> результат
func NatsBotEventJetStreamProvider(
	NatsClient *NatsClient,
) BotEventJetStreamFactory {
	return func(Context context.Context) (JetStream nats.JetStreamContext, ConnectionError error) {
		Context, Span := otel.Tracer("trading_bot/internal/di/transport/nats").Start(Context, "NatsClient.OpenEventJetStream")
		defer tracing.FinishSpan(Span, &ConnectionError)
		Connection, ConnectionError := NatsClient.WaitConnection(Context)
		if ConnectionError != nil {
			return nil, ConnectionError
		}
		// Ограничиваем Bind-запрос, сохраняя явное закрытие через Drain.
		return Connection.JetStream(nats.MaxWait(3 * time.Second))
	}
}
