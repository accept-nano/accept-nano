package nano

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"strings"

	ed25519 "github.com/accept-nano/ed25519-blake2b"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/blake2b"
)

type PendingBlock struct {
	Amount decimal.Decimal `json:"amount"`
	Source string          `json:"source"`
}

func (n *Node) Pending(account string, count int, threshold decimal.Decimal) (map[string]PendingBlock, error) {
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

func (n *Node) BlockCreate(previous, account, representative string, balance decimal.Decimal, link, key, work string) (string, error) { // nolint:interfacer
	block, err := blockCreate(previous, account, representative, balance.String(), link, key, work)
	return block.Block.String(), err
}

type createdBlock struct {
	Block blockType
	Hash  string
}

type blockType struct {
	Type           string `json:"type"`
	Account        string `json:"account"`
	Previous       string `json:"previous"`
	Representative string `json:"representative"`
	Balance        string `json:"balance"`
	Link           string `json:"link"`
	LinkAsAccount  string `json:"link_as_account"`
	Signature      string `json:"signature"`
	Work           string `json:"work"`
}

func (b *blockType) String() string {
	s, _ := json.Marshal(b)
	return string(s)
}

func blockCreate(previous, account, representative, balance, link, key, work string) (block createdBlock, err error) {
	const (
		keySize     = 32
		blockSize   = 32
		integerSize = 16
	)

	public, err := accountToPublicKey(account)
	if err != nil {
		return
	}

	prev, err := hex.DecodeString(previous)
	if err != nil {
		return
	}
	if len(prev) != blockSize {
		err = errors.New("invalid previous block size")
		return
	}

	repr, err := accountToPublicKey(representative)
	if err != nil {
		return
	}

	bal := make([]byte, integerSize)
	var balInt big.Int
	_, ok := balInt.SetString(balance, 10)
	if !ok {
		err = errors.New("invalid balance value")
		return
	}
	balInt.FillBytes(bal)

	var li []byte
	if strings.HasPrefix(link, "nano_") || strings.HasPrefix(link, "xrb_") {
		li, err = accountToPublicKey(link)
	} else {
		li, err = hex.DecodeString(link)
	}
	if err != nil {
		return
	}
	if len(li) != blockSize {
		err = errors.New("invalid link block size")
		return
	}

	private, err := hex.DecodeString(key)
	if err != nil {
		return
	}
	if len(private) != keySize {
		err = errors.New("invalid private key length")
		return
	}

	// https://docs.nano.org/integration-guides/the-basics/#self-signed-blocks
	hash := blockHash(public, prev, repr, bal, li)
	_, priv, _ := ed25519.GenerateKey(bytes.NewReader(private))
	signature := ed25519.Sign(priv, hash)

	blk := blockType{
		Type:           "state",
		Account:        account,
		Previous:       previous,
		Representative: representative,
		Balance:        balance,
		Link:           strings.ToUpper(hex.EncodeToString(li)),
		LinkAsAccount:  encodeAddress(li),
		Signature:      strings.ToUpper(hex.EncodeToString(signature)),
		Work:           work,
	}

	return createdBlock{
		Block: blk,
		Hash:  strings.ToUpper(hex.EncodeToString(hash)),
	}, nil
}

func blockHash(account, previous, representative, balance, link []byte) []byte {
	const stateBlockType = 6
	preamble := make([]byte, 32)
	preamble[31] = stateBlockType
	digest, _ := blake2b.New(32, nil) // nolint:gomnd
	digest.Write(preamble)            // nolint:errcheck
	digest.Write(account)             // nolint:errcheck
	digest.Write(previous)            // nolint:errcheck
	digest.Write(representative)      // nolint:errcheck
	digest.Write(balance)             // nolint:errcheck
	digest.Write(link)                // nolint:errcheck
	return digest.Sum(nil)
}

func accountToPublicKey(address string) (public []byte, err error) {
	// A valid nano address is 64 bytes long
	// First 5 are simply a hard-coded string nano_ for ease of use
	// The following 52 characters form the address, and the final
	// 8 are a checksum.
	switch {
	case address[:5] == "nano_":
		address = address[5:]
	case address[:4] == "xrb_":
		address = address[4:]
	default:
		err = errors.New("invalid address format")
		return
	}
	const addressLength = 60
	if len(address) != addressLength {
		err = errors.New("invalid address format")
		return
	}
	// The nano address string is 260bits which doesn't fall on a
	// byte boundary. pad with zeros to 280bits.
	// (zeros are encoded as 1 in nano's 32bit alphabet)
	keyB32nano := "1111" + address[0:52]
	inputChecksum := address[52:]

	public, err = addressEncoding.DecodeString(keyB32nano)
	if err != nil {
		return nil, err
	}
	// strip off upper 24 bits (3 bytes). 20 padding was added by us,
	// 4 is unused as account is 256 bits.
	public = public[3:]

	// nano checksum is calculated by hashing the key and reversing the bytes
	valid := addressEncoding.EncodeToString(calculateChecksum(public)) == inputChecksum
	if !valid {
		err = errors.New("invalid address checksum")
		return
	}
	return
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
