package actors

import (
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/entites"
	"github.com/jakostawind-trading-bot-ops/trading_bot/internal/states"
)

type GeneralActor struct {
	GeneralState      *states.GeneralState
	SubscribedMarkets []SubscribedMarket
}

// SubscribedMarket связывает рынок и команду подписки с актором этого рынка.
type SubscribedMarket struct {
	MarketID      string
	CorrelationID string
	MarketActor   *MarketActor
}

func (GeneralActor *GeneralActor) MarketSubscribe(AssetID string, Market entites.Market) {
	// Для первой подписки открыть канал рынков, для последующих обновить подписку.
	// Создать MarketActor и добавить запись в GeneralActor.SubscribedMarkets.
}
