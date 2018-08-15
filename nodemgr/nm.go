package nodemgr

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"sort"

	"github.com/lbryio/claimtrie/change"
	"github.com/lbryio/claimtrie/claim"
	"github.com/lbryio/claimtrie/trie"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

// NodeMgr ...
type NodeMgr struct {
	height claim.Height

	db          *leveldb.DB
	cache       map[string]*claim.Node
	nextUpdates todos
}

// New ...
func New(db *leveldb.DB) *NodeMgr {
	nm := &NodeMgr{
		db:          db,
		cache:       map[string]*claim.Node{},
		nextUpdates: todos{},
	}
	return nm
}

// Load loads the nodes from the database up to height ht.
func (nm *NodeMgr) Load(ht claim.Height) {
	nm.height = ht
	iter := nm.db.NewIterator(nil, nil)
	for iter.Next() {
		name := string(iter.Key())
		nm.cache[name] = nm.load(name, ht)
	}
	data, err := nm.db.Get([]byte("nextUpdates"), nil)
	if err == leveldb.ErrNotFound {
		return
	} else if err != nil {
		panic(err)
	}
	if err = gob.NewDecoder(bytes.NewBuffer(data)).Decode(&nm.nextUpdates); err != nil {
		panic(err)
	}
}

// Save saves the states to the database.
func (nm *NodeMgr) Save() error {
	buf := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buf).Encode(nm.nextUpdates); err != nil {
		return errors.Wrapf(err, "gob.Encode()")
	}
	if err := nm.db.Put([]byte("nextUpdates"), buf.Bytes(), nil); err != nil {
		return errors.Wrapf(err, "db.Put()")
	}
	return nil
}

// Get returns the latest node with name specified by key.
func (nm *NodeMgr) Get(key []byte) trie.Value {
	return nm.NodeAt(string(key), nm.height)
}

// Reset resets all nodes to specified height.
func (nm *NodeMgr) Reset(ht claim.Height) {
	nm.height = ht
	for name, n := range nm.cache {
		if n.Height() >= ht {
			nm.cache[name] = nm.load(name, ht)
		}
	}
}

// Size returns the number of nodes loaded into the cache.
func (nm *NodeMgr) Size() int {
	return len(nm.cache)
}

func (nm *NodeMgr) load(name string, ht claim.Height) *claim.Node {
	c := change.NewChangeList(nm.db, name).Load().Truncate(ht).Changes()
	return NewFromChanges(name, c, ht)
}

// NodeAt returns the node adjusted to specified height.
func (nm *NodeMgr) NodeAt(name string, ht claim.Height) *claim.Node {
	n, ok := nm.cache[name]
	if !ok {
		n = claim.NewNode(name)
		nm.cache[name] = n
	}

	// Cached version is too new.
	if n.Height() > nm.height || n.Height() > ht {
		n = nm.load(name, ht)
	}
	return n.AdjustTo(ht)
}

// ModifyNode returns the node adjusted to specified height.
func (nm *NodeMgr) ModifyNode(name string, chg *change.Change) error {
	ht := nm.height
	n := nm.NodeAt(name, ht)
	n.AdjustTo(ht)
	if err := execute(n, chg); err != nil {
		return errors.Wrapf(err, "claim.execute(n,chg)")
	}
	nm.cache[name] = n
	nm.nextUpdates.set(name, ht+1)
	change.NewChangeList(nm.db, name).Load().Append(chg).Save()
	return nil
}

// CatchUp ...
func (nm *NodeMgr) CatchUp(ht claim.Height, notifier func(key []byte)) {
	nm.height = ht
	for name := range nm.nextUpdates[ht] {
		notifier([]byte(name))
		if next := nm.NodeAt(name, ht).NextUpdate(); next > ht {
			nm.nextUpdates.set(name, next)
		}
	}
}

// VisitFunc ...
type VisitFunc func(n *claim.Node) (stop bool)

// Visit visits every node in the cache with VisiFunc.
func (nm *NodeMgr) Visit(v VisitFunc) {
	for _, n := range nm.cache {
		if v(n) {
			return
		}
	}
}

// Show is a conevenient function for debugging and velopment purpose.
// The proper way to handle user request would be a query function with filters specified.
func (nm *NodeMgr) Show(name string, ht claim.Height, dump bool) error {
	names := []string{}
	if len(name) != 0 {
		names = append(names, name)
	} else {
		for name := range nm.cache {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	for _, name := range names {
		n := nm.NodeAt(name, ht)
		if n.BestClaim() == nil {
			continue
		}
		fmt.Printf("[%s] %s\n", name, n)
		if dump {
			change.NewChangeList(nm.db, name).Load().Truncate(ht).Dump()
		}
	}
	return nil
}

// NewFromChanges ...
func NewFromChanges(name string, chgs []*change.Change, ht claim.Height) *claim.Node {
	return replay(name, chgs).AdjustTo(ht)
}

func replay(name string, chgs []*change.Change) *claim.Node {
	n := claim.NewNode(name)
	for _, chg := range chgs {
		if n.Height() < chg.Height-1 {
			n.AdjustTo(chg.Height - 1)
		}
		if n.Height() == chg.Height-1 {
			if err := execute(n, chg); err != nil {
				panic(err)
			}
		}
	}
	return n
}

func execute(n *claim.Node, c *change.Change) error {
	var err error
	switch c.Cmd {
	case change.AddClaim:
		err = n.AddClaim(c.OP, c.Amt, c.Value)
	case change.SpendClaim:
		err = n.SpendClaim(c.OP)
	case change.UpdateClaim:
		err = n.UpdateClaim(c.OP, c.Amt, c.ID, c.Value)
	case change.AddSupport:
		err = n.AddSupport(c.OP, c.Amt, c.ID)
	case change.SpendSupport:
		err = n.SpendSupport(c.OP)
	}
	return errors.Wrapf(err, "chg %s", c)
}

type todos map[claim.Height]map[string]bool

func (t todos) set(name string, ht claim.Height) {
	if t[ht] == nil {
		t[ht] = map[string]bool{}
	}
	t[ht][name] = true
}
