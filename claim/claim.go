package claim

import (
	"bytes"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

type (
	// Amount defines the amount in LBC.
	Amount int64

	// Height defines the height of a block.
	Height int32
)

// New returns a Claim (or Support) initialized with specified op and amt.
func New(op OutPoint, amt Amount) *Claim {
	return &Claim{OutPoint: op, Amt: amt}
}

// Claim defines a structure of a Claim (or Support).
type Claim struct {
	OutPoint OutPoint
	ID       ID
	Amt      Amount
	Accepted Height
	Value    []byte

	EffAmt   Amount
	ActiveAt Height
}

func (c *Claim) setOutPoint(op OutPoint) *Claim { c.OutPoint = op; return c }
func (c *Claim) setID(id ID) *Claim             { c.ID = id; return c }
func (c *Claim) setAmt(amt Amount) *Claim       { c.Amt = amt; return c }
func (c *Claim) setAccepted(ht Height) *Claim   { c.Accepted = ht; return c }
func (c *Claim) setActiveAt(ht Height) *Claim   { c.ActiveAt = ht; return c }
func (c *Claim) setValue(val []byte) *Claim     { c.Value = val; return c }
func (c *Claim) String() string                 { return claimToString(c) }

func (c *Claim) expireAt() Height {
	if c.Accepted+paramOriginalClaimExpirationTime > paramExtendedClaimExpirationForkHeight {
		return c.Accepted + paramExtendedClaimExpirationTime
	}
	return c.Accepted + paramOriginalClaimExpirationTime
}

// IsActiveAt ...
func IsActiveAt(c *Claim, ht Height) bool {
	return c != nil && c.ActiveAt <= ht && c.expireAt() > ht
}

func equal(a, b *Claim) bool {
	if a != nil && b != nil {
		return a.OutPoint == b.OutPoint
	}
	return a == nil && b == nil
}

// OutPoint tracks previous transaction outputs.
type OutPoint struct {
	wire.OutPoint
}

// NewOutPoint returns a new outpoint with the provided hash and index.
func NewOutPoint(hash *chainhash.Hash, index uint32) *OutPoint {
	return &OutPoint{
		*wire.NewOutPoint(hash, index),
	}
}

func outPointLess(a, b OutPoint) bool {
	switch cmp := bytes.Compare(a.Hash[:], b.Hash[:]); {
	case cmp > 0:
		return true
	case cmp < 0:
		return false
	default:
		return a.Index < b.Index
	}
}
