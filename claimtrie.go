package claimtrie

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"

	"github.com/lbryio/merkletrie"
)

// Height ...
type Height int64

// Amount ...
type Amount int64

// ClaimTrie implements a Merkle Trie supporting linear history of commits.
type ClaimTrie struct {

	// The highest block number commited to the ClaimTrie.
	bestBlock Height

	// Immutable linear history.
	head *merkletrie.Commit

	// An overlay supporting Copy-on-Write to the current tip commit.
	stg *merkletrie.Stage
}

// CommitMeta implements merkletrie.CommitMeta with commit-specific metadata.
type CommitMeta struct {
	Height Height
}

// New returns a ClaimTrie.
func New() *ClaimTrie {
	mt := merkletrie.New()
	return &ClaimTrie{
		head: merkletrie.NewCommit(nil, CommitMeta{0}, mt),
		stg:  merkletrie.NewStage(mt),
	}
}

func updateStageNode(stg *merkletrie.Stage, name string, modifier func(n *node) error) error {
	v, err := stg.Get(merkletrie.Key(name))
	if err != nil && err != merkletrie.ErrKeyNotFound {
		return err
	}
	var n *node
	if v == nil {
		n = newNode()
	} else {
		n = v.(*node).clone()
	}
	if err = modifier(n); err != nil {
		return err
	}
	return stg.Update(merkletrie.Key(name), n)
}

// AddClaim adds a Claim to the Stage of ClaimTrie.
func (ct *ClaimTrie) AddClaim(name string, op wire.OutPoint, amt Amount, accepted Height) error {
	return updateStageNode(ct.stg, name, func(n *node) error {
		return n.addClaim(NewClaim(op, amt, accepted))
	})
}

// AddSupport adds a Support to the Stage of ClaimTrie.
func (ct *ClaimTrie) AddSupport(name string, op wire.OutPoint, amt Amount, accepted Height, supported ClaimID) error {
	return updateStageNode(ct.stg, name, func(n *node) error {
		return n.addSupport(NewSupport(op, amt, accepted, supported))
	})
}

// SpendClaim removes a Claim in the Stage.
func (ct *ClaimTrie) SpendClaim(name string, op wire.OutPoint) error {
	return updateStageNode(ct.stg, name, func(n *node) error {
		return n.removeClaim(op)
	})
}

// SpendSupport removes a Support in the Stage.
func (ct *ClaimTrie) SpendSupport(name string, op wire.OutPoint) error {
	return updateStageNode(ct.stg, name, func(n *node) error {
		return n.removeSupport(op)
	})
}

// Traverse visits Nodes in the Stage of the ClaimTrie.
func (ct *ClaimTrie) Traverse(visit merkletrie.Visit, update, valueOnly bool) error {
	// wrapper function to make sure the node is updated before it's observed externally.
	fn := func(prefix merkletrie.Key, v merkletrie.Value) error {
		v.(*node).updateBestClaim(ct.bestBlock)
		return visit(prefix, v)
	}
	return ct.stg.Traverse(fn, update, valueOnly)
}

// MerkleHash returns the Merkle Hash of the Stage.
func (ct *ClaimTrie) MerkleHash() chainhash.Hash {
	return ct.stg.MerkleHash()
}

// BestBlock returns the highest height of blocks commited to the ClaimTrie.
func (ct *ClaimTrie) BestBlock() Height {
	return ct.bestBlock
}

// Commit commits the current Stage into commit database, and updates the BestBlock with the associated height.
// The height must be higher than the current BestBlock, or ErrInvalidHeight is returned.
func (ct *ClaimTrie) Commit(h Height) error {
	if h <= ct.bestBlock {
		return ErrInvalidHeight
	}
	visit := func(prefix merkletrie.Key, v merkletrie.Value) error {
		v.(*node).updateBestClaim(h)
		return nil
	}
	ct.Traverse(visit, true, true)

	commit, err := ct.stg.Commit(ct.head, CommitMeta{Height: h})
	if err != nil {
		return err
	}
	ct.head = commit
	ct.bestBlock = h
	return nil
}

// Reset reverts the Stage to a specified commit by height.
func (ct *ClaimTrie) Reset(h Height) error {
	for commit := ct.head; commit != nil; commit = commit.Prev {
		meta := commit.Meta.(CommitMeta)
		if meta.Height <= h {
			ct.head = commit
			ct.bestBlock = h
			return nil
		}
	}
	return ErrInvalidHeight
}

// Head returns the current tip commit in the commit database.
func (ct *ClaimTrie) Head() *merkletrie.Commit {
	return ct.head
}
