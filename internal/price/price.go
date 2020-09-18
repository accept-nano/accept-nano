package price

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
)

type priceWithTimestamp struct {
	Price     decimal.Decimal
	FetchedAt time.Time
}

var errBadTickerResponse = errors.New("bad ticker response")

type tickerResponse struct {
	Data map[string]struct {
		Quote map[string]struct {
			Price float64 `json:"price"`
		} `json:"quote"`
	} `json:"data"`
}

type API struct {
	apiKey        string
	cacheDuration time.Duration
	priceClient   http.Client

	// Cache price.
	mPrice sync.Mutex
	prices map[string]priceWithTimestamp
}

func NewAPI(apiKey string, clientTimeout, cacheDuration time.Duration) *API {
	return &API{
		apiKey:        apiKey,
		cacheDuration: cacheDuration,
		priceClient:   http.Client{Timeout: clientTimeout},
		prices:        make(map[string]priceWithTimestamp),
	}
}

func (p *API) GetNanoPrice(currency string) (price decimal.Decimal, err error) {
	if p.apiKey == "" {
		err = errors.New("empty CoinmarketcapAPIKey value in config")
		return
	}
	if currency == "" {
		currency = "USD"
	}
	currency = strings.ToUpper(currency)

	p.mPrice.Lock()
	defer p.mPrice.Unlock()

	if cached, ok := p.prices[currency]; ok && time.Since(cached.FetchedAt) < p.cacheDuration {
		return cached.Price, nil
	}

	req, err := http.NewRequest("GET", tickerURL, nil) // nolint:noctx // client timeout set
	if err != nil {
		return
	}

	q := url.Values{}
	q.Add("id", nanoID)
	q.Add("convert", currency)

	req.Header.Set("Accepts", "application/json")
	req.Header.Add("X-CMC_PRO_API_KEY", p.apiKey)
	req.URL.RawQuery = q.Encode()

	resp, err := p.priceClient.Do(req)
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
	var response tickerResponse
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
	p.prices[currency] = priceWithTimestamp{
		Price:     price,
		FetchedAt: time.Now(),
	}
	return
}
