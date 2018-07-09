package trie

import (
	"sync"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

var (
	// EmptyTrieHash represent the Merkle Hash of an empty MerkleTrie.
	EmptyTrieHash = *newHashFromStr("0000000000000000000000000000000000000000000000000000000000000001")
)

// Key defines the key type of the MerkleTrie.
type Key []byte

// Value implements value for the MerkleTrie.
type Value interface {
	Hash() chainhash.Hash
}

// MerkleTrie implements a 256-way prefix tree, which takes Key as key and any value that implements the Value interface.
type MerkleTrie struct {
	mu   *sync.RWMutex
	root *node
}

// New returns a MerkleTrie.
func New() *MerkleTrie {
	return &MerkleTrie{
		mu:   &sync.RWMutex{},
		root: newNode(nil),
	}
}

// Get returns the Value associated with the key, or nil with error.
// Most common error is ErrMissing, which indicates no Value is associated with the key.
// However, there could be other errors propagated from I/O layer (TBD).
func (t *MerkleTrie) Get(key Key) (Value, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	n := t.root
	for _, k := range key {
		if n.links[k] == nil {
			// Path does not exist.
			return nil, ErrKeyNotFound
		}
		n = n.links[k]
	}
	if n.value == nil {
		// Path exists, but no Value is associated.
		// This happens when the key had been deleted, but the MerkleTrie has not nullified yet.
		return nil, ErrKeyNotFound
	}
	return n.value, nil
}

// Update updates the MerkleTrie with specified key-value pair.
// Setting Value to nil deletes the Value, if exists, associated to the key.
func (t *MerkleTrie) Update(key Key, val Value) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	update(t.root, key, val)
	return nil
}

// Prune removes nodes that do not reach to any value node.
func (t *MerkleTrie) Prune() {
	t.mu.Lock()
	defer t.mu.Unlock()

	prune(t.root)
}

// Size returns the number of values.
func (t *MerkleTrie) Size() int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	size := 0 // captured in the closure.
	fn := func(prefix Key, v Value) error {
		if v != nil {
			size++
		}
		return nil
	}
	traverse(t.root, Key{}, fn)
	return size
}

// Visit implements callback function invoked when the Value is visited.
// During the traversal, if a non-nil error is returned, the traversal ends early.
type Visit func(prefix Key, val Value) error

// Traverse visits every Value in the MerkleTrie and returns error defined by specified Visit function.
// update indicates if the visit function modify the state of MerkleTrie.
func (t *MerkleTrie) Traverse(visit Visit, update, valueOnly bool) error {
	if update {
		t.mu.Lock()
		defer t.mu.Unlock()
	} else {
		t.mu.RLock()
		defer t.mu.RUnlock()
	}
	fn := func(prefix Key, value Value) error {
		if !valueOnly || value != nil {
			return visit(prefix, value)
		}
		return nil
	}
	return traverse(t.root, Key{}, fn)
}

// MerkleHash calculates the Merkle Hash of the MerkleTrie.
// If the MerkleTrie is empty, EmptyTrieHash is returned.
func (t *MerkleTrie) MerkleHash() chainhash.Hash {
	if merkle(t.root) == nil {
		return EmptyTrieHash
	}
	return *t.root.hash
}

func newHashFromStr(s string) *chainhash.Hash {
	h, _ := chainhash.NewHashFromStr(s)
	return h
}
