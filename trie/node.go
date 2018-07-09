package trie

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

type node struct {
	hash  *chainhash.Hash
	links [256]*node
	value Value
}

func newNode(val Value) *node {
	return &node{links: [256]*node{}, value: val}
}

// We clear the Merkle Hash for every node along the path, including the root.
// Calculation of the hash happens much less frequently then updating to the MerkleTrie.
func update(n *node, key Key, val Value) {
	// Follow the path to reach the node.
	for _, k := range key {
		if n.links[k] == nil {
			// The path didn't exist yet. Build it.
			n.links[k] = newNode(nil)
		}
		n.hash = nil
		n = n.links[k]
	}

	n.value = val
	n.hash = nil
}

func prune(n *node) *node {
	if n == nil {
		return nil
	}
	var ret *node
	for i, v := range n.links {
		if n.links[i] = prune(v); n.links[i] != nil {
			ret = n
		}
	}
	if n.value != nil {
		ret = n
	}
	return ret
}

func traverse(n *node, prefix Key, visit Visit) error {
	if n == nil {
		return nil
	}
	for i, v := range n.links {
		if v == nil {
			continue
		}
		p := append(prefix, byte(i))
		if err := visit(p, v.value); err != nil {
			return err
		}
		if err := traverse(v, p, visit); err != nil {
			return err
		}
	}
	return nil
}

// merkle recursively caculates the Merkle Hash of a given node
// It works with both pruned or unpruned nodes.
func merkle(n *node) *chainhash.Hash {
	if n.hash != nil {
		return n.hash
	}
	buf := Key{}
	for i, v := range n.links {
		if v == nil {
			continue
		}
		if h := merkle(v); h != nil {
			buf = append(buf, byte(i))
			buf = append(buf, h[:]...)
		}
	}
	if n.value != nil {
		h := n.value.Hash()
		buf = append(buf, h[:]...)
	}

	if len(buf) != 0 {
		// At least one of the sub nodes has contributed a value hash.
		h := chainhash.DoubleHashH(buf)
		n.hash = &h
	}
	return n.hash
}
