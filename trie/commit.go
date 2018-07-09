package trie

// CommitMeta ...
type CommitMeta interface{}

// NewCommit ...
func NewCommit(head *Commit, meta CommitMeta, mt *MerkleTrie) *Commit {
	commit := &Commit{
		Prev:       head,
		MerkleTrie: mt,
		Meta:       meta,
	}
	return commit
}

// Commit ...
type Commit struct {
	Prev       *Commit
	MerkleTrie *MerkleTrie
	Meta       CommitMeta
}
