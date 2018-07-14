package claim

import (
	"crypto/sha256"
	"encoding/binary"
	"strconv"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

func calNodeHash(op wire.OutPoint, tookover Height) chainhash.Hash {
	txHash := chainhash.DoubleHashH(op.Hash[:])

	nOut := []byte(strconv.Itoa(int(op.Index)))
	nOutHash := chainhash.DoubleHashH(nOut)

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(tookover))
	heightHash := chainhash.DoubleHashH(buf)

	h := make([]byte, 0, sha256.Size*3)
	h = append(h, txHash[:]...)
	h = append(h, nOutHash[:]...)
	h = append(h, heightHash[:]...)

	return chainhash.DoubleHashH(h)
}
