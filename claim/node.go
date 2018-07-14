package claim

import (
	"math"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"

	"github.com/lbryio/claimtrie/memento"
)

// Node ...
type Node struct {
	mem        memento.Memento
	height     Height
	bestClaims map[Height]*Claim

	claims   claims
	supports supports

	updateNext bool
}

// NewNode returns a new Node.
func NewNode() *Node {
	return &Node{
		mem:        memento.Memento{},
		bestClaims: map[Height]*Claim{0: nil},
		claims:     claims{},
		supports:   supports{},
	}
}

// Height returns the current height.
func (n *Node) Height() Height {
	return n.height
}

// BestClaim returns the best claim at the current height.
func (n *Node) BestClaim() *Claim {
	c, _ := bestClaim(n.height, n.bestClaims)
	return c
}

// Tookover returns the height at which current best claim took over.
func (n *Node) Tookover() Height {
	_, since := bestClaim(n.height, n.bestClaims)
	return since
}

// AdjustTo increments or decrements current height until it reaches the specific height.
func (n *Node) AdjustTo(h Height) error {
	for n.height < h {
		n.Increment()
	}
	for n.height > h {
		n.Decrement()
	}
	return nil
}

// Increment ...
// Increment also clears out the undone stack if it wasn't empty.
func (n *Node) Increment() error {
	n.height++
	n.processBlock()
	n.mem.Commit()
	return nil
}

// Decrement ...
func (n *Node) Decrement() error {
	n.height--
	n.mem.Undo()
	return nil
}

// Redo ...
func (n *Node) Redo() error {
	if err := n.mem.Redo(); err != nil {
		return err
	}
	n.height++
	return nil
}

// RollbackExecuted ...
func (n *Node) RollbackExecuted() error {
	n.mem.RollbackExecuted()
	return nil
}

// AddClaim ...
func (n *Node) AddClaim(c *Claim) error {
	if _, ok := n.claims.has(c.OutPoint); ok {
		return ErrDuplicate
	}
	next := n.height + 1
	c.SetAccepted(next).SetActiveAt(next)
	if n.BestClaim() != nil {
		c.SetActiveAt(calActiveHeight(next, next, n.Tookover()))
	}

	n.mem.Execute(cmdAddClaim{node: n, claim: c})
	return nil
}

// RemoveClaim ...
func (n *Node) RemoveClaim(op wire.OutPoint) error {
	c, ok := n.claims.has(op)
	if !ok {
		return ErrNotFound
	}
	n.mem.Execute(cmdRemoveClaim{node: n, claim: c})

	if *n.BestClaim() != *c {
		return nil
	}
	n.mem.Execute(updateNodeBestClaim{node: n, height: n.Tookover(), old: c, new: nil})
	updateActiveHeights(n.height, n.claims, n.supports, &n.mem)
	n.updateNext = true
	return nil
}

// AddSupport ...
func (n *Node) AddSupport(s *Support) error {
	next := n.height + 1
	s.SetAccepted(next).SetActiveAt(next)
	if n.BestClaim() == nil || n.BestClaim().ID != s.ClaimID {
		s.SetActiveAt(calActiveHeight(next, next, n.Tookover()))
	}

	for _, c := range n.claims {
		if c.ID != s.ClaimID {
			continue
		}
		n.mem.Execute(cmdAddSupport{node: n, support: s})
		return nil
	}

	// Is supporting an non-existing Claim aceepted?
	return ErrNotFound
}

// RemoveSupport ...
func (n *Node) RemoveSupport(op wire.OutPoint) error {
	s, ok := n.supports.has(op)
	if !ok {
		return ErrNotFound
	}
	n.supports = n.supports.remove(op)
	n.mem.Execute(cmdRemoveSupport{node: n, support: s})
	return nil
}

// FindNextUpdateHeight returns the smallest height in the future that the the state of the node might change.
// If no such height exists, the current height of the node is returned.
func (n *Node) FindNextUpdateHeight() Height {
	if n.updateNext {
		n.updateNext = false
		return n.height + 1
	}

	return findNextUpdateHeight(n.height, n.claims, n.supports)
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
	return nodeToJSON(n)
}

// String implements Stringer interface.
func (n *Node) String() string {
	return nodeToString(n)
}

func (n *Node) processBlock() {
	for {
		if c := n.BestClaim(); c != nil && !isActive(n.height, c.Accepted, c.ActiveAt) {
			n.mem.Execute(updateNodeBestClaim{node: n, height: n.height, old: n.bestClaims[n.height], new: nil})
			updateActiveHeights(n.height, n.claims, n.supports, &n.mem)
		}
		updateEffectiveAmounts(n.height, n.claims, n.supports)
		candidate := findCandiadte(n.height, n.claims)
		if n.BestClaim() == candidate {
			return
		}
		n.mem.Execute(updateNodeBestClaim{node: n, height: n.height, old: n.bestClaims[n.height], new: candidate})
		updateActiveHeights(n.height, n.claims, n.supports, &n.mem)
	}
}

func updateEffectiveAmounts(h Height, claims claims, supports supports) {
	for _, c := range claims {
		c.EffAmt = 0
		if !isActive(h, c.Accepted, c.ActiveAt) {
			continue
		}
		c.EffAmt = c.Amt
		for _, s := range supports {
			if !isActive(h, s.Accepted, s.ActiveAt) || s.ClaimID != c.ID {
				continue
			}
			c.EffAmt += s.Amt
		}
	}
}

func updateActiveHeights(h Height, claims claims, supports supports, mem *memento.Memento) {
	for _, v := range claims {
		if old, new := v.ActiveAt, calActiveHeight(v.Accepted, h, h); old != new {
			mem.Execute(cmdUpdateClaimActiveHeight{claim: v, old: old, new: new})
		}
	}
	for _, v := range supports {
		if old, new := v.ActiveAt, calActiveHeight(v.Accepted, h, h); old != new {
			mem.Execute(cmdUpdateSupportActiveHeight{support: v, old: old, new: new})
		}
	}
}

// bestClaim returns the best claim at specified height and since when it took over.
func bestClaim(at Height, bestClaims map[Height]*Claim) (*Claim, Height) {
	var latest Height
	for k := range bestClaims {
		if k > at {
			continue
		}
		if k > latest {
			latest = k
		}
	}
	return bestClaims[latest], latest
}

func findNextUpdateHeight(h Height, claims claims, supports supports) Height {
	next := Height(math.MaxInt64)
	for _, v := range claims {
		if v.ActiveAt > h && v.ActiveAt < next {
			next = v.ActiveAt
		}
	}
	for _, v := range supports {
		if v.ActiveAt > h && v.ActiveAt < next {
			next = v.ActiveAt
		}
	}
	if next == Height(math.MaxInt64) {
		return h
	}
	return next
}

func findCandiadte(h Height, claims claims) *Claim {
	var candidate *Claim
	for _, v := range claims {
		switch {
		case v.ActiveAt > h:
			continue
		case candidate == nil:
			candidate = v
		case v.EffAmt > candidate.EffAmt:
			candidate = v
		case v.EffAmt == candidate.EffAmt && v.seq < candidate.seq:
			candidate = v
		}
	}
	return candidate
}

func isActive(h, accepted, activeAt Height) bool {
	if activeAt > h {
		// Accepted, but not active yet.
		return false
	}
	if h >= paramExtendedClaimExpirationForkHeight && accepted+paramExtendedClaimExpirationTime <= h {
		// Expired on post-HF1807 duration
		return false
	}
	if h < paramExtendedClaimExpirationForkHeight && accepted+paramOriginalClaimExpirationTime <= h {
		// Expired on pre-HF1807 duration
		return false
	}
	return true
}

func calActiveHeight(Accepted, curr, tookover Height) Height {
	delay := (curr - tookover) / paramActiveDelayFactor
	if delay > paramMaxActiveDelay {
		delay = paramMaxActiveDelay
	}
	return Accepted + delay
}
