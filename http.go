package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/accept-nano/accept-nano/internal/hub"
	"github.com/accept-nano/accept-nano/internal/units"
	"github.com/cenkalti/log"
	"github.com/rs/cors"
	"github.com/shopspring/decimal"
	"github.com/ulule/limiter/v3/drivers/middleware/stdlib"
	"golang.org/x/net/websocket"
)

func runServer() {
	ratelimitMiddleware := stdlib.NewMiddleware(rateLimiter)

	mux := http.NewServeMux()
	mux.HandleFunc("/version", handleVersion)
	mux.Handle("/api/pay", ratelimitMiddleware.Handler(http.HandlerFunc(handlePay)))
	mux.Handle("/api/price", ratelimitMiddleware.Handler(http.HandlerFunc(handlePrice)))
	mux.HandleFunc("/api/verify", handleVerify)
	mux.Handle("/websocket", websocket.Handler(handleWebsocket))
	if config.AdminPassword != "" {
		mux.HandleFunc("/admin/payments/active", handleAdminGetActivePayments)
		mux.HandleFunc("/admin/payment", handleAdminGetPayment)
		mux.HandleFunc("/admin/check", handleAdminCheckPayment)
		mux.HandleFunc("/admin/receive", handleAdminReceivePending)
		mux.HandleFunc("/admin/send", handleAdminSendToMerchant)
	}

	server.Addr = config.ListenAddress
	server.Handler = cors.Default().Handler(mux)

	var err error
	if config.CertFile != "" && config.KeyFile != "" {
		err = server.ListenAndServeTLS(config.CertFile, config.KeyFile)
	} else {
		err = server.ListenAndServe()
	}
	if err == http.ErrServerClosed {
		return
	}
	log.Fatal(err)
}

func handleVersion(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(version))
	if err != nil {
		log.Debug(err)
	}
}

func handlePrice(w http.ResponseWriter, r *http.Request) {
	currency := r.FormValue("currency")
	price, err := priceAPI.GetNanoPrice(currency)
	if err != nil {
		log.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	b, err := json.Marshal(map[string]interface{}{"price": price})
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
		price, err2 := priceAPI.GetNanoPrice(currency)
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
	_, err = LoadPayment([]byte(key.Account))
	if err != nil {
		if err != errPaymentNotFound {
			log.Error(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	} else {
		log.Errorln("index collision:", index)
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	token, err := NewToken(index)
	if err != nil {
		log.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	payment := &Payment{
		account:          key.Account,
		Index:            index,
		Amount:           units.NanoToRaw(amount),
		AmountInCurrency: amountInCurrency,
		Currency:         currency,
		State:            r.FormValue("state"),
		CreatedAt:        time.Now().UTC(),
	}
	err = payment.Save()
	if err != nil {
		log.Error(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	payment.StartChecking()
	response := NewResponse(payment, token)
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
	if token == "" {
		http.Error(w, "invalid token", http.StatusBadRequest)
		return
	}
	claims, err := ParseToken(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusBadRequest)
		return
	}
	key, err := node.DeterministicKey(config.Seed, claims.Index)
	if err != nil {
		http.Error(w, "invalid token", http.StatusBadRequest)
		return
	}
	payment, err := LoadPayment([]byte(key.Account))
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
	response := NewResponse(payment, token)
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

func handleWebsocket(conn *websocket.Conn) {
	r := conn.Request()
	token := r.FormValue("token")
	if token == "" {
		return
	}
	claims, err := ParseToken(token)
	if err != nil {
		return
	}
	key, err := node.DeterministicKey(config.Seed, claims.Index)
	if err != nil {
		return
	}
	cancel := verifications.Subscribe(hub.Account(key.Account), func(e hub.Event) {
		pv := e.(PaymentVerified)
		response := NewResponse(&pv.Payment, token)
		b, err := json.Marshal(&response)
		if err != nil {
			return
		}
		_, _ = conn.Write(b)
	})
	defer cancel()
	const readBufferSize = 1024
	buf := make([]byte, readBufferSize)
	for {
		_, err := conn.Read(buf)
		if err != nil {
			return
		}
	}
}

type PaymentVerified struct {
	Payment
}

func (p PaymentVerified) Account() hub.Account {
	return hub.Account(p.Payment.account)
}
