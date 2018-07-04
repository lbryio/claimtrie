package claimtrie

import (
	"encoding/json"
	"fmt"

	"github.com/btcsuite/btcd/wire"
)

// newClaim ...
func newClaim(op wire.OutPoint, amt Amount, accepted Height) *claim {
	return &claim{
		op:       op,
		id:       NewClaimID(op),
		amt:      amt,
		accepted: accepted,
	}
}

type claim struct {
	op       wire.OutPoint
	id       ClaimID
	amt      Amount
	effAmt   Amount
	accepted Height
	activeAt Height
}

// MarshalJSON customizes the representation of JSON.
func (c *claim) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		OutPoint        string
		ClaimID         string
		Amount          Amount
		EffectiveAmount Amount
		Accepted        Height
		ActiveAt        Height
	}{
		OutPoint:        c.op.String(),
		ClaimID:         c.id.String(),
		Amount:          c.amt,
		EffectiveAmount: c.effAmt,
		Accepted:        c.accepted,
		ActiveAt:        c.activeAt,
	})
}

func (c *claim) String() string {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		fmt.Printf("can't marshal, err :%s", err)
	}
	return string(b)
}

func newSupport(op wire.OutPoint, amt Amount, accepted Height, supported ClaimID) *support {
	return &support{
		op:          op,
		amt:         amt,
		accepted:    accepted,
		supportedID: supported,
	}
}

type support struct {
	op       wire.OutPoint
	amt      Amount
	accepted Height
	activeAt Height

	supportedID    ClaimID
	supportedClaim *claim
}

// MarshalJSON ...
func (s *support) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		OutPoint         string
		SupportedClaimID string
		Amount           Amount
		Accepted         Height
		ActiveAt         Height
	}{
		OutPoint:         s.op.String(),
		SupportedClaimID: s.supportedID.String(),
		Amount:           s.amt,
		Accepted:         s.accepted,
		ActiveAt:         s.activeAt,
	})
}

func (s *support) String() string {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		fmt.Printf("can't marshal, err :%s", err)
	}
	return string(b)
}
