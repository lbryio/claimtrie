package claimnode

import "fmt"

var (
	// ErrInvalidHeight is returned when the height is invalid.
	ErrInvalidHeight = fmt.Errorf("invalid height")

	// ErrNotFound is returned when the Claim or Support is not found.
	ErrNotFound = fmt.Errorf("not found")

	// ErrDuplicate is returned when the Claim or Support already exists in the node.
	ErrDuplicate = fmt.Errorf("duplicate")
)
