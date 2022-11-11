package main

import (
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/mitchellh/mapstructure"
	"github.com/shopspring/decimal"
)

type Config struct {
	// Print debug level log messages to console.
	EnableDebugLog bool
	// Created payment requests are saved in this database. Do not lose this file.
	DatabasePath string
	// Listen address for HTTP server.
	ListenAddress string
	// Optional TLS certificate and key if you want to serve over HTTPS.
	CertFile, KeyFile string
	// URL of a running node.
	// You can use your own node or any public node proxy service.
	NodeURL string
	// Websocket URL of a running node.
	// With the help of Websocket support, sent payment can be detected immediatlely on the network.
	// Otherwise, accept-nano needs to poll the node for checking account balances.
	NodeWebsocketURL string
	// Disable subscribing confirmations over WebSocket
	// Only polling method will be used for checking balances.
	DisableWebsocket bool
	// Time to wait for connect and handshake to be completed.
	NodeWebsocketHandshakeTimeout time.Duration
	// TCP write timeout for Websocket connection
	NodeWebsocketWriteTimeout time.Duration
	// Time to wait for a subscription action to be completed.
	NodeWebsocketAckTimeout time.Duration
	// Duration between 2 keepalive messages.
	NodeWebsocketKeepAlivePeriod time.Duration
	// Timeout for requests made to Node URL.
	NodeTimeout time.Duration
	// Authorization HTTP header value for node requests.
	NodeAuthorizationHeader string
	// api-key header value for nano.nownodes.io service.
	NodeAPIKeyHeader string
	// Set this to your merchant account. Received funds will be sent to this address.
	Account string
	// Representative for created deposit accounts.
	// It is not that important for the network because after funds are detected in deposit account,
	// they will be transferred to your merchant account immediatlely.
	Representative string
	// Seed to generate the private key for deposit accounts.
	// Do not put your merchant account seed here!
	// You can generate a new seed with -seed flag.
	// This seed will also be used for signing JWT tokens.
	Seed string
	// When customer sends the funds, merhchant will be notified at this URL.
	NotificationURL string
	// Timeout for requests made to the merchant's NotificationURL
	NotificationRequestTimeout time.Duration
	// On shutdown of the server, give some time to unfinished HTTP requests before shutting down the server.
	ShutdownTimeout time.Duration
	// Limit payment creation requests to prevent DOS attack.
	RateLimit string
	// To protect against spam, payments below this amount are ignored and not going to be processed.
	ReceiveThreshold decimal.Decimal
	// Maximum number of payments allowed to fulfill the expected amount. Limited to prevent DOS.
	MaxPayments int
	// Up to this amount underpayments are accepted. Amount in NANO.
	UnderPaymentToleranceFixed decimal.Decimal
	// Up to this amount underpayments are accepted. Amount in percent.
	UnderPaymentTolerancePercent float64
	// Max allowed time for payment after it is created.
	// Payment is not going to be checked automatically after this duration passes from creation
	// but it can always be triggered from admin endpoint manually.
	AllowedDuration time.Duration
	// Parameter for calculating next check time of the payment.
	// Time passed since the creation of payment request is divided to this number.
	// For example, for a factor value of 20, if a minute has passed after creation, then the next check will be after 60/20=3 seconds.
	NextCheckDurationFactor float64
	// Min allowed duration to check the payment.
	// Value calculated using NextCheckDurationFactor cannot be smalled than this value.
	MinNextCheckDuration time.Duration
	// Max allowed duration to check the payment.
	// Value calculated using NextCheckDurationFactor cannot be larger than this value.
	MaxNextCheckDuration time.Duration
	// Password for accessing admin endpoints.
	// Admin endpoints are protected with HTTP basic auth. Username is always "admin".
	// If no password is set, admin endpoints are disabled.
	AdminPassword string
	// Coinmarketcap API Key for getting the price conversion for fiat moneys.
	// Get API key from: https://coinmarketcap.com/api/documentation/v1/
	CoinmarketcapAPIKey string
	// Timeout for HTTP requests made to Coinmarketcap
	CoinmarketcapRequestTimeout time.Duration
	// Cache price value for a duration
	CoinmarketcapCacheDuration time.Duration
}

var DefaultConfig = Config{
	DatabasePath:                  "accept-nano.db",
	ListenAddress:                 "127.0.0.1:8080",
	NodeURL:                       "http://127.0.0.1:7076",
	NodeWebsocketURL:              "ws://127.0.0.1:7078",
	NodeWebsocketHandshakeTimeout: 10 * time.Second,
	NodeWebsocketWriteTimeout:     10 * time.Second,
	NodeWebsocketAckTimeout:       10 * time.Second,
	NodeWebsocketKeepAlivePeriod:  time.Minute,
	NodeTimeout:                   time.Minute,
	Representative:                "nano_1ninja7rh37ehfp9utkor5ixmxyg8kme8fnzc4zty145ibch8kf5jwpnzr3r",
	ShutdownTimeout:               5 * time.Second,
	RateLimit:                     "60-H",
	ReceiveThreshold:              decimal.RequireFromString("0.001"),
	MaxPayments:                   10,
	AllowedDuration:               time.Hour,
	NextCheckDurationFactor:       20,
	MinNextCheckDuration:          10 * time.Second,
	MaxNextCheckDuration:          20 * time.Minute,
	CoinmarketcapRequestTimeout:   10 * time.Second,
	CoinmarketcapCacheDuration:    time.Minute,
	NotificationRequestTimeout:    time.Minute,
}

func (c *Config) Read() (err error) {
	*c = DefaultConfig
	k := koanf.New(".")
	var parser koanf.Parser
	ext := filepath.Ext(*configPath)
	if ext == ".yaml" || ext == ".yml" {
		parser = yaml.Parser()
	} else {
		parser = toml.Parser()
	}
	err = k.Load(file.Provider(*configPath), parser)
	if err != nil {
		return
	}
	err = k.Load(env.Provider("ACCEPTNANO_", ".", func(s string) string {
		return strings.Replace(strings.TrimPrefix(s, "ACCEPTNANO_"), "_", ".", -1)
	}), nil)
	if err != nil {
		return
	}
	conf := koanf.UnmarshalConf{
		DecoderConfig: &mapstructure.DecoderConfig{
			DecodeHook:       mapstructure.ComposeDecodeHookFunc(mapstructure.StringToTimeDurationHookFunc(), StringToDecimalHookFunc(), Float64ToDecimalHookFunc()),
			WeaklyTypedInput: true,
			Result:           c,
		},
	}
	err = k.UnmarshalWithConf("", c, conf)
	return
}

func StringToDecimalHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(decimal.Decimal{}) {
			return data, nil
		}
		return decimal.NewFromString(data.(string))
	}
}

func Float64ToDecimalHookFunc() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.Float64 {
			return data, nil
		}
		if t != reflect.TypeOf(decimal.Decimal{}) {
			return data, nil
		}
		return decimal.NewFromFloat(data.(float64)), nil
	}
}
