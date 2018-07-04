package claimtrie

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"strconv"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

type node struct {
	tookover  Height
	bestClaim *claim

	claims   map[string]*claim
	supports map[string]*support
}

func newNode() *node {
	return &node{
		claims:   map[string]*claim{},
		supports: map[string]*support{},
	}
}

func (n *node) addClaim(c *claim) error {
	if _, ok := n.claims[c.op.String()]; ok {
		return ErrDuplicate
	}
	c.activeAt = calActiveHeight(c.accepted, c.accepted, n.tookover)
	n.claims[c.op.String()] = c
	return nil
}

func (n *node) removeClaim(op wire.OutPoint) error {
	c, ok := n.claims[op.String()]
	if !ok {
		return ErrNotFound
	}
	delete(n.claims, op.String())
	if n.bestClaim == c {
		n.bestClaim = nil
	}
	for _, v := range n.supports {
		if c.id == v.supportedID {
			v.supportedClaim = nil
			return nil
		}
	}
	return nil
}

func (n *node) addSupport(s *support) error {
	if _, ok := n.supports[s.op.String()]; ok {
		return ErrDuplicate
	}
	for _, v := range n.claims {
		if v.id == s.supportedID {
			s.activeAt = calActiveHeight(s.accepted, s.accepted, n.tookover)
			s.supportedClaim = v
			n.supports[s.op.String()] = s
			return nil
		}
	}
	return ErrNotFound
}

func (n *node) removeSupport(op wire.OutPoint) error {
	if _, ok := n.supports[op.String()]; !ok {
		return ErrNotFound
	}
	delete(n.supports, op.String())
	return nil
}

// Hash calculates the Hash value based on the OutPoint and at which height it tookover.
func (n *node) Hash() chainhash.Hash {
	return calNodeHash(n.bestClaim.op, n.tookover)
}

// MarshalJSON customizes JSON marshaling of the Node.
func (n *node) MarshalJSON() ([]byte, error) {
	c := make([]*claim, 0, len(n.claims))
	for _, v := range n.claims {
		c = append(c, v)
	}
	s := make([]*support, 0, len(n.supports))
	for _, v := range n.supports {
		s = append(s, v)
	}
	return json.Marshal(&struct {
		Hash      string
		Tookover  Height
		BestClaim *claim
		Claims    []*claim
		Supports  []*support
	}{
		Hash:      n.Hash().String(),
		Tookover:  n.tookover,
		BestClaim: n.bestClaim,
		Claims:    c,
		Supports:  s,
	})
}

// String implements Stringer interface.
func (n *node) String() string {
	b, err := json.MarshalIndent(n, "", "  ")
	if err != nil {
		panic("can't marshal Node")
	}
	return string(b)
}

func (n *node) updateEffectiveAmounts(curr Height) {
	for _, v := range n.claims {
		v.effAmt = v.amt
	}
	for _, v := range n.supports {
		if v.supportedClaim == n.bestClaim || v.activeAt <= curr {
			v.supportedClaim.effAmt += v.amt
		}
	}
}

func (n *node) updateBestClaim(curr Height) {
	findCandiadte := func() *claim {
		candidate := n.bestClaim
		for _, v := range n.claims {
			if v.activeAt > curr {
				// Accepted claim, but noy activated yet.
				continue
			}
			if candidate == nil || v.effAmt > candidate.effAmt {
				candidate = v
			}
		}
		return candidate
	}
	takeover := func(candidate *claim) {
		n.bestClaim = candidate
		n.tookover = curr
		for _, v := range n.claims {
			v.activeAt = calActiveHeight(v.accepted, curr, curr)
		}
	}
	for {
		n.updateEffectiveAmounts(curr)
		candidate := findCandiadte()
		if n.bestClaim == nil {
			takeover(candidate)
			return

		}
		if n.bestClaim == candidate {
			return
		}
		takeover(candidate)
	}
}

func (n *node) clone() *node {
	clone := newNode()

	// shallow copy of value fields.
	*clone = *n

	// deep copy of reference and pointer fields.
	clone.claims = map[string]*claim{}
	for k, v := range n.claims {
		clone.claims[k] = v
	}
	clone.supports = map[string]*support{}
	for k, v := range n.supports {
		clone.supports[k] = v
	}
	return clone
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
