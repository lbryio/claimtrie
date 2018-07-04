package claimtrie

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"

	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

// NewClaimID ...
func NewClaimID(op wire.OutPoint) ClaimID {
	w := bytes.NewBuffer(op.Hash[:])
	if err := binary.Write(w, binary.BigEndian, op.Index); err != nil {
		panic(err)
	}
	var id ClaimID
	copy(id[:], btcutil.Hash160(w.Bytes()))
	return id
}

// NewClaimIDFromString ...
func NewClaimIDFromString(s string) (ClaimID, error) {
	b, err := hex.DecodeString(s)
	var id ClaimID
	copy(id[:], b)
	return id, err
}

// ClaimID ...
type ClaimID [20]byte

func (id ClaimID) String() string {
	return hex.EncodeToString(id[:])
}

func calActiveHeight(accepted, curr, tookover Height) Height {
	factor := Height(32)
	delay := (curr - tookover) / factor
	if delay > 4032 {
		delay = 4032
	}
	return accepted + delay
}
