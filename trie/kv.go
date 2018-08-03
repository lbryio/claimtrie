package trie

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// Key defines the key type of the MerkleTrie.
type Key []byte

// Value defines value for the MerkleTrie.
type Value interface {
	Hash() *chainhash.Hash
}

// KeyValue ...
type KeyValue interface {
	Get(k Key) (Value, error)
	Set(k Key, v Value)
}
