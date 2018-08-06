package trie

import (
	"bytes"
	"sync"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/syndtr/goleveldb/leveldb"
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
func (t *Trie) Update(key []byte) {
	n := t.root
	for _, ch := range key {
		t.resolve(n)
		if n.links[ch] == nil {
			n.links[ch] = newNode()
		}
		n.hash = nil
		n = n.links[ch]
	}
	t.resolve(n)
	n.hasValue = true
	n.hash = nil
}

func (t *Trie) resolve(n *node) {
	if n.hash == nil {
		return
	}
	b, err := t.db.Get(n.hash[:], nil)
	if err == leveldb.ErrNotFound {
		return
	} else if err != nil {
		panic(err)
	}

	nb := nbuf(b)
	n.hasValue = nb.hasValue()
	for i := 0; i < nb.entries(); i++ {
		p, h := nb.entry(i)
		n.links[p] = newNode()
		n.links[p].hash = h
	}
}

// MerkleHash returns the Merkle Hash of the Trie.
// All nodes must have been resolved before calling this function.
func (t *Trie) MerkleHash() *chainhash.Hash {
	t.batch = &leveldb.Batch{}
	buf := make([]byte, 0, 4096)
	if h := t.merkle(buf, t.root); h == nil {
		return EmptyTrieHash
	}
	if t.batch.Len() != 0 {
		if err := t.db.Write(t.batch, nil); err != nil {
			panic(err)
		}
	}
	return t.root.hash
}

// merkle recursively resolves the hashes of the node.
// All nodes must have been resolved before calling this function.
func (t *Trie) merkle(prefix []byte, n *node) *chainhash.Hash {
	if n.hash != nil {
		return n.hash
	}
	b := t.bufs.Get().(*bytes.Buffer)
	defer t.bufs.Put(b)
	b.Reset()

	for ch, n := range n.links {
		if n == nil {
			continue
		}
		p := append(prefix, byte(ch))
		if h := t.merkle(p, n); h != nil {
			b.WriteByte(byte(ch)) // nolint : errchk
			b.Write(h[:])         // nolint : errchk
		}
	}

	if n.hasValue {
		if h := t.kv.Get(prefix).Hash(); h != nil {
			b.Write(h[:]) // nolint : errchk
		}
	}

	if b.Len() != 0 {
		h := chainhash.DoubleHashH(b.Bytes())
		n.hash = &h
		t.batch.Put(h[:], b.Bytes())
	}
	return n.hash
}
