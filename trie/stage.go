package trie

// Stage implements Copy-on-Write staging area on top of a MerkleTrie.
type Stage struct {
	*MerkleTrie
}

// NewStage returns a Stage initialized with a specified MerkleTrie.
func NewStage(t *MerkleTrie) *Stage {
	s := &Stage{
		MerkleTrie: New(),
	}
	s.mu = t.mu
	s.root = newNode(nil)
	*s.root = *t.root
	return s
}

// Update updates the internal MerkleTrie in a Copy-on-Write manner.
func (s *Stage) Update(key Key, val Value) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	n := s.root
	n.hash = nil
	for _, k := range key {
		org := n.links[k]
		n.links[k] = newNode(nil)
		if org != nil {
			*n.links[k] = *org
		}
		n.hash = nil
		n = n.links[k]
	}
	n.value = val
	n.hash = nil
	return nil
}

// Commit ...
func (s *Stage) Commit(head *Commit, meta CommitMeta) (*Commit, error) {
	c := NewCommit(head, meta, s.MerkleTrie)

	s.MerkleTrie = New()
	s.mu = c.MerkleTrie.mu
	s.root = newNode(nil)
	*s.root = *c.MerkleTrie.root
	return c, nil
}

// CommitVisit ...
type CommitVisit func(c *Commit)

// Log ...
func Log(commit *Commit, visit CommitVisit) {
	for commit != nil {
		visit(commit)
		commit = commit.Prev
	}
}
