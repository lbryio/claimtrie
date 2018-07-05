package claimtrie

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"

	"github.com/lbryio/merkletrie"
)

// ClaimTrie implements a Merkle Trie supporting linear history of commits.
type ClaimTrie struct {

	// The highest block number commited to the ClaimTrie.
	bestBlock Height

	// Immutable linear history.
	head *merkletrie.Commit

	// An overlay supporting Copy-on-Write to the current tip commit.
	stg *merkletrie.Stage

	// pending keeps track update for future block height.
	pending map[Height][]string
}

// CommitMeta implements merkletrie.CommitMeta with commit-specific metadata.
type CommitMeta struct {
	Height Height
}

// New returns a ClaimTrie.
func New() *ClaimTrie {
	mt := merkletrie.New()
	return &ClaimTrie{
		head:    merkletrie.NewCommit(nil, CommitMeta{0}, mt),
		stg:     merkletrie.NewStage(mt),
		pending: map[Height][]string{},
	}
}

func updateStageNode(stg *merkletrie.Stage, name string, modifier func(n *Node) error) error {
	v, err := stg.Get(merkletrie.Key(name))
	if err != nil && err != merkletrie.ErrKeyNotFound {
		return err
	}
	var n *Node
	if v == nil {
		n = NewNode()
	} else {
		n = v.(*Node)
	}
	if err = modifier(n); err != nil {
		return err
	}
	return stg.Update(merkletrie.Key(name), n)
}

// AddClaim adds a Claim to the Stage of ClaimTrie.
func (ct *ClaimTrie) AddClaim(name string, op wire.OutPoint, amt Amount, accepted Height) error {
	return updateStageNode(ct.stg, name, func(n *Node) error {
		n.IncrementBlock(ct.bestBlock - n.height)
		_, err := n.addClaim(op, amt)
		next := ct.bestBlock + 1
		ct.pending[next] = append(ct.pending[next], name)
		return err
	})
}

// AddSupport adds a Support to the Stage of ClaimTrie.
func (ct *ClaimTrie) AddSupport(name string, op wire.OutPoint, amt Amount, accepted Height, supported ClaimID) error {
	return updateStageNode(ct.stg, name, func(n *Node) error {
		_, err := n.addSupport(op, amt, supported)
		next := ct.bestBlock + 1
		ct.pending[next] = append(ct.pending[next], name)
		return err
	})
}

// SpendClaim removes a Claim in the Stage.
func (ct *ClaimTrie) SpendClaim(name string, op wire.OutPoint) error {
	return updateStageNode(ct.stg, name, func(n *Node) error {
		next := ct.bestBlock + 1
		ct.pending[next] = append(ct.pending[next], name)
		return n.removeClaim(op)
	})
}

// SpendSupport removes a Support in the Stage.
func (ct *ClaimTrie) SpendSupport(name string, op wire.OutPoint) error {
	return updateStageNode(ct.stg, name, func(n *Node) error {
		next := ct.bestBlock + 1
		ct.pending[next] = append(ct.pending[next], name)
		return n.removeSupport(op)
	})
}

// Traverse visits Nodes in the Stage of the ClaimTrie.
func (ct *ClaimTrie) Traverse(visit merkletrie.Visit, update, valueOnly bool) error {
	return ct.stg.Traverse(visit, update, valueOnly)
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

	for i := ct.bestBlock + 1; i <= h; i++ {
		for _, prefix := range ct.pending[i] {
			// Brings the value node to date.
			catchup := func(n *Node) error {
				if err := n.IncrementBlock(i - n.height); err != nil {
					return err
				}

				// After the update, the node may subscribe to another pending update.
				if next := n.findNextUpdateHeights(); next > i {
					fmt.Printf("Subscribe pendings for %v to future Height at %d\n", prefix, next)
					ct.pending[next] = append(ct.pending[next], prefix)
				}
				return nil
			}

			// Update the node with the catchup modifier, and clear the Merkle Hash along the way.
			if err := updateStageNode(ct.stg, prefix, catchup); err != nil {
				return err
			}
		}
		delete(ct.pending, i)
	}
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
	visit := func(prefix merkletrie.Key, value merkletrie.Value) error {
		n := value.(*Node)
		return n.DecrementBlock(n.height - h)
	}
	if err := ct.stg.Traverse(visit, true, true); err != nil {
		return err
	}
	for commit := ct.head; commit != nil; commit = commit.Prev {
		meta := commit.Meta.(CommitMeta)
		if meta.Height <= h {
			ct.head = commit
			ct.bestBlock = h
			ct.stg = merkletrie.NewStage(commit.MerkleTrie)
			return nil
		}
	}
	return ErrInvalidHeight
}

// Head returns the current tip commit in the commit database.
func (ct *ClaimTrie) Head() *merkletrie.Commit {
	return ct.head
}
