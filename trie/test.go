package trie

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// Internal utility functions to facilitate the tests.

type strValue string

func (s strValue) Hash() *chainhash.Hash {
	h := chainhash.DoubleHashH([]byte(s))
	return &h
}

func dump(prefix Key, value Value) error {
	if value == nil {
		fmt.Printf("[%-8s]\n", prefix)
		return nil
	}
	fmt.Printf("[%-8s] %v\n", prefix, value)
	return nil
}

func buildNode(n *node, pairs []pair) *node {
	for _, val := range pairs {
		update(n, Key(val.k), val.v)
	}
	return n
}

func buildTrie(mt *MerkleTrie, pairs []pair) *MerkleTrie {
	for _, val := range pairs {
		mt.Update(Key(val.k), val.v)
	}
	return mt
}

func buildMap(m map[string]Value, pairs []pair) map[string]Value {
	for _, p := range pairs {
		if p.v == nil {
			delete(m, p.k)
		} else {
			m[p.k] = p.v
		}
	}
	return m
}

func newMap() map[string]Value {
	return map[string]Value{}
}

type pair struct {
	k string
	v Value
}

func pairs1() []pair {
	return []pair{
		{"alex", strValue("lion")},
		{"al", strValue("tiger")},
		{"tess", strValue("dolphin")},
		{"bob", strValue("pig")},
		{"ted", strValue("dog")},
		{"teddy", strValue("bear")},
		{"al", nil},
		{"alex", nil},
		{"bob", strValue("cat")},
	}
}

func prunedNode() *node {
	n := newNode(nil)
	n.links['b'] = newNode(nil)
	n.links['b'].links['o'] = newNode(nil)
	n.links['b'].links['o'].links['b'] = newNode(strValue("cat"))
	n.links['t'] = newNode(nil)
	n.links['t'].links['e'] = newNode(nil)
	n.links['t'].links['e'].links['d'] = newNode(strValue("dog"))
	n.links['t'].links['e'].links['d'].links['d'] = newNode(nil)
	n.links['t'].links['e'].links['d'].links['d'].links['y'] = newNode(strValue("bear"))
	n.links['t'].links['e'].links['s'] = newNode(nil)
	n.links['t'].links['e'].links['s'].links['s'] = newNode(strValue("dolphin"))
	return n
}

func unprunedNode() *node {
	n := newNode(nil)
	n.links['a'] = newNode(nil)
	n.links['a'].links['l'] = newNode(nil)
	n.links['a'].links['l'].links['e'] = newNode(nil)
	n.links['a'].links['l'].links['e'].links['x'] = newNode(nil)
	n.links['b'] = newNode(nil)
	n.links['b'].links['o'] = newNode(nil)
	n.links['b'].links['o'].links['b'] = newNode(strValue("cat"))
	n.links['t'] = newNode(nil)
	n.links['t'].links['e'] = newNode(nil)
	n.links['t'].links['e'].links['d'] = newNode(strValue("dog"))
	n.links['t'].links['e'].links['d'].links['d'] = newNode(nil)
	n.links['t'].links['e'].links['d'].links['d'].links['y'] = newNode(strValue("bear"))
	n.links['t'].links['e'].links['s'] = newNode(nil)
	n.links['t'].links['e'].links['s'].links['s'] = newNode(strValue("dolphin"))
	return n
}
