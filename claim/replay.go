package claim

import (
	"fmt"

	"github.com/pkg/errors"
)

type cmd int

// ...
const (
	CmdAddClaim cmd = 1 << iota
	CmdSpendClaim
	CmdUpdateClaim
	CmdAddSupport
	CmdSpendSupport
)

var cmdName = map[cmd]string{
	CmdAddClaim:     "+C",
	CmdSpendClaim:   "-C",
	CmdUpdateClaim:  "+U",
	CmdAddSupport:   "+S",
	CmdSpendSupport: "-S",
}

// Cmd ...
type Cmd struct {
	Height Height
	Cmd    cmd
	Name   string
	OP     OutPoint
	Amt    Amount
	ID     ID
	Value  []byte
}

func (c Cmd) String() string {
	return fmt.Sprintf("%6d %s %s %s %12d [%s]", c.Height, cmdName[c.Cmd], c.OP, c.ID, c.Amt, c.Name)
}

func (n *Node) record(c cmd, op OutPoint, amt Amount, id ID) *Cmd {
	r := &Cmd{Height: n.height + 1, Name: n.name, Cmd: c, OP: op, Amt: amt, ID: id}
	n.records = append(n.records, r)
	return r
}

// AdjustTo increments current height until it reaches the specific height.
func (n *Node) AdjustTo(h Height) error {
	if h < n.height {
		return errors.Wrapf(ErrInvalidHeight, "adjust n.height: %d > %d", n.height, h)
	}
	if h == n.height {
		return nil
	}
	for n.height < h {
		n.height++
		n.bid()
		next := n.NextUpdate()
		if next > h {
			n.height = h
			break
		}
		n.height = next
	}
	n.bid()
	return nil
}

// Recall ...
func (n *Node) Recall(h Height) error {
	if h >= n.height {
		return errors.Wrapf(ErrInvalidHeight, "h: %d >= n.height: %d", h, n.height)
	}
	fmt.Printf("n.Recall from %d to %d\n", n.height, h)
	err := n.replay(h, false)
	return errors.Wrapf(err, "reply(%d, false)", h)
}

// Reset rests ...
func (n *Node) Reset(h Height) error {
	if h > n.height {
		return nil
	}
	fmt.Printf("n.Reset from %d to %d\n", n.height, h)
	err := n.replay(h, true)
	return errors.Wrapf(err, "reply(%d, true)", h)
}

func (n *Node) replay(h Height, truncate bool) error {
	fmt.Printf("replay %s from %d to %d:\n", n.name, n.height, h)
	backup := n.records
	*n = *NewNode(n.name)
	n.records = backup

	var i int
	var r *Cmd
	for i < len(n.records) {
		r = n.records[i]
		if n.height == r.Height-1 {
			if err := n.execute(r); err != nil {
				return err
			}
			i++
			continue
		}
		n.height++
		n.bid()
		if n.height == h {
			break
		}
	}
	if truncate {
		n.records = n.records[:i]
	}
	return nil
}

func (n *Node) execute(c *Cmd) error {
	var err error
	switch c.Cmd {
	case CmdAddClaim:
		err = n.addClaim(c.OP, c.Amt)
	case CmdSpendClaim:
		err = n.spendClaim(c.OP)
	case CmdUpdateClaim:
		err = n.updateClaim(c.OP, c.Amt, c.ID)
	case CmdAddSupport:
		err = n.addSupport(c.OP, c.Amt, c.ID)
	case CmdSpendSupport:
		err = n.spendSupport(c.OP)
	}
	return errors.Wrapf(err, "cmd %s", c)
}
