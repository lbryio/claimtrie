package claim

import (
	"math"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/pkg/errors"
)

// Node ...
type Node struct {
	name string

	height Height

	best     *Claim
	tookover Height

	claims   List
	supports List

	// refer to updateClaim.
	removed List
}

// NewNode returns a new Node.
func NewNode(name string) *Node {
	return &Node{name: name}
}

// Name returns the Name where the Node blongs.
func (n *Node) Name() string {
	return n.name
}

// Height returns the current height.
func (n *Node) Height() Height {
	return n.height
}

// BestClaim returns the best claim at the current height.
func (n *Node) BestClaim() *Claim {
	return n.best
}

// Tookover returns the the height at when the current BestClaim tookover.
func (n *Node) Tookover() Height {
	return n.tookover
}

// Claims returns the claims at the current height.
func (n *Node) Claims() List {
	return n.claims
}

// Supports returns the supports at the current height.
func (n *Node) Supports() List {
	return n.supports
}

// AddClaim adds a Claim to the Node.
func (n *Node) AddClaim(op OutPoint, amt Amount, val []byte) error {
	if Find(ByOP(op), n.claims, n.supports) != nil {
		return ErrDuplicate
	}
	accepted := n.height + 1
	c := New(op, amt).setID(NewID(op)).setAccepted(accepted).setValue(val)
	c.setActiveAt(accepted + calDelay(accepted, n.tookover))
	if !IsActiveAt(n.best, accepted) {
		c.setActiveAt(accepted)
		n.best, n.tookover = c, accepted
	}
	n.claims = append(n.claims, c)
	return nil
}

// SpendClaim spends a Claim in the Node.
func (n *Node) SpendClaim(op OutPoint) error {
	var c *Claim
	if n.claims, c = Remove(n.claims, ByOP(op)); c == nil {
		return ErrNotFound
	}
	n.removed = append(n.removed, c)
	return nil
}

// UpdateClaim updates a Claim in the Node.
// A claim update is composed of two separate commands (2 & 3 below).
//
//   (1) blk  500: Add Claim (opA, amtA, NewID(opA)
//     ...
//   (2) blk 1000: Spend Claim (opA, idA)
//   (3) blk 1000: Update Claim (opB, amtB, idA)
//
// For each block, all the spent claims are kept in n.removed until committed.
// The paired (spend, update) commands has to happen in the same trasaction.
func (n *Node) UpdateClaim(op OutPoint, amt Amount, id ID, val []byte) error {
	if Find(ByOP(op), n.claims, n.supports) != nil {
		return ErrDuplicate
	}
	var c *Claim
	if n.removed, c = Remove(n.removed, ByID(id)); c == nil {
		return errors.Wrapf(ErrNotFound, "remove(n.removed, byID(%s)", id)
	}

	accepted := n.height + 1
	c.setOutPoint(op).setAmt(amt).setAccepted(accepted).setValue(val)
	c.setActiveAt(accepted + calDelay(accepted, n.tookover))
	if n.best != nil && n.best.ID == id {
		c.setActiveAt(n.tookover)
	}
	n.claims = append(n.claims, c)
	return nil
}

// AddSupport adds a Support to the Node.
func (n *Node) AddSupport(op OutPoint, amt Amount, id ID) error {
	if Find(ByOP(op), n.claims, n.supports) != nil {
		return ErrDuplicate
	}
	// Accepted by rules. No effects on bidding result though.
	// It may be spent later.
	if Find(ByID(id), n.claims, n.removed) == nil {
		// fmt.Printf("INFO: can't find suooported claim ID: %s for %s\n", id, n.name)
	}

	accepted := n.height + 1
	s := New(op, amt).setID(id).setAccepted(accepted)
	s.setActiveAt(accepted + calDelay(accepted, n.tookover))
	if n.best != nil && n.best.ID == id {
		s.setActiveAt(accepted)
	}
	n.supports = append(n.supports, s)
	return nil
}

// SpendSupport spends a support in the Node.
func (n *Node) SpendSupport(op OutPoint) error {
	var s *Claim
	if n.supports, s = Remove(n.supports, ByOP(op)); s != nil {
		return nil
	}
	return ErrNotFound
}

// AdjustTo increments current height until it reaches the specific height.
func (n *Node) AdjustTo(ht Height) *Node {
	if ht <= n.height {
		return n
	}
	for n.height < ht {
		n.height++
		n.bid()
		next := n.NextUpdate()
		if next > ht || next == n.height {
			n.height = ht
			break
		}
		n.height = next
		n.bid()
	}
	n.bid()
	return n
}

// NextUpdate returns the height at which pending updates should happen.
// When no pending updates exist, current height is returned.
func (n *Node) NextUpdate() Height {
	next := Height(math.MaxInt32)
	min := func(l List) Height {
		for _, v := range l {
			exp := v.expireAt()
			if n.height >= exp {
				continue
			}
			if v.ActiveAt > n.height && v.ActiveAt < next {
				next = v.ActiveAt
			}
			if exp > n.height && exp < next {
				next = exp
			}
		}
		return next
	}
	min(n.claims)
	min(n.supports)
	if next == Height(math.MaxInt32) {
		next = n.height
	}
	return next
}

func (n *Node) bid() {
	for {
		if n.best == nil || n.height >= n.best.expireAt() {
			n.best, n.tookover = nil, n.height
			updateActiveHeights(n, n.claims, n.supports)
		}
		updateEffectiveAmounts(n.height, n.claims, n.supports)
		c := findCandiadte(n.height, n.claims)
		if equal(n.best, c) {
			break
		}
		n.best, n.tookover = c, n.height
		updateActiveHeights(n, n.claims, n.supports)
	}
	n.removed = nil
}

func updateEffectiveAmounts(ht Height, claims, supports List) {
	for _, c := range claims {
		c.EffAmt = 0
		if !IsActiveAt(c, ht) {
			continue
		}
		c.EffAmt = c.Amt
		for _, s := range supports {
			if !IsActiveAt(s, ht) || s.ID != c.ID {
				continue
			}
			c.EffAmt += s.Amt
		}
	}
}

func updateActiveHeights(n *Node, lists ...List) {
	for _, l := range lists {
		for _, v := range l {
			if v.ActiveAt < n.height {
				continue
			}
			v.ActiveAt = v.Accepted + calDelay(n.height, n.tookover)
			if v.ActiveAt < n.height {
				v.ActiveAt = n.height
			}
		}
	}
}

func findCandiadte(ht Height, claims List) *Claim {
	var c *Claim
	for _, v := range claims {
		switch {
		case !IsActiveAt(v, ht):
			continue
		case c == nil:
			c = v
		case v.EffAmt > c.EffAmt:
			c = v
		case v.EffAmt < c.EffAmt:
			continue
		case v.Accepted < c.Accepted:
			c = v
		case v.Accepted > c.Accepted:
			continue
		case outPointLess(c.OutPoint, v.OutPoint):
			c = v
		}
	}
	return c
}

func calDelay(curr, tookover Height) Height {
	delay := (curr - tookover) / paramActiveDelayFactor
	if delay > paramMaxActiveDelay {
		return paramMaxActiveDelay
	}
	return delay
}

// Hash calculates the Hash value based on the OutPoint and when it tookover.
func (n *Node) Hash() *chainhash.Hash {
	if n.best == nil {
		return nil
	}
	return calNodeHash(n.best.OutPoint, n.tookover)
}

func (n *Node) String() string {
	return nodeToString(n)
}
