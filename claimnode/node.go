package claimnode

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"

	"github.com/lbryio/claimtrie/claim"
	"github.com/lbryio/claimtrie/memento"
)

var dbg bool

// Node ...
type Node struct {
	memento.Memento

	height     claim.Height
	bestClaims map[claim.Height]*claim.Claim
	claims     map[wire.OutPoint]*claim.Claim
	supports   map[wire.OutPoint]*claim.Support
}

// NewNode ...
func NewNode() *Node {
	return &Node{
		Memento:    memento.Memento{},
		bestClaims: map[claim.Height]*claim.Claim{0: nil},
		claims:     map[wire.OutPoint]*claim.Claim{},
		supports:   map[wire.OutPoint]*claim.Support{},
	}
}

// Height ...
func (n *Node) Height() claim.Height {
	return n.height
}

// BestClaim ...
func (n *Node) BestClaim() *claim.Claim {
	var latest claim.Height
	for k := range n.bestClaims {
		if k > latest {
			latest = k
		}
	}
	return n.bestClaims[latest]
}

// Tookover ...
func (n *Node) Tookover() claim.Height {
	var latest claim.Height
	for k := range n.bestClaims {
		if k > latest {
			latest = k
		}
	}
	return latest
}

// IncrementBlock ...
func (n *Node) IncrementBlock(h claim.Height) error {
	if h < 0 {
		return ErrInvalidHeight
	}
	for i := claim.Height(0); i < h; i++ {
		n.height++
		n.processBlock()
		n.Commit()
	}
	return nil
}

// DecrementBlock ...
func (n *Node) DecrementBlock(h claim.Height) error {
	if h < 0 {
		return ErrInvalidHeight
	}
	for i := claim.Height(0); i < h; i++ {
		n.height--
		n.Rollback()
	}
	return nil
}

// AddClaim ...
func (n *Node) AddClaim(op wire.OutPoint, amt claim.Amount) (*claim.Claim, error) {
	c := &claim.Claim{
		OutPoint: op,
		ID:       claim.NewID(op),
		Amt:      amt,
		Accepted: n.height + 1,
		ActiveAt: n.height + 1,
	}
	if n.BestClaim() != nil {
		c.ActiveAt = calActiveHeight(c.Accepted, c.Accepted, n.Tookover())
	}

	n.Execute(cmdAddClaim{node: n, claim: c})
	return c, nil
}

// RemoveClaim ...
func (n *Node) RemoveClaim(op wire.OutPoint) error {
	c, ok := n.claims[op]
	if !ok {
		return ErrNotFound
	}
	n.Execute(cmdRemoveClaim{node: n, claim: c})

	if n.BestClaim() != c {
		return nil
	}
	n.Execute(updateNodeBestClaim{node: n, height: n.Tookover(), old: c, new: nil})
	n.updateActiveHeights()
	return nil
}

// AddSupport ...
func (n *Node) AddSupport(op wire.OutPoint, amt claim.Amount, supported claim.ID) (*claim.Support, error) {
	s := &claim.Support{
		OutPoint:    op,
		Amt:         amt,
		SupportedID: supported,
		Accepted:    n.height + 1,
		ActiveAt:    n.height + 1,
	}
	if n.BestClaim() == nil || n.BestClaim().OutPoint != op {
		s.ActiveAt = calActiveHeight(s.Accepted, s.Accepted, n.Tookover())
	}

	for _, c := range n.claims {
		if c.ID != supported {
			continue
		}
		n.Execute(cmdAddSupport{node: n, support: s})
		return s, nil
	}

	// Is supporting an non-existing Claim aceepted?
	return nil, ErrNotFound
}

// RemoveSupport ...
func (n *Node) RemoveSupport(op wire.OutPoint) error {
	s, ok := n.supports[op]
	if !ok {
		return ErrNotFound
	}
	n.Execute(cmdRemoveSupport{node: n, support: s})
	return nil
}

// FindNextUpdateHeights ...
func (n *Node) FindNextUpdateHeights() claim.Height {
	next := claim.Height(math.MaxInt64)
	for _, v := range n.claims {
		if v.ActiveAt > n.height && v.ActiveAt < next {
			next = v.ActiveAt
		}
	}
	for _, v := range n.supports {
		if v.ActiveAt > n.height && v.ActiveAt < next {
			next = v.ActiveAt
		}
	}
	if next == claim.Height(math.MaxInt64) {
		return n.height
	}
	return next
}

// Hash calculates the Hash value based on the OutPoint and at which height it tookover.
func (n *Node) Hash() chainhash.Hash {
	if n.BestClaim() == nil {
		return chainhash.Hash{}
	}
	return calNodeHash(n.BestClaim().OutPoint, n.Tookover())
}

// MarshalJSON customizes JSON marshaling of the Node.
func (n *Node) MarshalJSON() ([]byte, error) {
	c := make([]*claim.Claim, 0, len(n.claims))
	for _, v := range n.claims {
		c = append(c, v)
	}
	s := make([]*claim.Support, 0, len(n.supports))
	for _, v := range n.supports {
		s = append(s, v)
	}
	return json.Marshal(&struct {
		Height    claim.Height
		Hash      string
		Tookover  claim.Height
		BestClaim *claim.Claim
		Claims    []*claim.Claim
		Supports  []*claim.Support
	}{
		Height:    n.height,
		Hash:      n.Hash().String(),
		Tookover:  n.Tookover(),
		BestClaim: n.BestClaim(),
		Claims:    c,
		Supports:  s,
	})
}

// String implements Stringer interface.
func (n *Node) String() string {
	if dbg {
		w := bytes.NewBuffer(nil)
		fmt.Fprintf(w, "H: %2d   BestClaims: ", n.height)
		for k, v := range n.bestClaims {
			if v == nil {
				fmt.Fprintf(w, "{%d, nil}, ", k)
				continue
			}
			fmt.Fprintf(w, "{%d, %d}, ", k, v.Index)
		}
		fmt.Fprintf(w, "\n")
		for _, v := range n.claims {
			fmt.Fprintf(w, "\n    %v", v)
			if v == n.BestClaim() {
				fmt.Fprintf(w, " <B> ")
			}
		}
		for _, v := range n.supports {
			fmt.Fprintf(w, "\n   %v", v)
		}
		fmt.Fprintf(w, "\n")
		return w.String()

	}
	b, err := json.MarshalIndent(n, "", "  ")
	if err != nil {
		panic("can't marshal Node")
	}
	return string(b)
}
func (n *Node) updateEffectiveAmounts() {
	for _, c := range n.claims {
		c.EffAmt = c.Amt
		if c.ActiveAt > n.height {
			c.EffAmt = 0
			continue
		}
		for _, s := range n.supports {
			if s.ActiveAt > n.height || s.SupportedID != c.ID {
				continue
			}
			c.EffAmt += s.Amt
		}
	}
}

func (n *Node) updateActiveHeights() {
	for _, v := range n.claims {
		if old, new := v.ActiveAt, calActiveHeight(v.Accepted, n.height, n.height); old != new {
			n.Execute(cmdUpdateClaimActiveHeight{claim: v, old: old, new: new})
		}
	}
	for _, v := range n.supports {
		if old, new := v.ActiveAt, calActiveHeight(v.Accepted, n.height, n.height); old != new {
			n.Execute(cmdUpdateSupportActiveHeight{support: v, old: old, new: new})
		}
	}
}
func (n *Node) processBlock() {
	for {
		candidate := findCandiadte(n)
		if n.BestClaim() == candidate {
			return
		}
		n.Execute(updateNodeBestClaim{node: n, height: n.height, old: n.bestClaims[n.height], new: candidate})
		n.updateActiveHeights()
	}
}

func findCandiadte(n *Node) *claim.Claim {
	n.updateEffectiveAmounts()
	var candidate *claim.Claim
	for _, v := range n.claims {
		if v.ActiveAt > n.height {
			continue
		}
		if candidate == nil || v.EffAmt > candidate.EffAmt {
			candidate = v
		}
	}
	return candidate
}

func calNodeHash(op wire.OutPoint, tookover claim.Height) chainhash.Hash {
	txHash := chainhash.DoubleHashH(op.Hash[:])

	nOut := []byte(strconv.Itoa(int(op.Index)))
	nOutHash := chainhash.DoubleHashH(nOut)

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(tookover))
	heightHash := chainhash.DoubleHashH(buf)

	h := make([]byte, 0, sha256.Size*3)
	h = append(h, txHash[:]...)
	h = append(h, nOutHash[:]...)
	h = append(h, heightHash[:]...)

	return chainhash.DoubleHashH(h)
}

var proportionalDelayFactor = claim.Height(32)

func calActiveHeight(Accepted, curr, tookover claim.Height) claim.Height {
	delay := (curr - tookover) / proportionalDelayFactor
	if delay > 4032 {
		delay = 4032
	}
	return Accepted + delay
}
