package claim

import (
	"sync/atomic"

	"github.com/btcsuite/btcd/wire"
)

type (
	// Amount ...
	Amount int64

	// Height ...
	Height int64
)

// seq is a strictly increasing sequence number determine relative order between Claims and Supports.
var seq uint64

// New ...
func New(op wire.OutPoint, amt Amount) *Claim {
	return &Claim{OutPoint: op, ID: NewID(op), Amt: amt, seq: atomic.AddUint64(&seq, 1)}
}

// Claim ...
type Claim struct {
	OutPoint wire.OutPoint
	ID       ID
	Amt      Amount
	EffAmt   Amount
	Accepted Height
	ActiveAt Height

	seq uint64
}

// SetOutPoint ...
func (c *Claim) SetOutPoint(op wire.OutPoint) *Claim {
	c.OutPoint = op
	c.ID = NewID(op)
	return c
}

// SetAmt ...
func (c *Claim) SetAmt(amt Amount) *Claim {
	c.Amt = amt
	return c
}

// SetAccepted ...
func (c *Claim) SetAccepted(h Height) *Claim {
	c.Accepted = h
	return c
}

// SetActiveAt ...
func (c *Claim) SetActiveAt(h Height) *Claim {
	c.ActiveAt = h
	return c
}

// String ...
func (c *Claim) String() string {
	return claimToString(c)
}

// MarshalJSON customizes the representation of JSON.
func (c *Claim) MarshalJSON() ([]byte, error) { return claimToJSON(c) }

// NewSupport ...
func NewSupport(op wire.OutPoint, amt Amount, claimID ID) *Support {
	return &Support{OutPoint: op, Amt: amt, ClaimID: claimID, seq: atomic.AddUint64(&seq, 1)}
}

// Support ...
type Support struct {
	OutPoint wire.OutPoint
	ClaimID  ID
	Amt      Amount
	Accepted Height
	ActiveAt Height

	seq uint64
}

// SetOutPoint ...
func (s *Support) SetOutPoint(op wire.OutPoint) *Support {
	s.OutPoint = op
	return s
}

// SetAmt ...
func (s *Support) SetAmt(amt Amount) *Support {
	s.Amt = amt
	return s
}

// SetClaimID ...
func (s *Support) SetClaimID(id ID) *Support {
	s.ClaimID = id
	return s
}

// SetAccepted ...
func (s *Support) SetAccepted(h Height) *Support {
	s.Accepted = h
	return s
}

// SetActiveAt ...
func (s *Support) SetActiveAt(h Height) *Support {
	s.ActiveAt = h
	return s
}

// String ...
func (s *Support) String() string {
	return supportToString(s)
}

// MarshalJSON customizes the representation of JSON.
func (s *Support) MarshalJSON() ([]byte, error) {
	return supportToJSON(s)
}

type claims []*Claim

func (cc claims) remove(op wire.OutPoint) claims {
	for i, v := range cc {
		if v.OutPoint != op {
			continue
		}
		cc[i] = cc[len(cc)-1]
		cc[len(cc)-1] = nil
		return cc[:len(cc)-1]
	}
	return cc
}

func (cc claims) has(op wire.OutPoint) (*Claim, bool) {
	for _, v := range cc {
		if v.OutPoint == op {
			return v, true
		}
	}
	return nil, false
}

type supports []*Support

func (ss supports) remove(op wire.OutPoint) supports {
	for i, v := range ss {
		if v.OutPoint != op {
			continue
		}
		ss[i] = ss[len(ss)-1]
		ss[len(ss)-1] = nil
		return ss[:len(ss)-1]
	}
	return ss
}

func (ss supports) has(op wire.OutPoint) (*Support, bool) {
	for _, v := range ss {
		if v.OutPoint == op {
			return v, true
		}
	}
	return nil, false
}
