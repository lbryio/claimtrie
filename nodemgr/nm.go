package nodemgr

import (
	"fmt"
	"sort"
	"sync"

	"github.com/lbryio/claimtrie/claim"
	"github.com/lbryio/claimtrie/trie"
	"github.com/syndtr/goleveldb/leveldb"
)

// NodeMgr ...
type NodeMgr struct {
	sync.RWMutex

	db          *leveldb.DB
	nodes       map[string]*claim.Node
	dirty       map[string]bool
	nextUpdates todos
}

// New ...
func New(db *leveldb.DB) *NodeMgr {
	nm := &NodeMgr{
		db:          db,
		nodes:       map[string]*claim.Node{},
		dirty:       map[string]bool{},
		nextUpdates: todos{},
	}
	return nm
}

// Get ...
func (nm *NodeMgr) Get(key trie.Key) (trie.Value, error) {
	nm.Lock()
	defer nm.Unlock()

	if n, ok := nm.nodes[string(key)]; ok {
		return n, nil
	}
	if nm.db != nil {
		b, err := nm.db.Get(key, nil)
		if err == nil {
			_ = b // TODO: Loaded. Deserialize it.
		} else if err != leveldb.ErrNotFound {
			// DB error. Propagated.
			return nil, err
		}
	}
	// New node.
	n := claim.NewNode(string(key))
	nm.nodes[string(key)] = n
	return n, nil
}

// Set ...
func (nm *NodeMgr) Set(key trie.Key, val trie.Value) {
	n := val.(*claim.Node)

	nm.Lock()
	defer nm.Unlock()

	nm.nodes[string(key)] = n
	nm.dirty[string(key)] = true

	// TODO: flush to disk.
}

// Reset resets all nodes to specified height.
func (nm *NodeMgr) Reset(h claim.Height) error {
	for _, n := range nm.nodes {
		if err := n.Reset(h); err != nil {
			return err
		}
	}
	return nil
}

// NodeAt returns the node adjusted to specified height.
func (nm *NodeMgr) NodeAt(name string, h claim.Height) (*claim.Node, error) {
	v, err := nm.Get(trie.Key(name))
	if err != nil {
		return nil, err
	}
	n := v.(*claim.Node)
	if err = n.AdjustTo(h); err != nil {
		return nil, err
	}
	return n, nil
}

// ModifyNode returns the node adjusted to specified height.
func (nm *NodeMgr) ModifyNode(name string, h claim.Height, modifier func(*claim.Node) error) error {
	n, err := nm.NodeAt(name, h)
	if err != nil {
		return err
	}
	if err = modifier(n); err != nil {
		return err
	}
	nm.nextUpdates.set(name, h+1)
	return nil
}

// CatchUp ...
func (nm *NodeMgr) CatchUp(h claim.Height, notifier func(key trie.Key) error) error {
	for name := range nm.nextUpdates[h] {
		n, err := nm.NodeAt(name, h)
		if err != nil {
			return err
		}
		if err = notifier(trie.Key(name)); err != nil {
			return err
		}
		if next := n.NextUpdate(); next > h {
			nm.nextUpdates.set(name, next)
		}
	}
	return nil
}

// Show ...
func (nm *NodeMgr) Show(name string) error {
	if len(name) != 0 {
		fmt.Printf("[%s] %s\n", name, nm.nodes[name])
		return nil
	}
	names := []string{}
	for name := range nm.nodes {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Printf("[%s] %s\n", name, nm.nodes[name])
	}
	return nil
}

// UpdateAll ...
func (nm *NodeMgr) UpdateAll(m func(key trie.Key) error) error {
	for name := range nm.nodes {
		m(trie.Key(name))
	}
	return nil
}

type todos map[claim.Height]map[string]bool

func (t todos) set(name string, h claim.Height) {
	if t[h] == nil {
		t[h] = map[string]bool{}
	}
	t[h][name] = true
}
