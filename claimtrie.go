package claimtrie

import (
	"fmt"

	"github.com/lbryio/claimtrie/claim"
	"github.com/lbryio/claimtrie/nodemgr"
	"github.com/lbryio/claimtrie/trie"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

// ClaimTrie implements a Merkle Trie supporting linear history of commits.
type ClaimTrie struct {
	height claim.Height
	head   *Commit
	stg    *trie.Trie
	nm     *nodemgr.NodeMgr
}

// New returns a ClaimTrie.
func New(dbTrie, dbNodeMgr *leveldb.DB) *ClaimTrie {
	nm := nodemgr.New(dbNodeMgr)
	return &ClaimTrie{
		head: newCommit(nil, CommitMeta{0}, trie.EmptyTrieHash),
		nm:   nm,
		stg:  trie.New(nm, dbTrie),
	}
}

// Height returns the highest height of blocks commited to the ClaimTrie.
func (ct *ClaimTrie) Height() claim.Height {
	return ct.height
}

// Head returns the tip commit in the commit database.
func (ct *ClaimTrie) Head() *Commit {
	return ct.head
}

// Trie returns the Stage of the claimtrie .
func (ct *ClaimTrie) Trie() *trie.Trie {
	return ct.stg
}

// NodeMgr returns the Node Manager of the claimtrie .
func (ct *ClaimTrie) NodeMgr() *nodemgr.NodeMgr {
	return ct.nm
}

// AddClaim adds a Claim to the Stage.
func (ct *ClaimTrie) AddClaim(name string, op claim.OutPoint, amt claim.Amount) error {
	modifier := func(n *claim.Node) error {
		return n.AddClaim(op, amt)
	}
	return ct.updateNode(name, modifier)
}

// SpendClaim spend a Claim in the Stage.
func (ct *ClaimTrie) SpendClaim(name string, op claim.OutPoint) error {
	modifier := func(n *claim.Node) error {
		return n.SpendClaim(op)
	}
	return ct.updateNode(name, modifier)
}

// UpdateClaim updates a Claim in the Stage.
func (ct *ClaimTrie) UpdateClaim(name string, op claim.OutPoint, amt claim.Amount, id claim.ID) error {
	modifier := func(n *claim.Node) error {
		return n.UpdateClaim(op, amt, id)
	}
	return ct.updateNode(name, modifier)
}

// AddSupport adds a Support to the Stage.
func (ct *ClaimTrie) AddSupport(name string, op claim.OutPoint, amt claim.Amount, id claim.ID) error {
	modifier := func(n *claim.Node) error {
		return n.AddSupport(op, amt, id)
	}
	return ct.updateNode(name, modifier)
}

// SpendSupport spend a support in the Stage.
func (ct *ClaimTrie) SpendSupport(name string, op claim.OutPoint) error {
	modifier := func(n *claim.Node) error {
		return n.SpendSupport(op)
	}
	return ct.updateNode(name, modifier)
}

// Traverse visits Nodes in the Stage.
func (ct *ClaimTrie) Traverse(visit trie.Visit) error {
	return ct.stg.Traverse(visit)
}

// MerkleHash returns the Merkle Hash of the Stage.
func (ct *ClaimTrie) MerkleHash() (*chainhash.Hash, error) {
	// ct.nm.UpdateAll(ct.stg.Update)
	return ct.stg.MerkleHash()
}

// Commit commits the current Stage into database.
func (ct *ClaimTrie) Commit(h claim.Height) error {
	if h < ct.height {
		return errors.Wrapf(ErrInvalidHeight, "%d < ct.height %d", h, ct.height)
	}

	for i := ct.height + 1; i <= h; i++ {
		if err := ct.nm.CatchUp(i, ct.stg.Update); err != nil {
			return errors.Wrapf(err, "nm.CatchUp(%d, stg.Update)", i)
		}
	}
	hash, err := ct.MerkleHash()
	if err != nil {
		return errors.Wrapf(err, "MerkleHash()")
	}
	commit := newCommit(ct.head, CommitMeta{Height: h}, hash)
	ct.head = commit
	ct.height = h
	ct.stg.SetRoot(hash)
	return nil
}

// Reset reverts the Stage to the current or previous height specified.
func (ct *ClaimTrie) Reset(h claim.Height) error {
	if h > ct.height {
		return errors.Wrapf(ErrInvalidHeight, "%d > ct.height %d", h, ct.height)
	}
	fmt.Printf("ct.Reset from %d to %d\n", ct.height, h)
	commit := ct.head
	for commit.Meta.Height > h {
		commit = commit.Prev
	}
	if err := ct.nm.Reset(h); err != nil {
		return errors.Wrapf(err, "nm.Reset(%d)", h)
	}
	ct.head = commit
	ct.height = h
	ct.stg.SetRoot(commit.MerkleRoot)
	return nil
}

func (ct *ClaimTrie) updateNode(name string, modifier func(n *claim.Node) error) error {
	if err := ct.nm.ModifyNode(name, ct.height, modifier); err != nil {
		return errors.Wrapf(err, "nm.ModifyNode(%s, %d)", name, ct.height)
	}
	if err := ct.stg.Update(trie.Key(name)); err != nil {
		return errors.Wrapf(err, "stg.Update(%s)", name)
	}
	return nil
}
