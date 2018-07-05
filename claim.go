package claimtrie

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/btcsuite/btcd/wire"
)

var dbg bool

// Claim ...
type Claim struct {
	op       wire.OutPoint
	id       ClaimID
	amt      Amount
	effAmt   Amount
	accepted Height
	activeAt Height
}

func (c *Claim) String() string {
	if dbg {
		w := bytes.NewBuffer(nil)
		fmt.Fprintf(w, "C%-3d amt: %2d, effamt: %v, accepted: %2d, active: %2d, id: %s", c.op.Index, c.amt, c.effAmt, c.accepted, c.activeAt, c.id)
		return w.String()
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		fmt.Printf("can't marshal, err :%s", err)
	}
	return string(b)
}

// Support ...
type Support struct {
	op       wire.OutPoint
	amt      Amount
	accepted Height
	activeAt Height

	supportedID ClaimID
}

// MarshalJSON customizes the representation of JSON.
func (c *Claim) MarshalJSON() ([]byte, error) {
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

// MarshalJSON ...
func (s *Support) MarshalJSON() ([]byte, error) {
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

func (s *Support) String() string {
	if dbg {
		w := bytes.NewBuffer(nil)
		fmt.Fprintf(w, "S%-3d amt: %2d,            accepted: %2d, active: %2d, id: %s", s.op.Index, s.amt, s.accepted, s.activeAt, s.supportedID)
		return w.String()
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		fmt.Printf("can't marshal, err :%s", err)
	}
	return string(b)
}
