package claimtrie

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
	"github.com/lbryio/claimtrie/memento"
)

// Amount ...
type Amount int64

// Height ...
type Height int64

// Node ...
type Node struct {
	memento.Memento

	height     Height
	bestClaims map[Height]*Claim
	claims     map[wire.OutPoint]*Claim
	supports   map[wire.OutPoint]*Support
}

// NewNode ...
func NewNode() *Node {
	return &Node{
		Memento:    memento.Memento{},
		bestClaims: map[Height]*Claim{0: nil},
		claims:     map[wire.OutPoint]*Claim{},
		supports:   map[wire.OutPoint]*Support{},
	}
}

// BestClaim ...
func (n *Node) BestClaim() *Claim {
	var latest Height
	for k := range n.bestClaims {
		if k > latest {
			latest = k
		}
	}
	return n.bestClaims[latest]
}

// Tookover ...
func (n *Node) Tookover() Height {
	var latest Height
	for k := range n.bestClaims {
		if k > latest {
			latest = k
		}
	}
	return latest
}

// IncrementBlock ...
func (n *Node) IncrementBlock(h Height) error {
	if h < 0 {
		return ErrInvalidHeight
	}
	for i := Height(0); i < h; i++ {
		n.height++
		n.processBlock()
		n.Commit()
	}
	return nil
}

// DecrementBlock ...
func (n *Node) DecrementBlock(h Height) error {
	if h < 0 {
		return ErrInvalidHeight
	}
	for i := Height(0); i < h; i++ {
		n.height--
		n.Rollback()
	}
	return nil
}

func (n *Node) addClaim(op wire.OutPoint, amt Amount) (*Claim, error) {
	c := &Claim{
		op:       op,
		id:       NewClaimID(op),
		amt:      amt,
		accepted: n.height + 1,
		activeAt: n.height + 1,
	}
	if n.BestClaim() != nil {
		c.activeAt = calActiveHeight(c.accepted, c.accepted, n.Tookover())
	}

	n.Execute(cmdAddClaim{node: n, claim: c})
	return c, nil
}

func (n *Node) removeClaim(op wire.OutPoint) error {
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

func (n *Node) addSupport(op wire.OutPoint, amt Amount, supported ClaimID) (*Support, error) {
	s := &Support{
		op:          op,
		amt:         amt,
		supportedID: supported,
		accepted:    n.height + 1,
		activeAt:    n.height + 1,
	}
	if n.BestClaim() == nil || n.BestClaim().op != op {
		s.activeAt = calActiveHeight(s.accepted, s.accepted, n.Tookover())
	}

	for _, c := range n.claims {
		if c.id != supported {
			continue
		}
		n.Execute(cmdAddSupport{node: n, support: s})
		return s, nil
	}

	// Is supporting an non-existing Claim aceepted?
	return nil, ErrNotFound
}

func (n *Node) removeSupport(op wire.OutPoint) error {
	s, ok := n.supports[op]
	if !ok {
		return ErrNotFound
	}
	n.Execute(cmdRemoveSupport{node: n, support: s})
	return nil
}

func (n *Node) updateEffectiveAmounts() {
	for _, c := range n.claims {
		c.effAmt = c.amt
		if c.activeAt > n.height {
			c.effAmt = 0
			continue
		}
		for _, s := range n.supports {
			if s.activeAt > n.height || s.supportedID != c.id {
				continue
			}
			c.effAmt += s.amt
		}
	}
}

func (n *Node) updateActiveHeights() {
	for _, v := range n.claims {
		if old, new := v.activeAt, calActiveHeight(v.accepted, n.height, n.height); old != new {
			n.Execute(cmdUpdateClaimActiveHeight{claim: v, old: old, new: new})
		}
	}
	for _, v := range n.supports {
		if old, new := v.activeAt, calActiveHeight(v.accepted, n.height, n.height); old != new {
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

func (n *Node) findNextUpdateHeights() Height {
	next := Height(math.MaxInt64)
	for _, v := range n.claims {
		if v.activeAt > n.height && v.activeAt < next {
			next = v.activeAt
		}
	}
	for _, v := range n.supports {
		if v.activeAt > n.height && v.activeAt < next {
			next = v.activeAt
		}
	}
	if next == Height(math.MaxInt64) {
		return n.height
	}
	return next
}

// Hash calculates the Hash value based on the OutPoint and at which height it tookover.
func (n *Node) Hash() chainhash.Hash {
	if n.BestClaim() == nil {
		return chainhash.Hash{}
	}
	return calNodeHash(n.BestClaim().op, n.Tookover())
}

// MarshalJSON customizes JSON marshaling of the Node.
func (n *Node) MarshalJSON() ([]byte, error) {
	c := make([]*Claim, 0, len(n.claims))
	for _, v := range n.claims {
		c = append(c, v)
	}
	s := make([]*Support, 0, len(n.supports))
	for _, v := range n.supports {
		s = append(s, v)
	}
	return json.Marshal(&struct {
		Height    Height
		Hash      string
		Tookover  Height
		BestClaim *Claim
		Claims    []*Claim
		Supports  []*Support
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
			fmt.Fprintf(w, "{%d, %d}, ", k, v.op.Index)
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

// func (n *Node) clone() *Node {
// 	clone := NewNode()

// 	// shallow copy of value fields.
// 	*clone = *n

// 	// deep copy of reference and pointer fields.
// 	clone.claims = map[wire.OutPoint]*Claim{}
// 	for k, v := range n.claims {
// 		clone.claims[k] = v
// 	}
// 	clone.supports = map[wire.OutPoint]*Support{}
// 	for k, v := range n.supports {
// 		clone.supports[k] = v
// 	}
// 	return clone
// }

func findCandiadte(n *Node) *Claim {
	n.updateEffectiveAmounts()
	var candidate *Claim
	for _, v := range n.claims {
		if v.activeAt > n.height {
			continue
		}
		if candidate == nil || v.effAmt > candidate.effAmt {
			candidate = v
		}
	}
	return candidate
}

func calNodeHash(op wire.OutPoint, tookover Height) chainhash.Hash {
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

var proportionalDelayFactor = Height(32)

func calActiveHeight(accepted, curr, tookover Height) Height {
	delay := (curr - tookover) / proportionalDelayFactor
	if delay > 4032 {
		delay = 4032
	}
	return accepted + delay
}
