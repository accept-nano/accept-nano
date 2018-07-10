package main

import (
	"github.com/accept-nano/accept-nano/nano"
	"github.com/cenkalti/log"
	"github.com/shopspring/decimal"
)

func receiveBlock(hash, amount, account, privateKey, publicKey string) error {
	sentAmount, err := decimal.NewFromString(amount)
	if err != nil {
		return err
	}
	log.Debugln("amount:", sentAmount)
	var newReceiverBlockPreviousHash string
	var newReceiverBalance decimal.Decimal
	var workHash string
	receiverAccountInfo, err := node.AccountInfo(account)
	switch err {
	case nano.ErrAccountNotFound:
		// First block in account chain. This is the common case.
		newReceiverBlockPreviousHash = "00000000000000000000000000000000"
		newReceiverBalance = sentAmount
		workHash = publicKey
	case nil:
		// More than one payment is made to the account.
		log.Debugf("account info: %#v", receiverAccountInfo)
		newReceiverBlockPreviousHash = receiverAccountInfo.Frontier
		currentReceiverBalance, err2 := decimal.NewFromString(receiverAccountInfo.Balance)
		if err2 != nil {
			return err2
		}
		newReceiverBalance = currentReceiverBalance.Add(sentAmount)
		workHash = newReceiverBlockPreviousHash
	default:
		return err
	}
	work, err := nano.GenerateWork(workHash)
	if err != nil {
		return err
	}
	newReceiverBlock, err := node.BlockCreate(newReceiverBlockPreviousHash, account, config.Representative, newReceiverBalance.String(), hash, privateKey, work)
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
