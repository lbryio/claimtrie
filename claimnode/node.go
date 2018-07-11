package claimnode

import (
	"crypto/sha256"
	"encoding/binary"
	"math"
	"strconv"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"

	"github.com/lbryio/claimtrie/claim"
	"github.com/lbryio/claimtrie/memento"
)

// Node ...
type Node struct {
	mem        memento.Memento
	height     claim.Height
	bestClaims map[claim.Height]*claim.Claim

	// To ensure the Claims and Supports are totally ordered, we assign a
	// strictly increasing seq to each Claim or Support added to the node.
	seq      claim.Seq
	claims   map[wire.OutPoint]*claim.Claim
	supports map[wire.OutPoint]*claim.Support

	updateNext bool
}

// NewNode returns a new Node.
func NewNode() *Node {
	return &Node{
		mem:        memento.Memento{},
		bestClaims: map[claim.Height]*claim.Claim{0: nil},
		claims:     map[wire.OutPoint]*claim.Claim{},
		supports:   map[wire.OutPoint]*claim.Support{},
	}
}

// Height returns the current height.
func (n *Node) Height() claim.Height {
	return n.height
}

// BestClaim returns the best claim at the current height.
func (n *Node) BestClaim() *claim.Claim {
	c, _ := BestClaimAt(n, n.height)
	return c
}

// Tookover returns the height at which current best claim tookover.
func (n *Node) Tookover() claim.Height {
	_, since := BestClaimAt(n, n.height)
	return since
}

// AdjustTo increments or decrements current height until it reaches the specific height.
func (n *Node) AdjustTo(h claim.Height) error {
	for n.height < h {
		n.height++
		n.processBlock()
		n.mem.Commit()
	}
	for n.height > h {
		n.height--
		n.mem.Rollback()
	}
	return nil
}

// Reset ...
func (n *Node) Reset() error {
	n.mem.RollbackUncommited()
	return nil
}

// AddClaim ...
func (n *Node) AddClaim(op wire.OutPoint, amt claim.Amount) (*claim.Claim, error) {
	n.seq++
	c := &claim.Claim{
		OutPoint: op,
		ID:       claim.NewID(op),
		Amt:      amt,
		Accepted: n.height + 1,
		ActiveAt: n.height + 1,
		Seq:      n.seq,
	}
	if n.BestClaim() != nil {
		c.ActiveAt = calActiveHeight(c.Accepted, c.Accepted, n.Tookover())
	}

	n.mem.Execute(cmdAddClaim{node: n, claim: c})
	return c, nil
}

// RemoveClaim ...
func (n *Node) RemoveClaim(op wire.OutPoint) error {
	c, ok := n.claims[op]
	if !ok {
		return ErrNotFound
	}
	n.mem.Execute(cmdRemoveClaim{node: n, claim: c})

	if n.BestClaim() != c {
		return nil
	}
	n.mem.Execute(updateNodeBestClaim{node: n, height: n.Tookover(), old: c, new: nil})
	n.updateActiveHeights()
	n.updateNext = true
	return nil
}

// AddSupport ...
func (n *Node) AddSupport(op wire.OutPoint, amt claim.Amount, supported claim.ID) (*claim.Support, error) {
	n.seq++
	s := &claim.Support{
		OutPoint: op,
		Amt:      amt,
		ClaimID:  supported,
		Accepted: n.height + 1,
		ActiveAt: n.height + 1,
		Seq:      n.seq,
	}
	if n.BestClaim() == nil || n.BestClaim().ID != supported {
		s.ActiveAt = calActiveHeight(s.Accepted, s.Accepted, n.Tookover())
	}

	for _, c := range n.claims {
		if c.ID != supported {
			continue
		}
		n.mem.Execute(cmdAddSupport{node: n, support: s})
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
	n.mem.Execute(cmdRemoveSupport{node: n, support: s})
	return nil
}

// FindNextUpdateHeight returns the smallest height in the future that the the state of the node might change.
// If no such height exists, the current height of the node is returned.
func (n *Node) FindNextUpdateHeight() claim.Height {
	if n.updateNext {
		n.updateNext = false
		return n.height + 1
	}

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
	return toJSON(n)
}

// String implements Stringer interface.
func (n *Node) String() string {
	return toString(n)
}

func (n *Node) updateEffectiveAmounts() {
	for _, c := range n.claims {
		c.EffAmt = c.Amt
		if c.ActiveAt > n.height {
			c.EffAmt = 0
			continue
		}
		for _, s := range n.supports {
			if s.ActiveAt > n.height || s.ClaimID != c.ID {
				continue
			}
			c.EffAmt += s.Amt
		}
	}
}

func (n *Node) updateActiveHeights() {
	for _, v := range n.claims {
		if old, new := v.ActiveAt, calActiveHeight(v.Accepted, n.height, n.height); old != new {
			n.mem.Execute(cmdUpdateClaimActiveHeight{claim: v, old: old, new: new})
		}
	}
	for _, v := range n.supports {
		if old, new := v.ActiveAt, calActiveHeight(v.Accepted, n.height, n.height); old != new {
			n.mem.Execute(cmdUpdateSupportActiveHeight{support: v, old: old, new: new})
		}
	}
}

func (n *Node) processBlock() {
	for {
		n.updateEffectiveAmounts()
		candidate := findCandiadte(n)
		if n.BestClaim() == candidate {
			return
		}
		n.mem.Execute(updateNodeBestClaim{node: n, height: n.height, old: n.bestClaims[n.height], new: candidate})
		n.updateActiveHeights()
	}
}

func findCandiadte(n *Node) *claim.Claim {
	var candidate *claim.Claim
	for _, v := range n.claims {
		switch {
		case v.ActiveAt > n.height:
			continue
		case candidate == nil:
			candidate = v
		case v.EffAmt > candidate.EffAmt:
			candidate = v
		case v.EffAmt == candidate.EffAmt && v.Seq < candidate.Seq:
			candidate = v
		}
	}
	return candidate
}

// BestClaimAt returns the BestClaim at specified Height along with the height when the claim tookover.
func BestClaimAt(n *Node, at claim.Height) (best *claim.Claim, since claim.Height) {
	var latest claim.Height
	for k := range n.bestClaims {
		if k > at {
			continue
		}
		if k > latest {
			latest = k
		}
	}
	return n.bestClaims[latest], latest
}

// clone copies (deeply) the contents (except memento) of src to dst.
func clone(dst, src *Node) {
	dst.height = src.height
	for k, v := range src.bestClaims {
		if v == nil {
			dst.bestClaims[k] = nil
			continue
		}
		dup := *v
		dst.bestClaims[k] = &dup
	}
	for k, v := range src.claims {
		dup := *v
		dst.claims[k] = &dup
	}
	for k, v := range src.supports {
		dup := *v
		dst.supports[k] = &dup
	}
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
