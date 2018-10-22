package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/cenkalti/log"
	"github.com/shopspring/decimal"
)

const tickerURL = "https://api.coinmarketcap.com/v2/ticker/1567/?convert="

var errBadTickerResponse = errors.New("bad ticker response")

type TickerResponse struct {
	Data struct {
		Quotes map[string]interface{}
	}
}

func getNanoPrice(currency string) (price decimal.Decimal, err error) {
	if currency == "" {
		currency = "USD"
	}
	currency = strings.ToUpper(currency)
	url := tickerURL + currency
	resp, err := http.Get(url) // nolint: gosec
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
	currencyValue, ok := response.Data.Quotes[currency]
	if !ok {
		err = errors.New("bad currency")
		return
	}
	currencyMap, ok := currencyValue.(map[string]interface{})
	if !ok {
		err = errBadTickerResponse
		return
	}
	priceValue, ok := currencyMap["price"]
	if !ok {
		err = errBadTickerResponse
		return
	}
	priceNumber, ok := priceValue.(float64)
	if !ok {
		err = errBadTickerResponse
		return
	}
	price = decimal.NewFromFloat(priceNumber)
	if price.LessThanOrEqual(decimal.NewFromFloat(0)) {
		err = errors.New("bad price")
		return
	}
	return
}
