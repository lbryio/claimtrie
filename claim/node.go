package claim

import (
	"fmt"
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

	claims   list
	supports list

	// refer to updateClaim.
	removed list

	records []*Cmd
}

// NewNode returns a new Node.
func NewNode(name string) *Node {
	return &Node{name: name}
}

// Height returns the current height.
func (n *Node) Height() Height {
	return n.height
}

// BestClaim returns the best claim at the current height.
func (n *Node) BestClaim() *Claim {
	return n.best
}

// AddClaim adds a claim to the node.
func (n *Node) AddClaim(op OutPoint, amt Amount) error {
	return n.execute(n.record(CmdAddClaim, op, amt, ID{}))
}

// SpendClaim spends a claim in the node.
func (n *Node) SpendClaim(op OutPoint) error {
	return n.execute(n.record(CmdSpendClaim, op, 0, ID{}))
}

// UpdateClaim updates a claim in the node.
func (n *Node) UpdateClaim(op OutPoint, amt Amount, id ID) error {
	return n.execute(n.record(CmdUpdateClaim, op, amt, id))
}

// AddSupport adds a support in the node.
func (n *Node) AddSupport(op OutPoint, amt Amount, id ID) error {
	return n.execute(n.record(CmdAddSupport, op, amt, id))
}

// SpendSupport spends a spport in the node.
func (n *Node) SpendSupport(op OutPoint) error {
	return n.execute(n.record(CmdSpendSupport, op, 0, ID{}))
}

func (n *Node) addClaim(op OutPoint, amt Amount) error {
	if find(byOP(op), n.claims, n.supports) != nil {
		return ErrDuplicate
	}

	accepted := n.height + 1
	c := New(op, amt).setID(NewID(op)).setAccepted(accepted)
	c.setActiveAt(accepted + calDelay(accepted, n.tookover))
	if !isActiveAt(n.best, accepted) {
		c.setActiveAt(accepted)
		n.best, n.tookover = c, accepted
	}
	n.claims = append(n.claims, c)
	return nil
}

func (n *Node) spendClaim(op OutPoint) error {
	var c *Claim
	if n.claims, c = remove(n.claims, byOP(op)); c == nil {
		return ErrNotFound
	}
	n.removed = append(n.removed, c)
	return nil
}

// A claim update is composed of two separate commands (2 & 3 below).
//
//   (1) blk  500: Add Claim (opA, amtA, NewID(opA)
//     ...
//   (2) blk 1000: Spend Claim (opA, idA)
//   (3) blk 1000: Update Claim (opB, amtB, idA)
//
// For each block, all the spent claims are kept in n.removed until committed.
// The paired (spend, update) commands has to happen in the same trasaction.
func (n *Node) updateClaim(op OutPoint, amt Amount, id ID) error {
	if find(byOP(op), n.claims, n.supports) != nil {
		return ErrDuplicate
	}
	var c *Claim
	if n.removed, c = remove(n.removed, byID(id)); c == nil {
		return errors.Wrapf(ErrNotFound, "remove(n.removed, byID(%s)", id)
	}

	accepted := n.height + 1
	c.setOutPoint(op).setAmt(amt).setAccepted(accepted)
	c.setActiveAt(accepted + calDelay(accepted, n.tookover))
	if n.best != nil && n.best.ID == id {
		c.setActiveAt(n.tookover)
	}
	n.claims = append(n.claims, c)
	return nil
}

func (n *Node) addSupport(op OutPoint, amt Amount, id ID) error {
	if find(byOP(op), n.claims, n.supports) != nil {
		return ErrDuplicate
	}
	// Accepted by rules. No effects on bidding result though.
	// It may be spent later.
	if find(byID(id), n.claims, n.removed) == nil {
		fmt.Printf("INFO: can't find suooported claim ID: %s\n", id)
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

func (n *Node) spendSupport(op OutPoint) error {
	var s *Claim
	if n.supports, s = remove(n.supports, byOP(op)); s != nil {
		return nil
	}
	return ErrNotFound
}

// NextUpdate returns the height at which pending updates should happen.
// When no pending updates exist, current height is returned.
func (n *Node) NextUpdate() Height {
	next := Height(math.MaxInt32)
	min := func(l list) Height {
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

func updateEffectiveAmounts(h Height, claims, supports list) {
	for _, c := range claims {
		c.EffAmt = 0
		if !isActiveAt(c, h) {
			continue
		}
		c.EffAmt = c.Amt
		for _, s := range supports {
			if !isActiveAt(s, h) || s.ID != c.ID {
				continue
			}
			c.EffAmt += s.Amt
		}
	}
}

func updateActiveHeights(n *Node, lists ...list) {
	for _, l := range lists {
		for _, v := range l {
			v.ActiveAt = v.Accepted + calDelay(n.height, n.tookover)
		}
	}
}

func findCandiadte(h Height, claims list) *Claim {
	var c *Claim
	for _, v := range claims {
		switch {
		case !isActiveAt(v, h):
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
