package nano

import (
	"errors"

	"github.com/shopspring/decimal"
)

type AccountInfo struct {
	Frontier            string          `json:"frontier"`
	OpenBlock           string          `json:"open_block"`
	RepresentativeBlock string          `json:"representative_block"`
	Balance             decimal.Decimal `json:"balance"`
	ModifiedTimestamp   string          `json:"modified_timestamp"`
	BlockCount          string          `json:"block_count"`
	Representative      string          `json:"representative"`
}

var ErrAccountNotFound = errors.New("account not found")

func (n *Node) AccountInfo(account string) (*AccountInfo, error) {
	args := map[string]interface{}{
		"account": account,
	}
	var response AccountInfo
	err := n.call("account_info", args, &response)
	if err2, ok := err.(*NodeError); ok && err2.Error() == "Account not found" {
		return nil, ErrAccountNotFound
	}
	return &response, err
}
