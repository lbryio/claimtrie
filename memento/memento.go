package memento

// Command ...
type Command interface {
	Execute()
	Undo()
}

type commands []Command

// Memento ...
type Memento struct {
	stack []commands
	cmds  commands
}

// Execute ...
func (m *Memento) Execute(cmd Command) {
	m.cmds = append(m.cmds, cmd)
	cmd.Execute()
}

// Commit ...
func (m *Memento) Commit() {
	m.stack = append(m.stack, m.cmds)
	m.cmds = nil
}

// Rollback ...
func (m *Memento) Rollback() {
	n := len(m.stack)
	cmds := m.stack[n-1]
	m.stack = m.stack[:n-1]
	m.cmds = nil
	for i := len(cmds) - 1; i >= 0; i-- {
		cmds[i].Undo()
	}
}

// RollbackUncommited ...
func (m *Memento) RollbackUncommited() {
	for i := len(m.cmds) - 1; i >= 0; i-- {
		m.cmds[i].Undo()
	}
	m.cmds = nil
}
