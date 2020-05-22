package main

import (
	"time"

	"github.com/accept-nano/accept-nano/nano"
	"github.com/cenkalti/log"
)

func runSubscriber() {
	for {
		err := subscribe()
		if err != nil {
			log.Errorf("websocket error: %s", err.Error())
			time.Sleep(time.Second)
		}
	}
}

func subscribe() (err error) {
	ws := nano.NewWebsocket(config.NodeWebsocketURL)
	err = ws.Connect()
	if err != nil {
		return err
	}
	defer ws.Close()
	err = ws.Send("subscribe", "confirmation", true, map[string]interface{}{"include_election_info": "false", "include_block": "true"})
	if err != nil {
		return err
	}
	var msg struct {
		Message struct {
			FromAccount string `json:"account"`
			Block       struct {
				ToAccount string `json:"link_as_account"`
			} `json:"block"`
		} `json:"message"`
	}
	for {
		err := ws.Recv(&msg)
		if err != nil {
			return err
		}
		confirmations <- msg.Message.Block.ToAccount
	}
}

func runChecker() {
	for account := range confirmations {
		p, err := LoadPayment([]byte(account))
		if err == errPaymentNotFound {
			continue
		}
		if err != nil {
			log.Errorf("cannot load payment: %s", err.Error())
			continue
		}
		log.Debugf("received confirmation from websocket, checking account: %s", account)
		go p.checkOnce()
	}
}
