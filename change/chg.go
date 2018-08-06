package change

import (
	"fmt"

	"github.com/lbryio/claimtrie/claim"
)

// Cmd defines the type of Change.
type Cmd int

// The list of command currently supported.
const (
	AddClaim Cmd = 1 << iota
	SpendClaim
	UpdateClaim
	AddSupport
	SpendSupport
)

var names = map[Cmd]string{
	AddClaim:     "+C",
	SpendClaim:   "-C",
	UpdateClaim:  "+U",
	AddSupport:   "+S",
	SpendSupport: "-S",
}

// Change represent a record of changes to the node of Name at Height.
type Change struct {
	Height claim.Height
	Cmd    Cmd
	Name   string
	OP     claim.OutPoint
	Amt    claim.Amount
	ID     claim.ID
	Value  []byte
}

func (c Change) String() string {
	return fmt.Sprintf("%6d %s %s %s %12d [%s]", c.Height, names[c.Cmd], c.OP, c.ID, c.Amt, c.Name)
}

// New returns a Change initialized with Cmd.
func New(cmd Cmd) *Change {
	return &Change{Cmd: cmd}
}

// SetName sets name to the Change.
func (c *Change) SetName(name string) *Change { c.Name = name; return c }

// SetHeight sets height to the Change.
func (c *Change) SetHeight(h claim.Height) *Change { c.Height = h; return c }

// SetOP sets OP to the Change.
func (c *Change) SetOP(op claim.OutPoint) *Change { c.OP = op; return c }

// SetAmt sets amt to the Change.
func (c *Change) SetAmt(amt claim.Amount) *Change { c.Amt = amt; return c }

// SetID sets id to the Change.
func (c *Change) SetID(id claim.ID) *Change { c.ID = id; return c }

// SetValue sets value to the Change.
func (c *Change) SetValue(v []byte) *Change { c.Value = v; return c }
