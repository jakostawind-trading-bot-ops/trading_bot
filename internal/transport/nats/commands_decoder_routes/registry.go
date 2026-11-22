package transport_nats_commands_decoder_routes

import (
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/ports"
	v1 "github.com/jakostawind-trading-bot-ops/trading_bot/internal/transport/nats/commands_decoder_routes/v1"
)

// ControlCommandDecoderRoutesRegistry перечисляет команды, доступные подписчику бота.
func ControlCommandDecoderRoutesRegistry() []ports.ControlCommandDecoderRoute {
	return []ports.ControlCommandDecoderRoute{
		v1.StartBotCommandV1DecoderRoute(),
		v1.RunBotCommandV1DecoderRoute(),
	}
}
