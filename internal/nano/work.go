package nano

import (
	"encoding/binary"
	"encoding/hex"
	"hash"
	"runtime"

	"github.com/cenkalti/log"
	"golang.org/x/crypto/blake2b"
)

var (
	workThresholdForSend uint64 = 0xfffffff800000000
	workThresholdForRecv uint64 = 0xfffffe0000000000
)

func GenerateWork(hash string, forSend bool) (string, error) {
	b, err := hex.DecodeString(hash)
	if err != nil {
		return "", err
	}
	const hashSize = 8
	digest, err := blake2b.New(hashSize, nil)
	if err != nil {
		return "", err
	}
	var workThreshold uint64
	if forSend {
		workThreshold = workThresholdForSend
	} else {
		workThreshold = workThresholdForRecv
	}
	var nonce uint64
	log.Debug("starting work")
	for ; !validateWork(digest, b, nonce, workThreshold); nonce++ {
		if nonce%1000 == 0 {
			runtime.Gosched()
		}
	}
	log.Debug("work finished")
	work := make([]byte, 8)
	binary.BigEndian.PutUint64(work, nonce)
	return hex.EncodeToString(work), nil
}

func validateWork(digest hash.Hash, block []byte, work uint64, workThreshold uint64) bool {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, work)

	digest.Reset()
	_, _ = digest.Write(b)
	_, _ = digest.Write(block)

	sum := digest.Sum(nil)
	return binary.LittleEndian.Uint64(sum) >= workThreshold
}
