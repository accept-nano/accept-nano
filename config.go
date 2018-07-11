package main

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
	// HTTP Basic Authentication user name for Node URL.
	NodeAuthUsername string
	// HTTP Basic Authentication password for Node URL.
	NodeAuthPassword string
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
	// Give some time to unfinished HTTP requests before shutting down the server.
	ShutdownTimeout uint
	// Limit payment creation requests to prevent DOS attack.
	RateLimit string
	// Payments below this amount are ignored.
	ReceiveThreshold string
	// Maximum number of payments allowed to fulfill the expected amount.
	MaxPayments int
	// Max allowed time (in seconds) for payment after it is created.
	AllowedDuration int
}
