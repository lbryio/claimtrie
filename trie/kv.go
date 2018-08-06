package trie

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// Value defines value for the MerkleTrie.
type Value interface {
	Hash() *chainhash.Hash
}

// KeyValue ...
type KeyValue interface {
	Get(key []byte) Value
}
