package claimtrie

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"

	"github.com/lbryio/claimtrie/claim"
	"github.com/lbryio/claimtrie/claimnode"

	"github.com/lbryio/claimtrie/trie"
)

// ClaimTrie implements a Merkle Trie supporting linear history of commits.
type ClaimTrie struct {

	// The highest block number commited to the ClaimTrie.
	height claim.Height

	// Immutable linear history.
	head *trie.Commit

	// An overlay supporting Copy-on-Write to the current tip commit.
	stg *trie.Stage

	// todos tracks pending updates for future block height.
	//
	// A claim or support has a dynamic active peroid (ActiveAt, ExipresAt).
	// This makes the state of each node dynamic as the ClaimTrie increases/decreases its height.
	// Instead of polling every node for updates everytime ClaimTrie changes, the node is evaluated
	// for the nearest future height it may change the states, and add that height to the todos.
	//
	// When a ClaimTrie at height h1 is committed with h2, the pending updates from todos (h1, h2]
	// will be applied to bring the nodes up to date.
	todos map[claim.Height][]string
}

// CommitMeta implements trie.CommitMeta with commit-specific metadata.
type CommitMeta struct {
	Height claim.Height
}

// New returns a ClaimTrie.
func New() *ClaimTrie {
	mt := trie.New()
	return &ClaimTrie{
		head:  trie.NewCommit(nil, CommitMeta{0}, mt),
		stg:   trie.NewStage(mt),
		todos: map[claim.Height][]string{},
	}
}

// Height returns the highest height of blocks commited to the ClaimTrie.
func (ct *ClaimTrie) Height() claim.Height {
	return ct.height
}

// Head returns the tip commit in the commit database.
func (ct *ClaimTrie) Head() *trie.Commit {
	return ct.head
}

// AddClaim adds a Claim to the Stage of ClaimTrie.
func (ct *ClaimTrie) AddClaim(name string, op wire.OutPoint, amt claim.Amount) error {
	modifier := func(n *claimnode.Node) error {
		_, err := n.AddClaim(op, amt)
		return err
	}
	return updateNode(ct, ct.height, name, modifier)
}

// AddSupport adds a Support to the Stage of ClaimTrie.
func (ct *ClaimTrie) AddSupport(name string, op wire.OutPoint, amt claim.Amount, supported claim.ID) error {
	modifier := func(n *claimnode.Node) error {
		_, err := n.AddSupport(op, amt, supported)
		return err
	}
	return updateNode(ct, ct.height, name, modifier)
}

// SpendClaim removes a Claim in the Stage.
func (ct *ClaimTrie) SpendClaim(name string, op wire.OutPoint) error {
	modifier := func(n *claimnode.Node) error {
		return n.RemoveClaim(op)
	}
	return updateNode(ct, ct.height, name, modifier)
}

// SpendSupport removes a Support in the Stage.
func (ct *ClaimTrie) SpendSupport(name string, op wire.OutPoint) error {
	modifier := func(n *claimnode.Node) error {
		return n.RemoveSupport(op)
	}
	return updateNode(ct, ct.height, name, modifier)
}

// Traverse visits Nodes in the Stage.
func (ct *ClaimTrie) Traverse(visit trie.Visit, update, valueOnly bool) error {
	return ct.stg.Traverse(visit, update, valueOnly)
}

// MerkleHash returns the Merkle Hash of the Stage.
func (ct *ClaimTrie) MerkleHash() chainhash.Hash {
	return ct.stg.MerkleHash()
}

// Commit commits the current Stage into commit database.
// If h is lower than the current height, ErrInvalidHeight is returned.
//
// As Stage can be always cleanly reset to a specific commited snapshot,
// any error occurred during the commit would leave the Stage partially updated
// so the caller can inspect the status if interested.
//
// Changes to the ClaimTrie status, such as height or todos, are all or nothing.
func (ct *ClaimTrie) Commit(h claim.Height) error {

	// Already caught up.
	if h <= ct.height {
		return ErrInvalidHeight
	}

	// Apply pending updates in todos (ct.Height, h].
	// Note that ct.Height is excluded while h is included.
	for i := ct.height + 1; i <= h; i++ {
		for _, name := range ct.todos[i] {
			// dummy modifier to have the node brought up to date.
			modifier := func(n *claimnode.Node) error { return nil }
			if err := updateNode(ct, i, name, modifier); err != nil {
				return err
			}
		}
	}
	commit, err := ct.stg.Commit(ct.head, CommitMeta{Height: h})
	if err != nil {
		return err
	}

	// No more errors. Change the ClaimTrie status.
	ct.head = commit
	for i := ct.height + 1; i <= h; i++ {
		delete(ct.todos, i)
	}
	ct.height = h
	return nil
}

// Reset reverts the Stage to a specified commit by height.
func (ct *ClaimTrie) Reset(h claim.Height) error {
	if h > ct.height {
		return ErrInvalidHeight
	}

	// Find the most recent commit that is equal or earlier than h.
	commit := ct.head
	for commit != nil {
		if commit.Meta.(CommitMeta).Height <= h {
			break
		}
		commit = commit.Prev
	}

	// The commit history is not deep enough.
	if commit == nil {
		return ErrInvalidHeight
	}

	// Drop (rollback) any uncommited change, and adjust to the specified height.
	rollback := func(prefix trie.Key, value trie.Value) error {
		n := value.(*claimnode.Node)
		n.Reset()
		return n.AdjustTo(h)
	}
	if err := ct.stg.Traverse(rollback, true, true); err != nil {
		// Rollback a node to a known state can't go wrong.
		// It's a programming error, and can't recover.
		panic(err)
	}

	// Update ClaimTrie status
	ct.head = commit
	ct.height = h
	for k := range ct.todos {
		if k >= h {
			delete(ct.todos, k)
		}
	}
	ct.stg = trie.NewStage(commit.MerkleTrie)
	return nil
}

// updateNode implements a get-modify-set sequence to the node associated with name.
// After the modifier is applied, the node is evaluated for how soon in the
// nearest future change. And register it, if any, to the todos for the next updateNode.
func updateNode(ct *ClaimTrie, h claim.Height, name string, modifier func(n *claimnode.Node) error) error {

	// Get the node from the Stage, or create one if it did not exist yet.
	v, err := ct.stg.Get(trie.Key(name))
	if err == trie.ErrKeyNotFound {
		v = claimnode.NewNode()
	} else if err != nil {
		return err
	}

	n := v.(*claimnode.Node)

	// Bring the node state up to date.
	if err = n.AdjustTo(h); err != nil {
		return err
	}

	// Apply the modifier on the node.
	if err = modifier(n); err != nil {
		return err
	}

	// Register pending update, if any, for future height.
	next := n.FindNextUpdateHeight()
	if next > h {
		ct.todos[next] = append(ct.todos[next], name)
	}

	// Store the modified value back to the Stage, clearing out all the Merkle Hash on the path.
	return ct.stg.Update(trie.Key(name), n)
}
