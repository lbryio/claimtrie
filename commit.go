package claimtrie

import (
	"bytes"
	"encoding/gob"

	"github.com/lbryio/claimtrie/claim"
	"github.com/lbryio/claimtrie/trie"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

// CommitVisit ...
type CommitVisit func(c *Commit)

// CommitMeta represent the meta associated with each commit.
type CommitMeta struct {
	Height claim.Height
}

func newCommit(head *Commit, meta CommitMeta, h *chainhash.Hash) *Commit {
	return &Commit{
		MerkleRoot: h,
		Meta:       meta,
	}
}

// Commit ...
type Commit struct {
	MerkleRoot *chainhash.Hash
	Meta       CommitMeta
}

// CommitMgr ...
type CommitMgr struct {
	db      *leveldb.DB
	commits []*Commit
	head    *Commit
}

// NewCommitMgr ...
func NewCommitMgr(db *leveldb.DB) *CommitMgr {
	head := newCommit(nil, CommitMeta{0}, trie.EmptyTrieHash)
	cm := CommitMgr{
		db:   db,
		head: head,
	}
	cm.commits = append(cm.commits, head)
	return &cm
}

// Head ...
func (cm *CommitMgr) Head() *Commit {
	return cm.head
}

// Commit ...
func (cm *CommitMgr) Commit(ht claim.Height, merkle *chainhash.Hash) {
	if ht == 0 {
		return
	}
	c := newCommit(cm.head, CommitMeta{ht}, merkle)
	cm.commits = append(cm.commits, c)
	cm.head = c
}

// Reset ...
func (cm *CommitMgr) Reset(ht claim.Height) {
	for i := len(cm.commits) - 1; i >= 0; i-- {
		c := cm.commits[i]
		if c.Meta.Height <= ht {
			cm.head = c
			cm.commits = cm.commits[:i+1]
			break
		}
	}
	if cm.head.Meta.Height == ht {
		return
	}
	cm.Commit(ht, cm.head.MerkleRoot)
}

// Save ...
func (cm *CommitMgr) Save() error {
	exported := struct {
		Commits []*Commit
		Head    *Commit
	}{
		Commits: cm.commits,
		Head:    cm.head,
	}

	buf := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buf).Encode(exported); err != nil {
		return errors.Wrapf(err, "gob.Encode()", err)
	}
	if err := cm.db.Put([]byte("CommitMgr"), buf.Bytes(), nil); err != nil {
		return errors.Wrapf(err, "db.Put(CommitMgr)")
	}
	return nil
}

// Load ...
func (cm *CommitMgr) Load() error {
	exported := struct {
		Commits []*Commit
		Head    *Commit
	}{}

	data, err := cm.db.Get([]byte("CommitMgr"), nil)
	if err != nil {
		return errors.Wrapf(err, "db.Get(CommitMgr)")
	}
	if err := gob.NewDecoder(bytes.NewBuffer(data)).Decode(&exported); err != nil {
		return errors.Wrapf(err, "gob.Encode()", err)
	}
	cm.commits = exported.Commits
	cm.head = exported.Head
	return nil
}

// Log ...
func (cm *CommitMgr) Log(ht claim.Height, visit CommitVisit) {
	for i := len(cm.commits) - 1; i >= 0; i-- {
		c := cm.commits[i]
		if c.Meta.Height > ht {
			continue
		}
		visit(c)
	}
}
