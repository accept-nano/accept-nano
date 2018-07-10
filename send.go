package main

import (
	"github.com/cenkalti/accept-nano/nano"
	"github.com/cenkalti/log"
)

func sendAll(account, destination, privateKey string) error {
	log.Debugln("sending from", account)
	info, err := node.AccountInfo(account)
	if err != nil {
		return err
	}
	work, err := nano.GenerateWork(info.Frontier)
	if err != nil {
		return err
	}
	block, err := node.BlockCreate(info.Frontier, account, config.Representative, "0", destination, privateKey, work)
	if err != nil {
		return err
	}
	hash, err := node.Process(block)
	if err != nil {
		return err
	}
	log.Debugln("published new block:", hash)
	return nil
}
