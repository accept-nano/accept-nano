package nano

import (
	"encoding/binary"
	"encoding/hex"
	"hash"
	"runtime"

	"github.com/cenkalti/log"
	"github.com/golang/crypto/blake2b"
)

var workThreshold = uint64(0xffffffc000000000)

func GenerateWork(hash string) (string, error) {
	b, err := hex.DecodeString(hash)
	if err != nil {
		return "", err
	}
	digest, err := blake2b.New(8, nil)
	if err != nil {
		return "", err
	}
	var nonce uint64
	log.Debug("starting work")
	for ; !validateWork(digest, b, nonce); nonce++ {
		if nonce%1000 == 0 {
			runtime.Gosched()
		}
	}
	log.Debug("work finished")
	work := make([]byte, 8)
	binary.BigEndian.PutUint64(work, nonce)
	return hex.EncodeToString(work), nil
}

func validateWork(digest hash.Hash, block []byte, work uint64) bool {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, work)

	digest.Reset()
	_, _ = digest.Write(b)
	_, _ = digest.Write(block)

	sum := digest.Sum(nil)
	return binary.LittleEndian.Uint64(sum) >= workThreshold
}
