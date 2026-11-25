package usecases_trading

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/JaKoSTaWiND/go-polymarket-sdk/markets"
	tracing "github.com/jakostawind-trading-bot-ops/trading_bot/internal/observability/tracing"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/ports"
	"go.opentelemetry.io/otel"
)

type SubscribeMarketById struct {
	Logger             *slog.Logger
	HttpClient         *http.Client
	EventPublisherPort ports.EventPublisherPort
}

func (SubscribeMarketById *SubscribeMarketById) Execute(
	ctx context.Context,
	MarketID string,
) (ExecutionError error) {
	ctx, Span := otel.Tracer("trading_bot/internal/usecases/trading").Start(ctx, "RunBotSubscribeMarketByIdUC.Execute")
	defer tracing.FinishSpan(Span, &ExecutionError)

	// 1 получить данный market по id
	Market, err := markets.GetMarketByID(ctx, SubscribeMarketById.HttpClient, MarketID)
	if err != nil {
		//паблишь ивента что не удалось получить маркет в сдк
	}

	// 2 GeneralActor.MarketSubscribe

	// опубликовать ивент
	return nil
}
