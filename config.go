package main

import (
	"github.com/BurntSushi/toml"
	"github.com/kelseyhightower/envconfig"
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
	NodeURL string `envconfig:"NODE_URL"`
	// Websocket URL of a running node.
	NodeWebsocketURL string `envconfig:"NODE_WEBSOCKET_URL"`
	// Timeout for requests made to Node URL (milliseconds).
	NodeTimeout uint
	// Funds will be sent to this address.
	Account string `envconfig:"ACCOUNT"`
	// Representative for created deposit accounts.
	Representative string
	// Seed to generate private keys from.
	// This is not your Account seed!
	// You can generate a new seed with -seed flag.
	// This seed will also be used for signing JWT tokens.
	Seed string `envconfig:"SEED"`
	// When customer sends the funds, merhchant will be notified at this URL.
	NotificationURL string
	// Give some time to unfinished HTTP requests before shutting down the server (milliseconds).
	ShutdownTimeout uint
	// Limit payment creation requests to prevent DOS attack.
	RateLimit string
	// Payments below this amount are ignored.
	ReceiveThreshold string
	// Maximum number of payments allowed to fulfill the expected amount.
	MaxPayments int
	// Up to this amount underpayments are accepted. Amount in NANO.
	UnderPaymentToleranceFixed float64
	// Up to this amount underpayments are accepted. Amount in percent.
	UnderPaymentTolerancePercent float64
	// Max allowed time for payment after it is created (seconds).
	AllowedDuration int
	// Parameter for calculating next check time of the payment.
	// Time passed since the creation of payment request is divided to this number.
	NextCheckDurationFactor int
	// Min allowed duration to check the payment (seconds).
	MinNextCheckDuration int
	// Max allowed duration to check the payment (seconds).
	MaxNextCheckDuration int
	// Password for accessing admin endpoints.
	// Admin endpoints are protected with HTTP basic auth. Username is "admin".
	AdminPassword string `envconfig:"ADMIN_PASSWORD"`
	// Coinmarketcap API Key
	CoinmarketcapAPIKey string
}

func (c *Config) Read() error {
	_, err := toml.DecodeFile(*configPath, c)
	if err != nil {
		return err
	}
	err = envconfig.Process("", c)
	if err != nil {
		return err
	}
	c.setDefaults()
	return nil
}

func (c *Config) setDefaults() {
	if c.DatabasePath == "" {
		c.DatabasePath = "accept-nano.db"
	}
	if c.ListenAddress == "" {
		c.ListenAddress = "127.0.0.1:8080"
	}
	if c.NodeURL == "" {
		c.NodeURL = "http://127.0.0.1:7076"
	}
	if c.NodeWebsocketURL == "" {
		c.NodeWebsocketURL = "ws://127.0.0.1:7078"
	}
	if c.NodeTimeout == 0 {
		c.NodeTimeout = 600000
	}
	if c.Representative == "" {
		c.Representative = "xrb_1nanode8ngaakzbck8smq6ru9bethqwyehomf79sae1k7xd47dkidjqzffeg"
	}
	if c.ShutdownTimeout == 0 {
		c.ShutdownTimeout = 5000
	}
	if c.RateLimit == "" {
		c.RateLimit = "60-H"
	}
	if c.ReceiveThreshold == "" {
		c.ReceiveThreshold = "0.001"
	}
	if c.MaxPayments == 0 {
		c.MaxPayments = 10
	}
	if c.AllowedDuration == 0 {
		c.AllowedDuration = 3600
	}
	if c.NextCheckDurationFactor == 0 {
		c.NextCheckDurationFactor = 20
	}
	if c.MinNextCheckDuration == 0 {
		c.MinNextCheckDuration = 10
	}
	if c.MaxNextCheckDuration == 0 {
		c.MaxNextCheckDuration = 1200
	}
}
