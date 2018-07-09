package claim

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"

	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

// NewID ...
func NewID(op wire.OutPoint) ID {
	w := bytes.NewBuffer(op.Hash[:])
	if err := binary.Write(w, binary.BigEndian, op.Index); err != nil {
		panic(err)
	}
	var id ID
	copy(id[:], btcutil.Hash160(w.Bytes()))
	return id
}

// NewIDFromString ...
func NewIDFromString(s string) (ID, error) {
	b, err := hex.DecodeString(s)
	var id ID
	copy(id[:], b)
	return id, err
}

// ID ...
type ID [20]byte

func (id ID) String() string {
	return hex.EncodeToString(id[:])
}
