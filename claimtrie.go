package claimtrie

import (
	"fmt"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"

	"github.com/lbryio/claimtrie/claim"
	"github.com/lbryio/claimtrie/claimnode"

	"github.com/lbryio/claimtrie/trie"
)

// ClaimTrie implements a Merkle Trie supporting linear history of commits.
type ClaimTrie struct {

	// The highest block number commited to the ClaimTrie.
	bestBlock claim.Height

	// Immutable linear history.
	head *trie.Commit

	// An overlay supporting Copy-on-Write to the current tip commit.
	stg *trie.Stage

	// pending keeps track update for future block height.
	pending map[claim.Height][]string
}

// CommitMeta implements trie.CommitMeta with commit-specific metadata.
type CommitMeta struct {
	Height claim.Height
}

// New returns a ClaimTrie.
func New() *ClaimTrie {
	mt := trie.New()
	return &ClaimTrie{
		head:    trie.NewCommit(nil, CommitMeta{0}, mt),
		stg:     trie.NewStage(mt),
		pending: map[claim.Height][]string{},
	}
}

func updateStageNode(stg *trie.Stage, name string, modifier func(n *claimnode.Node) error) error {
	v, err := stg.Get(trie.Key(name))
	if err != nil && err != trie.ErrKeyNotFound {
		return err
	}
	var n *claimnode.Node
	if v == nil {
		n = claimnode.NewNode()
	} else {
		n = v.(*claimnode.Node)
	}
	if err = modifier(n); err != nil {
		return err
	}
	return stg.Update(trie.Key(name), n)
}

// AddClaim adds a Claim to the Stage of ClaimTrie.
func (ct *ClaimTrie) AddClaim(name string, op wire.OutPoint, amt claim.Amount) error {
	return updateStageNode(ct.stg, name, func(n *claimnode.Node) error {
		if err := n.IncrementBlock(claim.Height(ct.bestBlock) - n.Height()); err != nil {
			return err
		}
		_, err := n.AddClaim(op, claim.Amount(amt))
		next := ct.bestBlock + 1
		ct.pending[next] = append(ct.pending[next], name)
		return err
	})
}

// AddSupport adds a Support to the Stage of ClaimTrie.
func (ct *ClaimTrie) AddSupport(name string, op wire.OutPoint, amt claim.Amount, supported claim.ID) error {
	return updateStageNode(ct.stg, name, func(n *claimnode.Node) error {
		_, err := n.AddSupport(op, amt, supported)
		next := ct.bestBlock + 1
		ct.pending[next] = append(ct.pending[next], name)
		return err
	})
}

// SpendClaim removes a Claim in the Stage.
func (ct *ClaimTrie) SpendClaim(name string, op wire.OutPoint) error {
	return updateStageNode(ct.stg, name, func(n *claimnode.Node) error {
		next := ct.bestBlock + 1
		ct.pending[next] = append(ct.pending[next], name)
		return n.RemoveClaim(op)
	})
}

// SpendSupport removes a Support in the Stage.
func (ct *ClaimTrie) SpendSupport(name string, op wire.OutPoint) error {
	return updateStageNode(ct.stg, name, func(n *claimnode.Node) error {
		next := ct.bestBlock + 1
		ct.pending[next] = append(ct.pending[next], name)
		return n.RemoveSupport(op)
	})
}

// Traverse visits Nodes in the Stage of the ClaimTrie.
func (ct *ClaimTrie) Traverse(visit trie.Visit, update, valueOnly bool) error {
	return ct.stg.Traverse(visit, update, valueOnly)
}

// MerkleHash returns the Merkle Hash of the Stage.
func (ct *ClaimTrie) MerkleHash() chainhash.Hash {
	return ct.stg.MerkleHash()
}

// BestBlock returns the highest height of blocks commited to the ClaimTrie.
func (ct *ClaimTrie) BestBlock() claim.Height {
	return ct.bestBlock
}

// Commit commits the current Stage into commit database, and updates the BestBlock with the associated height.
// The height must be higher than the current BestBlock, or ErrInvalidHeight is returned.
func (ct *ClaimTrie) Commit(h claim.Height) error {
	if h <= ct.bestBlock {
		return ErrInvalidHeight
	}

	for i := claim.Height(ct.bestBlock) + 1; i <= h; i++ {
		for _, prefix := range ct.pending[i] {
			// Brings the value node to date.
			catchup := func(n *claimnode.Node) error {
				if err := n.IncrementBlock(i - n.Height()); err != nil {
					return err
				}

				// After the update, the node may subscribe to another pending update.
				if next := n.FindNextUpdateHeights(); next > i {
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
func (ct *ClaimTrie) Reset(h claim.Height) error {
	visit := func(prefix trie.Key, value trie.Value) error {
		n := value.(*claimnode.Node)
		return n.DecrementBlock(n.Height() - claim.Height(h))
	}
	if err := ct.stg.Traverse(visit, true, true); err != nil {
		return err
	}
	for commit := ct.head; commit != nil; commit = commit.Prev {
		meta := commit.Meta.(CommitMeta)
		if meta.Height <= h {
			ct.head = commit
			ct.bestBlock = h
			ct.stg = trie.NewStage(commit.MerkleTrie)
			return nil
		}
	}
	return ErrInvalidHeight
}

// Head returns the current tip commit in the commit database.
func (ct *ClaimTrie) Head() *trie.Commit {
	return ct.head
}
