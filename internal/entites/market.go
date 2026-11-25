package entites

import "github.com/shopspring/decimal"

type Market struct {
	ID          string
	Question    string
	QuestionID  string
	ConditionId string
	Slug        string

	StartDate     string
	EndDate       string
	GameStartTime string

	Tokens []MarketToken

	EventID string
}

type MarketToken struct {
	TokenID string
	Outcome string
	Price   decimal.Decimal
}
