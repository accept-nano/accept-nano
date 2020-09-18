package main

import (
	"time"

	"github.com/accept-nano/accept-nano/internal/units"
	"github.com/shopspring/decimal"
)

// Response that we return from API endpoints.
type Response struct {
	Token            string                        `json:"token"`
	Account          string                        `json:"account"`
	Amount           decimal.Decimal               `json:"amount"`
	AmountInCurrency decimal.Decimal               `json:"amountInCurrency"`
	Currency         string                        `json:"currency"`
	Balance          decimal.Decimal               `json:"balance"`
	SubPayments      map[string]SubPaymentResponse `json:"subPayments"`
	RemainingSeconds int                           `json:"remainingSeconds"`
	State            string                        `json:"state"`
	Fulfilled        bool                          `json:"fulfilled"`
	MerchantNotified bool                          `json:"merchantNotified"`
}

type SubPaymentResponse struct {
	Amount  decimal.Decimal `json:"amount"`
	Account string          `json:"account"`
}

func NewResponse(p *Payment, token string) *Response {
	subPayments := make(map[string]SubPaymentResponse, len(p.SubPayments))
	for k, v := range p.SubPayments {
		subPayments[k] = SubPaymentResponse{Account: v.Account, Amount: units.RawToNano(v.Amount)}
	}
	return &Response{
		Token:            token,
		Account:          p.Account,
		Amount:           units.RawToNano(p.Amount),
		AmountInCurrency: p.AmountInCurrency,
		Currency:         p.Currency,
		Balance:          units.RawToNano(p.Balance),
		State:            p.State,
		SubPayments:      subPayments,
		RemainingSeconds: int(p.remainingDuration() / time.Second),
		Fulfilled:        p.FulfilledAt != nil,
		MerchantNotified: p.NotifiedAt != nil,
	}
}
