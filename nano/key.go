package nano

import (
	"bytes"
	"encoding/base32"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"

	ed25519 "github.com/accept-nano/ed25519-blake2b"
	"golang.org/x/crypto/blake2b"
)

type Key struct {
	Private string `json:"private"`
	Public  string `json:"public"`
	Account string `json:"account"`
}

func (n *Node) DeterministicKey(seed string, index string) (*Key, error) {
	key, err := deterministicKey(seed, index)
	return &key, err
}

func deterministicKey(seed, index string) (key Key, err error) {
	seedBytes, err := hex.DecodeString(seed)
	if err != nil {
		return
	}
	const seedLen = 32
	if len(seedBytes) != seedLen {
		err = errors.New("invalid seed length")
		return
	}
	i, err := strconv.ParseUint(index, 10, 64)
	if err != nil {
		return
	}
	indexBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(indexBytes, uint32(i))
	digest, _ := blake2b.New(32, nil) // nolint:gomnd
	digest.Write(seedBytes)           // nolint:errcheck
	digest.Write(indexBytes)          // nolint:errcheck
	private := digest.Sum(nil)
	key.Private = strings.ToUpper(hex.EncodeToString(private))

	publicKey, _, _ := ed25519.GenerateKey(bytes.NewReader(private))
	public := []byte(publicKey)
	key.Public = strings.ToUpper(hex.EncodeToString(public))

	key.Account = encodeAddress(public)
	return
}

var addressEncoding = base32.NewEncoding("13456789abcdefghijkmnopqrstuwxyz")

func reverse(s []byte) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func calculateChecksum(pub []byte) []byte {
	hash, _ := blake2b.New(5, nil) // nolint:gomnd
	hash.Write(pub)                // nolint:errcheck
	sum := hash.Sum(nil)
	reverse(sum)
	return sum
}

func encodeAddress(pub []byte) string {
	padded := append([]byte{0, 0, 0}, pub...)
	address := addressEncoding.EncodeToString(padded)[4:]
	checksum := addressEncoding.EncodeToString(calculateChecksum(pub))
	return "nano_" + address + checksum
}
