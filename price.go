package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/cenkalti/log"
	"github.com/shopspring/decimal"
)

const (
	tickerURL = "https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest"
	nanoID    = "1567"
	// Coinmarketcap updates quotes every 60 seconds.
	priceUpdateInterval = 60 * time.Second
	priceFetchTimeout   = 10 * time.Second
)

type PriceWithTimestamp struct {
	Price     decimal.Decimal
	FetchedAt time.Time
}

var (
	errBadTickerResponse = errors.New("bad ticker response")

	// Cache price
	mPrice sync.Mutex
	prices = make(map[string]PriceWithTimestamp)

	priceClient = &http.Client{
		Timeout: priceFetchTimeout,
	}
)

type TickerResponse struct {
	Data map[string]struct {
		Quote map[string]struct {
			Price float64 `json:"price"`
		} `json:"quote"`
	} `json:"data"`
}

func getNanoPrice(currency string) (price decimal.Decimal, err error) {
	if config.CoinmarketcapAPIKey == "" {
		err = errors.New("empty CoinmarketcapAPIKey value in config")
		return
	}
	if currency == "" {
		currency = "USD"
	}
	currency = strings.ToUpper(currency)

	mPrice.Lock()
	defer mPrice.Unlock()

	if cached, ok := prices[currency]; ok && time.Since(cached.FetchedAt) < priceUpdateInterval {
		return cached.Price, nil
	}

	req, err := http.NewRequest("GET", tickerURL, nil)
	if err != nil {
		return
	}

	q := url.Values{}
	q.Add("id", nanoID)
	q.Add("convert", currency)

	req.Header.Set("Accepts", "application/json")
	req.Header.Add("X-CMC_PRO_API_KEY", config.CoinmarketcapAPIKey)
	req.URL.RawQuery = q.Encode()

	resp, err := priceClient.Do(req)
	if err != nil {
		return
	}
	defer func() {
		if err2 := resp.Body.Close(); err2 != nil {
			log.Debug(err2)
		}
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err = errBadTickerResponse
		return
	}
	var response TickerResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return
	}
	currencyVal, ok := response.Data[nanoID]
	if !ok {
		err = errors.New("bad currency")
		return
	}
	quoteVal, ok := currencyVal.Quote[currency]
	if !ok {
		err = errBadTickerResponse
		return
	}
	price = decimal.NewFromFloat(quoteVal.Price)
	if price.LessThanOrEqual(decimal.NewFromFloat(0)) {
		err = errors.New("bad price")
		return
	}
	// Cache new value
	prices[currency] = PriceWithTimestamp{
		Price:     price,
		FetchedAt: time.Now(),
	}
	return
}
