package main

import (
	"time"

	"github.com/shopspring/decimal"
)

type Notification struct {
	Account          string          `json:"account"`
	Amount           decimal.Decimal `json:"amount"`
	AmountInCurrency decimal.Decimal `json:"amountInCurrency"`
	Currency         string          `json:"currency"`
	Balance          decimal.Decimal `json:"balance"`
	State            string          `json:"state"`
	Fulfilled        bool            `json:"fulfilled"`
	FulfilledAt      *time.Time      `json:"fulfillAt"`
}
