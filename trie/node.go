package trie

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

type node struct {
	hash     *chainhash.Hash
	links    [256]*node
	hasValue bool
}

func newNode() *node {
	return &node{}
}

// nbuf decodes the on-disk format of a node, which has the following form:
//   ch(1B) hash(32B)
//   ...
//   ch(1B) hash(32B)
//   vhash(32B)
type nbuf []byte

func (nb nbuf) entries() int {
	return len(nb) / 33
}

func (nb nbuf) entry(i int) (byte, *chainhash.Hash) {
	h := chainhash.Hash{}
	copy(h[:], nb[33*i+1:])
	return nb[33*i], &h
}

func (nb nbuf) hasValue() bool {
	return len(nb)%33 == 32
}
