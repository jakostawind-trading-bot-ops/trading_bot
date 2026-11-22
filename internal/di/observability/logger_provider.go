package di_observability

import (
	"log/slog"
	"os"

	"github.com/jakostawind-trading-bot-ops/trading_bot/internal"
	logging "github.com/jakostawind-trading-bot-ops/trading_bot/internal/observability/logging"
	"go.uber.org/fx"
)

// ServiceName задаёт общее имя сервиса для логов и спанов бота.
const ServiceName = "trading-bot"

// LoggerProvider создаёт общий slog-логгер по Settings. Lifecycle выгружает
// оставшиеся записи в stdout после остановки потребителей.
func LoggerProvider(Lifecycle fx.Lifecycle, Settings *internal.BootstrapSettings) (*slog.Logger, error) {
	Configuration := logging.Logger{
		ServiceName: ServiceName,
		Stage:       Settings.Stage,
		LoggerLevel: Settings.LoggerLevel,
	}
	Logger, Provider, SetupError := Configuration.CreateOTelLogger(os.Stdout)
	if SetupError != nil {
		return nil, SetupError
	}
	Lifecycle.Append(fx.Hook{OnStop: Provider.Shutdown})
	return Logger, nil
}
