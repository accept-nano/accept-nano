package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/cenkalti/log"
	"github.com/shopspring/decimal"
	"github.com/ulule/limiter/drivers/middleware/stdlib"
)

func runServer() {
	ratelimitMiddleware := stdlib.NewMiddleware(rateLimiter, stdlib.WithForwardHeader(true))

	mux := http.NewServeMux()
	mux.Handle("/api/pay", ratelimitMiddleware.Handler(http.HandlerFunc(handlePay)))
	mux.HandleFunc("/api/verify", handleVerify)

	server.Addr = config.ListenAddress
	server.Handler = mux

	err := server.ListenAndServe()
	if err == http.ErrServerClosed {
		return
	}
	if err != nil {
		log.Fatal(err)
	}
}

type Response struct {
	Token            string          `json:"token"`
	Account          string          `json:"account"`
	Amount           decimal.Decimal `json:"amount"`
	AmountInCurrency decimal.Decimal `json:"amountInCurrency"`
	Currency         string          `json:"currency"`
	Balance          decimal.Decimal `json:"balance"`
	State            string          `json:"state"`
	Fulfilled        bool            `json:"fulfilled"`
}

func handlePay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	var amount decimal.Decimal
	amountInCurrency, err := decimal.NewFromString(r.FormValue("amount"))
	if err != nil {
		log.Debug(err)
		http.Error(w, "invalid amount", http.StatusBadRequest)
		return
	}
	currency := r.FormValue("currency")
	if currency != "" {
		price, err2 := getNanoPrice(currency)
		if err2 != nil {
			log.Error(err2)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		amount = amountInCurrency.DivRound(price, 6)
	} else {
		amount = amountInCurrency
		currency = "NANO"
	}
	currency = strings.ToUpper(currency)
	index, err := NewIndex()
	if err != nil {
		log.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	key, err := node.DeterministicKey(config.Seed, index)
	if err != nil {
		log.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	token, err := NewToken(index)
	if err != nil {
		log.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	payment := Payment{
		PublicKey:        key.Public,
		Account:          key.Account,
		Amount:           NanoToRaw(amount),
		AmountInCurrency: NanoToRaw(amountInCurrency),
		Currency:         currency,
		State:            r.FormValue("state"),
		CreatedAt:        time.Now().UTC(),
		token:            token,
	}
	err = payment.Save()
	if err != nil {
		log.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	payment.StartChecking()
	response := Response{
		Token:            token,
		Account:          key.Account,
		Amount:           RawToNano(payment.Amount),
		AmountInCurrency: RawToNano(payment.AmountInCurrency),
		Currency:         payment.Currency,
		State:            payment.State,
	}
	b, err := json.Marshal(&response)
	if err != nil {
		log.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	log.Debugf("created new payment: %s", b)
	_, err = w.Write(b)
	if err != nil {
		log.Debug(err)
	}
}

func handleVerify(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	payment, err := LoadPayment([]byte(token))
	if err == errPaymentNotFound {
		log.Debugln("token not found:", token)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	if err != nil {
		log.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	response := Response{
		Token:            payment.token,
		Account:          payment.Account,
		Amount:           RawToNano(payment.Amount),
		AmountInCurrency: RawToNano(payment.AmountInCurrency),
		Currency:         payment.Currency,
		Balance:          RawToNano(payment.Balance),
		State:            payment.State,
		Fulfilled:        payment.FulfilledAt != nil,
	}
	b, err := json.Marshal(&response)
	if err != nil {
		log.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	_, err = w.Write(b)
	if err != nil {
		log.Debug(err)
	}
}
