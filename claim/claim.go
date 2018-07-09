package claim

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/btcsuite/btcd/wire"
)

// Amount ...
type Amount int64

// Height ...
type Height int64

var dbg bool

// Claim ...
type Claim struct {
	wire.OutPoint

	ID       ID
	Amt      Amount
	EffAmt   Amount
	Accepted Height
	ActiveAt Height
}

func (c *Claim) String() string {
	if dbg {
		w := bytes.NewBuffer(nil)
		fmt.Fprintf(w, "C%-3d amt: %2d, effamt: %v, accepted: %2d, active: %2d, id: %s", c.Index, c.Amt, c.EffAmt, c.Accepted, c.ActiveAt, c.ID)
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
	wire.OutPoint

	Amt      Amount
	Accepted Height
	ActiveAt Height

	SupportedID ID
}

// MarshalJSON customizes the representation of JSON.
func (c *Claim) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		OutPoint        string
		ID              string
		Amount          Amount
		EffectiveAmount Amount
		Accepted        Height
		ActiveAt        Height
	}{
		OutPoint:        c.OutPoint.String(),
		ID:              c.ID.String(),
		Amount:          c.Amt,
		EffectiveAmount: c.EffAmt,
		Accepted:        c.Accepted,
		ActiveAt:        c.ActiveAt,
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
		OutPoint:         s.OutPoint.String(),
		SupportedClaimID: s.SupportedID.String(),
		Amount:           s.Amt,
		Accepted:         s.Accepted,
		ActiveAt:         s.ActiveAt,
	})
}

func (s *Support) String() string {
	if dbg {
		w := bytes.NewBuffer(nil)
		fmt.Fprintf(w, "S%-3d amt: %2d,            accepted: %2d, active: %2d, id: %s", s.Index, s.Amt, s.Accepted, s.ActiveAt, s.SupportedID)
		return w.String()
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		fmt.Printf("can't marshal, err :%s", err)
	}
	return string(b)
}
