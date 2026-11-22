package di_observability

import (
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal"
	tracing "github.com/jakostawind-trading-bot-ops/trading_bot/internal/observability/tracing"
	"go.uber.org/fx"
)

// TracerProvider создаёт трейсинг для окружения Settings.Stage и регистрирует
// в Lifecycle запуск SDK и выгрузку спанов при остановке.
func TracerProvider(Lifecycle fx.Lifecycle, Settings *internal.BootstrapSettings) *tracing.Tracing {
	Tracing := &tracing.Tracing{ServiceName: ServiceName, Stage: Settings.Stage}
	Lifecycle.Append(fx.Hook{OnStart: Tracing.Start, OnStop: Tracing.Shutdown})
	return Tracing
}
