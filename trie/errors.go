package trie

import "errors"

var (
	// ErrKeyNotFound is returned when the key doesn't exist in the MerkleTrie.
	ErrKeyNotFound = errors.New("key not found")
)
