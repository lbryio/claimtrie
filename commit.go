package claimtrie

import (
	"github.com/lbryio/claimtrie/claim"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// CommitMeta represent the meta associated with each commit.
type CommitMeta struct {
	Height claim.Height
}

func newCommit(head *Commit, meta CommitMeta, h *chainhash.Hash) *Commit {
	return &Commit{
		Prev:       head,
		MerkleRoot: h,
		Meta:       meta,
	}
}

// Commit ...
type Commit struct {
	Prev       *Commit
	MerkleRoot *chainhash.Hash
	Meta       CommitMeta
}

// CommitVisit ...
type CommitVisit func(c *Commit)

// Log ...
func Log(commit *Commit, visit CommitVisit) error {
	for commit != nil {
		visit(commit)
		commit = commit.Prev
	}
	return nil
}
