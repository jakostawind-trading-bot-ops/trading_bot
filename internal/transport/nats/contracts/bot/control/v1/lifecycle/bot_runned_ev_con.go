package transport_nats_contracts_bot_control_v1_lifecycle

import transport_nats_contracts_bot_control_v1_common "github.com/jakostawind-trading-bot-ops/trading_bot/internal/transport/nats/contracts/bot/control/v1/common"

type BotRunnedEventContractV1 struct {
	transport_nats_contracts_bot_control_v1_common.BaseBotEventContractV1
	PreviousBotStatus string `json:"previous_bot_status"`
	CurrentBotStatus  string `json:"current_bot_status"`
}

type BotRunnedErrorEventContractV1 struct {
	transport_nats_contracts_bot_control_v1_common.BaseBotEventContractV1
	BotStatus string `json:"bot_status"`
	Error     string `json:"error"`
}
