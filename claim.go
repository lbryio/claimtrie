package claimtrie

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

// NewClaim ...
func NewClaim(op wire.OutPoint, amt Amount, accepted Height) *Claim {
	return &Claim{
		op:       op,
		id:       NewClaimID(op),
		amt:      amt,
		accepted: accepted,
	}
}

// Claim ...
type Claim struct {
	op       wire.OutPoint
	id       ClaimID
	amt      Amount
	effAmt   Amount
	accepted Height
}

// ActivateAt ...
func (c *Claim) ActivateAt(best *Claim, curr, tookover Height) Height {
	if best == nil || best == c {
		return c.accepted
	}
	return calActiveHeight(c.accepted, curr, tookover)
}

// MarshalJSON customizes the representation of JSON.
func (c *Claim) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		OutPoint        string
		ClaimID         string
		Amount          Amount
		EffectiveAmount Amount
		Accepted        Height
	}{
		OutPoint:        c.op.String(),
		ClaimID:         c.id.String(),
		Amount:          c.amt,
		EffectiveAmount: c.effAmt,
		Accepted:        c.accepted,
	})
}

func (c *Claim) String() string {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		fmt.Printf("can't marshal, err :%s", err)
	}
	return string(b)
}

// NewSupport ...
func NewSupport(op wire.OutPoint, amt Amount, accepted Height, supported ClaimID) *Support {
	return &Support{
		op:          op,
		amt:         amt,
		accepted:    accepted,
		supportedID: supported,
	}
}

// Support ...
type Support struct {
	op       wire.OutPoint
	amt      Amount
	accepted Height

	supportedID    ClaimID
	supportedClaim *Claim
}

// ActivateAt ...
func (s *Support) ActivateAt(best *Claim, curr, tookover Height) Height {
	if best == nil || best == s.supportedClaim {
		return s.accepted
	}
	return calActiveHeight(s.accepted, curr, tookover)
}

// MarshalJSON ...
func (s *Support) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		OutPoint         string
		SupportedClaimID string
		Amount           Amount
		Accepted         Height
	}{
		OutPoint:         s.op.String(),
		SupportedClaimID: s.supportedID.String(),
		Amount:           s.amt,
		Accepted:         s.accepted,
	})
}

func (s *Support) String() string {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		fmt.Printf("can't marshal, err :%s", err)
	}
	return string(b)
}

// NewClaimID ...
func NewClaimID(p wire.OutPoint) ClaimID {
	w := bytes.NewBuffer(p.Hash[:])
	if err := binary.Write(w, binary.BigEndian, p.Index); err != nil {
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
