package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/cenkalti/log"
	"github.com/coreos/bbolt"
	"github.com/shopspring/decimal"
)

var (
	errPaymentNotFound     = errors.New("payment not found")
	errPaymentNotFulfilled = errors.New("payment not fulfilled")
	errNoPendingBlock      = errors.New("no pending block")
)

// Payment is the data type stored in the database in JSON format.
type Payment struct {
	// Customer sends money to this account.
	Account string `json:"account"`
	// Public key of Account.
	PublicKey string `json:"publicKey"`
	// Index for generating deterministic key.
	Index string `json:"index"`
	// Currency of amount in original request.
	Currency string `json:"currency"`
	// Original amount requested by client. Amount * Price(Currency)
	AmountInCurrency decimal.Decimal `json:"amountInCurrency"`
	// In NANO currency. Payment is fulfilled when Account contains this amount.
	Amount decimal.Decimal `json:"amount"`
	// Current balance in Account
	Balance decimal.Decimal `json:"balance"`
	// Free text field to pass from customer to merchant.
	State string `json:"state"`
	// Set when customer created the payment request via API.
	CreatedAt time.Time `json:"createdAt"`
	// Set every time Account is checked for incoming funds.
	LastCheckedAt *time.Time `json:"lastCheckedAt"`
	// Set when detected customer has sent enough funds to Account.
	FulfilledAt *time.Time `json:"fulfilledAt"`
	// Set when merchant is notified.
	NotifiedAt *time.Time `json:"notifiedAt"`
	// Set when pending funds are accepted to Account.
	ReceivedAt *time.Time `json:"receivedAt"`
	// Set when Amount is sent to the merchant account.
	SentAt *time.Time `json:"sentAt"`
}

// LoadPayment fetches a Payment object from database by key.
func LoadPayment(key []byte) (*Payment, error) {
	var value []byte
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(paymentsBucket))
		v := b.Get(key)
		if v == nil {
			return nil
		}
		value = make([]byte, len(v))
		copy(value, v)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, errPaymentNotFound
	}
	var payment Payment
	err = json.Unmarshal(value, &payment)
	return &payment, err
}

// Save the Payment object in database.
func (p *Payment) Save() error {
	key := []byte(p.Account)
	value, err := json.Marshal(&p)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(paymentsBucket))
		return b.Put(key, value)
	})
}

// NextCheck returns the next timestamp payment should be checked at.
func (p Payment) NextCheck() time.Duration {
	if p.LastCheckedAt == nil {
		return 0
	}
	create := p.CreatedAt
	lastCheck := *p.LastCheckedAt

	now := time.Now().UTC()
	minWait := 2 * time.Second
	maxWait := 20 * time.Minute
	passed := now.Sub(create)
	next := passed / 20
	if next < minWait {
		next = minWait
	} else if next > maxWait {
		next = maxWait
	}
	nextCheck := lastCheck.Add(next)
	return nextCheck.Sub(now)
}

func (p Payment) finished() bool {
	if now().Sub(p.CreatedAt) > time.Duration(config.AllowedDuration)*time.Second {
		return true
	}
	return p.SentAt != nil
}

func (p Payment) remainingDuration() time.Duration {
	allow := time.Duration(config.AllowedDuration) * time.Second
	return p.CreatedAt.Add(allow).Sub(*now())
}

// StartChecking starts a goroutine to check the payment periodically.
func (p *Payment) StartChecking() {
	if p.finished() {
		return
	}
	checkPaymentWG.Add(1)
	go p.checkLoop()
}

func (p *Payment) checkLoop() {
	defer checkPaymentWG.Done()
	for {
		if p.finished() {
			return
		}
		select {
		case <-time.After(p.NextCheck()):
			err := p.check()
			if err != nil {
				log.Error(err)
			}
		case <-stopCheckPayments:
			return
		}
	}
}

func (p *Payment) check() error {
	log.Debugln("checking payment:", p.Account)
	err := p.process()
	switch err {
	case errNoPendingBlock, errPaymentNotFulfilled:
		log.Debug(err)
		return nil
	case nil:
		p.LastCheckedAt = now()
		return p.Save()
	default:
		return err
	}
}

var locks = NewMapLock()

func (p *Payment) process() error {
	locks.Lock(p.Account)
	defer locks.Unlock(p.Account)

	if p.SentAt == nil {
		if p.ReceivedAt == nil {
			if p.NotifiedAt == nil {
				if p.FulfilledAt == nil {
					err := p.checkPending()
					if err != nil {
						return err
					}
					p.FulfilledAt = now()
					err = p.Save()
					if err != nil {
						return err
					}
				}
				err := p.notifyMerchant()
				if err != nil {
					return err
				}
				p.NotifiedAt = now()
				err = p.Save()
				if err != nil {
					return err
				}
			}
			err := p.receivePending()
			if err != nil {
				return err
			}
			p.ReceivedAt = now()
			err = p.Save()
			if err != nil {
				return err
			}
		}
		err := p.sendToMerchant()
		if err != nil {
			return err
		}
		p.SentAt = now()
		err = p.Save()
		if err != nil {
			return err
		}
	}
	return nil
}

func now() *time.Time {
	t := time.Now().UTC()
	return &t
}

func (p *Payment) checkPending() error {
	threshold, err := decimal.NewFromString(config.ReceiveThreshold)
	if err != nil {
		return err
	}
	pendingBlocks, err := node.Pending(p.Account, config.MaxPayments, NanoToRaw(threshold).String())
	if err != nil {
		return err
	}
	if len(pendingBlocks) == 0 {
		return errNoPendingBlock
	}
	var totalAmount decimal.Decimal
	for hash, pendingBlock := range pendingBlocks {
		log.Debugf("received new block: %#v", hash)
		amount, err2 := decimal.NewFromString(pendingBlock.Amount)
		if err2 != nil {
			return err2
		}
		log.Debugln("amount:", amount)
		totalAmount = totalAmount.Add(amount)
	}
	log.Debugln("total amount:", totalAmount)
	if p.Balance != totalAmount {
		p.Balance = totalAmount
		err = p.Save()
		if err != nil {
			return err
		}
	}
	if p.Balance.LessThan(p.Amount) {
		return errPaymentNotFulfilled
	}
	return nil
}

func (p *Payment) receivePending() error {
	threshold, err := decimal.NewFromString(config.ReceiveThreshold)
	if err != nil {
		return err
	}
	pendingBlocks, err := node.Pending(p.Account, config.MaxPayments, NanoToRaw(threshold).String())
	if err != nil {
		return err
	}
	if len(pendingBlocks) == 0 {
		return errNoPendingBlock
	}
	key, err := node.DeterministicKey(config.Seed, p.Index)
	if err != nil {
		return err
	}
	for hash, pendingBlock := range pendingBlocks {
		err = receiveBlock(hash, pendingBlock.Amount, p.Account, key.Private, p.PublicKey)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Payment) sendToMerchant() error {
	key, err := node.DeterministicKey(config.Seed, p.Index)
	if err != nil {
		return err
	}
	return sendAll(p.Account, config.Account, key.Private)
}

func (p *Payment) notifyMerchant() error {
	if config.NotificationURL == "" {
		return nil
	}
	notification := Notification{
		Account:          p.Account,
		Amount:           RawToNano(p.Amount),
		AmountInCurrency: p.AmountInCurrency,
		Currency:         p.Currency,
		Balance:          RawToNano(p.Balance),
		State:            p.State,
		Fulfilled:        p.FulfilledAt != nil,
		FulfilledAt:      p.FulfilledAt,
	}
	data, err := json.Marshal(notification)
	if err != nil {
		return err
	}
	resp, err := http.Post(config.NotificationURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer func() {
		if err2 := resp.Body.Close(); err2 != nil {
			log.Debug(err2)
		}
	}()
	if resp.StatusCode != 200 {
		return errors.New("bad notification response")
	}
	return nil
}
