package claim

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"

	"github.com/btcsuite/btcutil"
)

// ErrInvalidID is returned when the ID does not conform to the format.
var ErrInvalidID = errors.New("ID must be a 20-character hexadecimal string")

// NewID returns a Claim ID caclculated from Ripemd160(Sha256(OUTPOINT).
func NewID(op OutPoint) ID {
	w := bytes.NewBuffer(op.Hash[:])
	if err := binary.Write(w, binary.BigEndian, op.Index); err != nil {
		panic(err)
	}
	var id ID
	copy(id[:], btcutil.Hash160(w.Bytes()))
	return id
}

// NewIDFromString returns a Claim ID from a string.
func NewIDFromString(s string) (ID, error) {
	var id ID
	if len(s) != 40 {
		return id, ErrInvalidID
	}
	_, err := hex.Decode(id[:], []byte(s))
	for i, j := 0, len(id)-1; i < j; i, j = i+1, j-1 {
		id[i], id[j] = id[j], id[i]
	}
	return id, err
}

// ID represents a Claim's ID.
type ID [20]byte

func (id ID) String() string {
	for i, j := 0, len(id)-1; i < j; i, j = i+1, j-1 {
		id[i], id[j] = id[j], id[i]
	}
	return hex.EncodeToString(id[:])
}
