package nano

import (
	"encoding/json"
	"errors"
)

type PendingBlock struct {
	Amount string `json:"amount"`
	Source string `json:"source"`
}

func (n *Node) Pending(account string, count int, threshold string) (map[string]PendingBlock, error) {
	args := map[string]interface{}{
		"account":   account,
		"count":     count,
		"threshold": threshold,
		"source":    "true",
	}
	var nodeResponse struct {
		Blocks *json.RawMessage `json:"blocks"`
	}
	err := n.call("pending", args, &nodeResponse)
	if err != nil {
		return nil, err
	}
	if nodeResponse.Blocks == nil {
		return nil, errors.New("invalid node response")
	}
	if string(*nodeResponse.Blocks) == "\"\"" {
		return nil, nil
	}
	ret := make(map[string]PendingBlock)
	err = json.Unmarshal(*nodeResponse.Blocks, &ret)
	return ret, err
}

func (n *Node) BlockCreate(previous, account, representative, balance, link, key, work string) (string, error) {
	args := map[string]interface{}{
		"type":           "state",
		"previous":       previous,
		"account":        account,
		"representative": representative,
		"balance":        balance,
		"link":           link,
		"key":            key,
		"work":           work,
	}
	var response struct {
		Hash  string `json:"hash"`
		Block string `json:"block"`
	}
	err := n.call("block_create", args, &response)
	if err != nil {
		return "", err
	}
	return response.Block, nil
}

func (n *Node) Process(block string) (string, error) {
	args := map[string]interface{}{
		"block": block,
	}
	var response struct {
		Hash string `json:"hash"`
	}
	err := n.call("process", args, &response)
	if err != nil {
		return "", err
	}
	return response.Hash, nil
}
