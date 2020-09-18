package main

import (
	"github.com/accept-nano/accept-nano/internal/nano"
	"github.com/accept-nano/accept-nano/internal/units"
	"github.com/cenkalti/log"
	"github.com/shopspring/decimal"
)

func receiveBlock(hash string, amount decimal.Decimal, account, privateKey, publicKey string) error {
	log.Debugln("amount:", units.RawToNano(amount).String())
	var newReceiverBlockPreviousHash string
	var newReceiverBalance decimal.Decimal
	var workHash string
	receiverAccountInfo, err := node.AccountInfo(account)
	switch err {
	case nano.ErrAccountNotFound:
		// First block in account chain. This is the common case.
		newReceiverBlockPreviousHash = "0000000000000000000000000000000000000000000000000000000000000000"
		newReceiverBalance = amount
		workHash = publicKey
	case nil:
		// More than one payment is made to the account.
		log.Debugf("account info: %#v", receiverAccountInfo)
		newReceiverBlockPreviousHash = receiverAccountInfo.Frontier
		newReceiverBalance = receiverAccountInfo.Balance.Add(amount)
		workHash = newReceiverBlockPreviousHash
	default:
		return err
	}
	work, err := nano.GenerateWork(workHash, false)
	if err != nil {
		return err
	}
	newReceiverBlock, err := node.BlockCreate(newReceiverBlockPreviousHash, account, config.Representative, newReceiverBalance, hash, privateKey, work)
	if err != nil {
		return err
	}
	log.Debugf("new block: %#v", newReceiverBlock)
	newHash, err := node.Process(newReceiverBlock)
	if err != nil {
		return err
	}
	log.Debugln("published new block:", newHash)
	return nil
}
