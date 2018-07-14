package memento

import (
	"errors"
)

var (
	// ErrCommandStackEmpty is returned when the command stack is empty.
	ErrCommandStackEmpty = errors.New("command stack empty")
)

// Command specifies the interface of a command for the Memento.
type Command interface {
	Execute()
	Undo()
}

// CommandList is a list of command.
type CommandList []Command

func (l CommandList) empty() bool                { return len(l) == 0 }
func (l CommandList) top() Command               { return l[len(l)-1] }
func (l CommandList) pop() CommandList           { return l[:len(l)-1] }
func (l CommandList) push(c Command) CommandList { return append(l, c) }

// CommandStack is stack of command list.
type CommandStack []CommandList

func (s CommandStack) empty() bool                     { return len(s) == 0 }
func (s CommandStack) top() CommandList                { return s[len(s)-1] }
func (s CommandStack) pop() CommandStack               { return s[:len(s)-1] }
func (s CommandStack) push(l CommandList) CommandStack { return append(s, l) }

// Memento implements functionality for state to be undo and redo.
type Memento struct {
	// Executed but not yet commited command list.
	executed CommandList

	// A stack of command list that have been commited.
	commited CommandStack

	// A stack of commited command list that have been undone.
	undone CommandStack
}

// Executed returns a list of executed command that have not been commited.
func (m *Memento) Executed() CommandList { return m.executed }

// Commited returns the stack of command liist that have been commited.
func (m *Memento) Commited() CommandStack { return m.commited }

// Undone returns the stack of command list that have been undone.
func (m *Memento) Undone() CommandStack { return m.undone }

// Execute executes a command and appends it to the Executed command list.
// Any command list on the Undone will discarded, and can no longer be redone.
func (m *Memento) Execute(c Command) error {
	m.executed = m.executed.push(c)
	c.Execute()
	m.undone = nil
	return nil
}

// Commit commits the Executed command list to the Commited Stack, and empty the Executed List.
func (m *Memento) Commit() {
	m.commited = m.commited.push(m.executed)
	m.executed = nil
	m.undone = nil
}

// Undo undos the most recent command list on the Commited stack, and moves it to the Undone Stack.
func (m *Memento) Undo() error {
	if m.commited.empty() {
		return ErrCommandStackEmpty
	}
	m.commited, m.undone = process(m.commited, m.undone, true)
	return nil
}

// Redo redos the most recent command list on the Undone Stack, and moves it back to the Commited Stack.
func (m *Memento) Redo() error {
	if m.undone.empty() {
		return ErrCommandStackEmpty
	}
	m.undone, m.commited = process(m.undone, m.commited, false)
	return nil
}

// RollbackExecuted undos commands on the Executed list, and empty the list.
func (m *Memento) RollbackExecuted() {
	for !m.executed.empty() {
		m.executed.top().Undo()
		m.executed = m.executed.pop()
	}
	m.executed = nil
}

func process(a, b CommandStack, undo bool) (CommandStack, CommandStack) {
	processed := CommandList{}
	for cmds := a.top(); !cmds.empty(); cmds = cmds.pop() {
		if undo {
			cmds.top().Undo()
		} else {
			cmds.top().Execute()
		}
		processed = processed.push(cmds.top())
	}
	return a.pop(), b.push(processed)
}
