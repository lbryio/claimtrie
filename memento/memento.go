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
	cmds := m.stack[len(m.stack)-1]
	m.stack = m.stack[:len(m.stack)-1]
	for i := len(cmds) - 1; i >= 0; i-- {
		cmds[i].Undo()
	}
}
