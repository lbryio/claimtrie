package trie

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

var (
	// ErrResolve is returned when an error occured during resolve.
	ErrResolve = fmt.Errorf("can't resolve")
)
var (
	// EmptyTrieHash represents the Merkle Hash of an empty Trie.
	// "0000000000000000000000000000000000000000000000000000000000000001"
	EmptyTrieHash = &chainhash.Hash{1}
)

// Trie implements a 256-way prefix tree.
type Trie struct {
	kv KeyValue
	db *leveldb.DB

	root  *node
	bufs  *sync.Pool
	batch *leveldb.Batch
}

// New returns a Trie.
func New(kv KeyValue, db *leveldb.DB) *Trie {
	return &Trie{
		kv:   kv,
		db:   db,
		root: newNode(),
		bufs: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
}

// SetRoot drops all resolved nodes in the Trie, and set the root with specified hash.
func (t *Trie) SetRoot(h *chainhash.Hash) {
	t.root = newNode()
	t.root.hash = h
}

// Update updates the nodes along the path to the key.
// Each node is resolved or created with their Hash cleared.
func (t *Trie) Update(key Key) error {
	n := t.root
	for _, ch := range key {
		if err := t.resolve(n); err != nil {
			return ErrResolve
		}
		if n.links[ch] == nil {
			n.links[ch] = newNode()
		}
		n.hash = nil
		n = n.links[ch]
	}
	if err := t.resolve(n); err != nil {
		return ErrResolve
	}
	n.hasValue = true
	n.hash = nil
	return nil
}

func (t *Trie) resolve(n *node) error {
	if n.hash == nil {
		return nil
	}
	b, err := t.db.Get(n.hash[:], nil)
	if err == leveldb.ErrNotFound {
		return nil
	} else if err != nil {
		return errors.Wrapf(err, "db.Get(%s)", n.hash)
	}

	nb := nbuf(b)
	n.hasValue = nb.hasValue()
	for i := 0; i < nb.entries(); i++ {
		p, h := nb.entry(i)
		n.links[p] = newNode()
		n.links[p].hash = h
	}
	return nil
}

// Visit implements callback function invoked when the Value is visited.
// During the traversal, if a non-nil error is returned, the traversal ends early.
type Visit func(prefix Key, val Value) error

// Traverse implements preorder traversal visiting each Value node.
func (t *Trie) Traverse(visit Visit) error {
	var traverse func(prefix Key, n *node) error
	traverse = func(prefix Key, n *node) error {
		if n == nil {
			return nil
		}
		for ch, n := range n.links {
			if n == nil || !n.hasValue {
				continue
			}

			p := append(prefix, byte(ch))
			val, err := t.kv.Get(p)
			if err != nil {
				return errors.Wrapf(err, "kv.Get(%s)", p)
			}
			if err := visit(p, val); err != nil {
				return err
			}

			if err := traverse(p, n); err != nil {
				return err
			}
		}
		return nil
	}
	buf := make([]byte, 0, 4096)
	return traverse(buf, t.root)
}

// MerkleHash returns the Merkle Hash of the Trie.
// All nodes must have been resolved before calling this function.
func (t *Trie) MerkleHash() (*chainhash.Hash, error) {
	t.batch = &leveldb.Batch{}
	buf := make([]byte, 0, 4096)
	if err := t.merkle(buf, t.root); err != nil {
		return nil, err
	}
	if t.root.hash == nil {
		return EmptyTrieHash, nil
	}
	if t.db != nil && t.batch.Len() != 0 {
		if err := t.db.Write(t.batch, nil); err != nil {
			return nil, errors.Wrapf(err, "db.Write(t.batch, nil)")
		}
	}
	return t.root.hash, nil
}

// merkle recursively resolves the hashes of the node.
// All nodes must have been resolved before calling this function.
func (t *Trie) merkle(prefix Key, n *node) error {
	if n.hash != nil {
		return nil
	}
	b := t.bufs.Get().(*bytes.Buffer)
	defer t.bufs.Put(b)
	b.Reset()

	for ch, n := range n.links {
		if n == nil {
			continue
		}
		p := append(prefix, byte(ch))
		if err := t.merkle(p, n); err != nil {
			return err
		}
		if n.hash == nil {
			continue
		}
		if err := b.WriteByte(byte(ch)); err != nil {
			panic(err) // Can't happen. Kepp linter happy.
		}
		if _, err := b.Write(n.hash[:]); err != nil {
			panic(err) // Can't happen. Kepp linter happy.
		}
	}

	if n.hasValue {
		val, err := t.kv.Get(prefix)
		if err != nil {
			return errors.Wrapf(err, "t.kv.get(%s)", prefix)
		}
		if h := val.Hash(); h != nil {
			if _, err = b.Write(h[:]); err != nil {
				panic(err) // Can't happen. Kepp linter happy.
			}
		}
	}

	if b.Len() == 0 {
		return nil
	}
	h := chainhash.DoubleHashH(b.Bytes())
	n.hash = &h
	if t.db != nil {
		t.batch.Put(n.hash[:], b.Bytes())
	}
	return nil
}
