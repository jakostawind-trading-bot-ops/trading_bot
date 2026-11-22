package transport_nats_commands_decoder_routes_v1

import (
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/ports"
	contracts "github.com/jakostawind-trading-bot-ops/trading_bot/internal/transport/nats/contracts/control/bot/v1/lifecycle"
)

// StartBotCommandV1DecoderRoute сохраняет маршрут существующей команды StartBot.
func StartBotCommandV1DecoderRoute() ports.ControlCommandDecoderRoute {
	return ports.ControlCommandDecoderRoute{MessageType: "v1.lifecycle.start_bot_command", Decode: contracts.DecodeStartBotCommand}
}

// RunBotCommandV1DecoderRoute связывает команду RunBot с её JSON-контрактом.
func RunBotCommandV1DecoderRoute() ports.ControlCommandDecoderRoute {
	return ports.ControlCommandDecoderRoute{MessageType: "v1.lifecycle.run_bot_command", Decode: contracts.DecodeRunBotCommand}
}
