package claimtrie

import (
	"fmt"

	"github.com/lbryio/claimtrie/change"
	"github.com/lbryio/claimtrie/claim"
	"github.com/lbryio/claimtrie/nodemgr"
	"github.com/lbryio/claimtrie/trie"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

// ClaimTrie implements a Merkle Trie supporting linear history of commits.
type ClaimTrie struct {
	cm *CommitMgr
	nm *nodemgr.NodeMgr
	tr *trie.Trie
}

// New returns a ClaimTrie.
func New(dbCommit, dbTrie, dbNodeMgr *leveldb.DB) *ClaimTrie {
	nm := nodemgr.New(dbNodeMgr)
	cm := NewCommitMgr(dbCommit)

	return &ClaimTrie{
		cm: cm,
		nm: nm,
		tr: trie.New(nm, dbTrie),
	}
}

// Load loads ClaimTrie, NodeManager, Trie from databases.
func (ct *ClaimTrie) Load() error {
	if err := ct.cm.Load(); err != nil {
		return errors.Wrapf(err, "cm.Load()")
	}
	fmt.Printf("%d of commits loaded. Head: %d\n", len(ct.cm.commits), ct.cm.head.Meta.Height)

	ct.nm.Load(ct.Height())
	fmt.Printf("%d of nodes loaded.\n", ct.nm.Size())

	ct.tr.SetRoot(ct.cm.Head().MerkleRoot)
	fmt.Printf("Trie root: %s.\n", ct.MerkleHash())
	return nil
}

// Save saves ClaimTrie state to database.
func (ct *ClaimTrie) Save() error {
	if err := ct.cm.Save(); err != nil {
		return errors.Wrapf(err, "cm.Save()")
	}
	return nil
}

// Height returns the highest height of blocks commited to the ClaimTrie.
func (ct *ClaimTrie) Height() claim.Height {
	return ct.cm.Head().Meta.Height
}

// Head returns the tip commit in the commit database.
func (ct *ClaimTrie) Head() *Commit {
	return ct.cm.Head()
}

// Trie returns the MerkleTrie of the ClaimTrie .
func (ct *ClaimTrie) Trie() *trie.Trie {
	return ct.tr
}

// NodeMgr returns the Node Manager of the ClaimTrie .
func (ct *ClaimTrie) NodeMgr() *nodemgr.NodeMgr {
	return ct.nm
}

// CommitMgr returns the Commit Manager of the ClaimTrie .
func (ct *ClaimTrie) CommitMgr() *CommitMgr {
	return ct.cm
}

// AddClaim adds a Claim to the ClaimTrie.
func (ct *ClaimTrie) AddClaim(name string, op claim.OutPoint, amt claim.Amount, val []byte) error {
	c := change.New(change.AddClaim).SetOP(op).SetAmt(amt).SetValue(val)
	return ct.modify(name, c)
}

// SpendClaim spend a Claim in the ClaimTrie.
func (ct *ClaimTrie) SpendClaim(name string, op claim.OutPoint) error {
	c := change.New(change.SpendClaim).SetOP(op)
	return ct.modify(name, c)
}

// UpdateClaim updates a Claim in the ClaimTrie.
func (ct *ClaimTrie) UpdateClaim(name string, op claim.OutPoint, amt claim.Amount, id claim.ID, val []byte) error {
	c := change.New(change.UpdateClaim).SetOP(op).SetAmt(amt).SetID(id).SetValue(val)
	return ct.modify(name, c)
}

// AddSupport adds a Support to the ClaimTrie.
func (ct *ClaimTrie) AddSupport(name string, op claim.OutPoint, amt claim.Amount, id claim.ID) error {
	c := change.New(change.AddSupport).SetOP(op).SetAmt(amt).SetID(id)
	return ct.modify(name, c)
}

// SpendSupport spend a support in the ClaimTrie.
func (ct *ClaimTrie) SpendSupport(name string, op claim.OutPoint) error {
	c := change.New(change.SpendSupport).SetOP(op)
	return ct.modify(name, c)
}

func (ct *ClaimTrie) modify(name string, c *change.Change) error {
	c.SetHeight(ct.Height() + 1).SetName(name)
	if err := ct.nm.ModifyNode(name, c); err != nil {
		return err
	}
	ct.tr.Update([]byte(name))
	return nil
}

// MerkleHash returns the Merkle Hash of the ClaimTrie.
func (ct *ClaimTrie) MerkleHash() *chainhash.Hash {
	return ct.tr.MerkleHash()
}

// Commit commits the current changes into database.
func (ct *ClaimTrie) Commit(ht claim.Height) {
	if ht < ct.Height() {
		return
	}
	for i := ct.Height() + 1; i <= ht; i++ {
		ct.nm.CatchUp(i, ct.tr.Update)
	}
	h := ct.MerkleHash()
	ct.cm.Commit(ht, h)
	ct.tr.SetRoot(h)
}

// Reset resets the tip commit to a previous height specified.
func (ct *ClaimTrie) Reset(ht claim.Height) error {
	if ht > ct.Height() {
		return ErrInvalidHeight
	}
	ct.cm.Reset(ht)
	ct.nm.Reset(ht)
	ct.tr.SetRoot(ct.Head().MerkleRoot)
	return nil
}
