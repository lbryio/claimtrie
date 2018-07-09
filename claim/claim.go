package claim

import (
	"encoding/json"
	"fmt"

	"github.com/btcsuite/btcd/wire"
)

// Amount ...
type Amount int64

// Height ...
type Height int64

// Seq is a strictly increasing sequence number determine relative order between Claims and Supports.
type Seq uint64

// Claim ...
type Claim struct {
	OutPoint wire.OutPoint
	ID       ID
	Amt      Amount
	EffAmt   Amount
	Accepted Height
	ActiveAt Height

	// TODO: Get rid of this. Implement ordered map in upper layer.
	Seq Seq
}

func (c *Claim) String() string {
	return fmt.Sprintf("C-%-68s amt: %-3d  effamt: %-3d  accepted: %-3d  active: %-3d  id: %s",
		c.OutPoint, c.Amt, c.EffAmt, c.Accepted, c.ActiveAt, c.ID)
}

// MarshalJSON customizes the representation of JSON.
func (c *Claim) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		OutPoint  string
		ID        string
		Amount    Amount
		EffAmount Amount
		Accepted  Height
		ActiveAt  Height
	}{
		OutPoint:  c.OutPoint.String(),
		ID:        c.ID.String(),
		Amount:    c.Amt,
		EffAmount: c.EffAmt,
		Accepted:  c.Accepted,
		ActiveAt:  c.ActiveAt,
	})
}

// Support ...
type Support struct {
	OutPoint wire.OutPoint
	ClaimID  ID
	Amt      Amount
	Accepted Height
	ActiveAt Height
	Seq      Seq
}

func (s *Support) String() string {
	return fmt.Sprintf("S-%-68s amt: %-3d               accepted: %-3d  active: %-3d  id: %s",
		s.OutPoint, s.Amt, s.Accepted, s.ActiveAt, s.ClaimID)
}

// MarshalJSON customizes the representation of JSON.
func (s *Support) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		OutPoint string
		ClaimID  string
		Amount   Amount
		Accepted Height
		ActiveAt Height
	}{
		OutPoint: s.OutPoint.String(),
		ClaimID:  s.ClaimID.String(),
		Amount:   s.Amt,
		Accepted: s.Accepted,
		ActiveAt: s.ActiveAt,
	})
}
