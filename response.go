package main

import (
	"time"

	"github.com/shopspring/decimal"
)

// Response that we return from API endpoints.
type Response struct {
	Token            string          `json:"token"`
	Account          string          `json:"account"`
	Amount           decimal.Decimal `json:"amount"`
	AmountInCurrency decimal.Decimal `json:"amountInCurrency"`
	Currency         string          `json:"currency"`
	Balance          decimal.Decimal `json:"balance"`
	RemainingSeconds int             `json:"remainingSeconds"`
	State            string          `json:"state"`
	Fulfilled        bool            `json:"fulfilled"`
	MerchantNotified bool            `json:"merchantNotified"`
}

func NewResponse(p *Payment) *Response {
	return &Response{
		Token:            p.token,
		Account:          p.Account,
		Amount:           RawToNano(p.Amount),
		AmountInCurrency: p.AmountInCurrency,
		Currency:         p.Currency,
		Balance:          RawToNano(p.Balance),
		State:            p.State,
		RemainingSeconds: int(p.remainingDuration() / time.Second),
		Fulfilled:        p.FulfilledAt != nil,
		MerchantNotified: p.NotifiedAt != nil,
	}
}
