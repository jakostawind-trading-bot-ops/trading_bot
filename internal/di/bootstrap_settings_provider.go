package di

import (
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal"
	"go.uber.org/fx"
)

// BootstrapSettingsProvider передаёт настройки запуска в конструкторы Fx.
func BootstrapSettingsProvider(BootstrapSettings *internal.BootstrapSettings) fx.Option {
	return fx.Supply(BootstrapSettings)
}
