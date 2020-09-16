package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/accept-nano/accept-nano/nano"
	"github.com/cenkalti/log"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"go.etcd.io/bbolt"
)

const paymentsBucket = "payments"

var (
	Version           = ""
	generateSeed      = flag.Bool("seed", false, "generate a seed and exit")
	configPath        = flag.String("config", "config.toml", "config file path")
	version           = flag.Bool("version", false, "display version and exit")
	config            Config
	db                *bbolt.DB
	server            http.Server
	rateLimiter       *limiter.Limiter
	node              *nano.Node
	stopCheckPayments = make(chan struct{})
	checkPaymentWG    sync.WaitGroup
	confirmations     = make(chan string)
	verifications     Hub
)

func main() {
	if Version == "" {
		Version = "v0.0.0"
	}

	flag.Parse()

	if *version {
		fmt.Println(Version)
		return
	}

	if *generateSeed {
		seed, err := NewSeed()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(seed)
		return
	}

	err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	if config.EnableDebugLog {
		log.SetLevel(log.DEBUG)
	}

	if config.CoinmarketcapAPIKey == "" {
		log.Warning("empty CoinmarketcapAPIKey in config, fiat conversions will not work")
	}

	rate, err := limiter.NewRateFromFormatted(config.RateLimit)
	if err != nil {
		log.Fatal(err)
	}

	rateLimiter = limiter.New(memory.NewStore(), rate, limiter.WithTrustForwardHeader(true))
	node = nano.New(config.NodeURL)
	node.SetTimeout(time.Duration(config.NodeTimeout) * time.Millisecond)

	notificationClient.Timeout = config.NotificationRequestTimeout
	priceClient.Timeout = config.CoinmarketcapRequestTimeout

	log.Debugln("opening db:", config.DatabasePath)
	db, err = bbolt.Open(config.DatabasePath, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Debugln("db has been opened successfully")

	err = db.Update(func(tx *bbolt.Tx) error {
		_, txErr := tx.CreateBucketIfNotExists([]byte(paymentsBucket))
		return txErr
	})
	if err != nil {
		log.Fatal(err)
	}

	// Check existing payments.
	payments, err := LoadActivePayments()
	if err != nil {
		log.Fatal(err)
	}
	for _, p := range payments {
		p.StartChecking()
	}

	if !config.DisableWebsocket && config.NodeWebsocketURL != "" {
		go runSubscriber()
		go runChecker()
	}

	go runServer()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	close(stopCheckPayments)

	shutdownTimeout := time.Duration(config.ShutdownTimeout) * time.Millisecond
	log.Noticeln("shutting down with timeout:", shutdownTimeout)

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	err = server.Shutdown(ctx)
	if err != nil {
		log.Errorln("shutdown error:", err)
	}

	checkPaymentWG.Wait()

	err = db.Close()
	if err != nil {
		log.Fatal(err)
	}
}
