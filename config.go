package main

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
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
	NodeURL string
	// Websocket URL of a running node.
	NodeWebsocketURL string
	// Disable subscribing confirmations over WebSocket
	DisableWebsocket bool
	// Timeout for requests made to Node URL.
	NodeTimeout time.Duration
	// Funds will be sent to this address.
	Account string
	// Representative for created deposit accounts.
	Representative string
	// Seed to generate private keys from.
	// This is not your Account seed!
	// You can generate a new seed with -seed flag.
	// This seed will also be used for signing JWT tokens.
	Seed string
	// When customer sends the funds, merhchant will be notified at this URL.
	NotificationURL string
	// Timeout for requests made to the merchant's NotificationURL
	NotificationRequestTimeout time.Duration
	// Give some time to unfinished HTTP requests before shutting down the server.
	ShutdownTimeout time.Duration
	// Limit payment creation requests to prevent DOS attack.
	RateLimit string
	// To protect against spam, payments below this amount are ignored.
	ReceiveThreshold string
	// Maximum number of payments allowed to fulfill the expected amount.
	MaxPayments int
	// Up to this amount underpayments are accepted. Amount in NANO.
	UnderPaymentToleranceFixed float64
	// Up to this amount underpayments are accepted. Amount in percent.
	UnderPaymentTolerancePercent float64
	// Max allowed time for payment after it is created.
	AllowedDuration time.Duration
	// Parameter for calculating next check time of the payment.
	// Time passed since the creation of payment request is divided to this number.
	NextCheckDurationFactor float64
	// Min allowed duration to check the payment.
	MinNextCheckDuration time.Duration
	// Max allowed duration to check the payment.
	MaxNextCheckDuration time.Duration
	// Password for accessing admin endpoints.
	// Admin endpoints are protected with HTTP basic auth. Username is "admin".
	// If no password is set, admin endpoints are disabled.
	AdminPassword string
	// Coinmarketcap API Key
	// https://coinmarketcap.com/api/documentation/v1/
	CoinmarketcapAPIKey string
	// Timeout for HTTP requests made to Coinmarketcap
	CoinmarketcapRequestTimeout time.Duration
	// Cache price value for a duration
	CoinmarketcapCacheDuration time.Duration
}

var DefaultConfig = Config{
	DatabasePath:                "accept-nano.db",
	ListenAddress:               "127.0.0.1:8080",
	NodeURL:                     "http://127.0.0.1:7076",
	NodeWebsocketURL:            "ws://127.0.0.1:7078",
	NodeTimeout:                 time.Minute,
	Representative:              "nano_1ninja7rh37ehfp9utkor5ixmxyg8kme8fnzc4zty145ibch8kf5jwpnzr3r",
	ShutdownTimeout:             5 * time.Second,
	RateLimit:                   "60-H",
	ReceiveThreshold:            "0.001",
	MaxPayments:                 10,
	AllowedDuration:             time.Hour,
	NextCheckDurationFactor:     20,
	MinNextCheckDuration:        10 * time.Second,
	MaxNextCheckDuration:        20 * time.Minute,
	CoinmarketcapRequestTimeout: 10 * time.Second,
	CoinmarketcapCacheDuration:  time.Minute,
	NotificationRequestTimeout:  time.Minute,
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
	err = k.Unmarshal("", &c)
	return
}
