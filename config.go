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
	// HTTP Basic Authentication user name for Node URL.
	NodeAuthUsername string `envconfig:"NODE_AUTH_USERNAME"`
	// HTTP Basic Authentication password for Node URL.
	NodeAuthPassword string `envconfig:"NODE_AUTH_PASSWORD"`
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
	// Max allowed time for payment after it is created (seconds).
	AllowedDuration int
	// Password for accessing admin endpoints.
	// Admin endpoints are protected with HTTP basic auth. Username is "admin".
	AdminPassword string `envconfig:"ADMIN_PASSWORD"`
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
}
