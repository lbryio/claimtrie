package claimtrie

import (
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// pending ...
func TestClaimTrie_Commit(t *testing.T) {
	ct := New()

	tests := []struct {
		name string
		curr Height
		amt  Amount
		want chainhash.Hash
	}{
		{name: "0-0", curr: 5, amt: 11},
		{name: "0-0", curr: 6, amt: 10},
		{name: "0-0", curr: 7, amt: 14},
		{name: "0-0", curr: 8, amt: 18},
		{name: "0-0", curr: 100, amt: 0},
		{name: "0-0", curr: 101, amt: 30},
		{name: "0-0", curr: 102, amt: 00},
		{name: "0-0", curr: 103, amt: 00},
		{name: "0-0", curr: 104, amt: 00},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.amt != 0 {
				ct.AddClaim("HELLO", *newOutPoint(0), tt.amt, tt.curr)
			}
			ct.Commit(tt.curr)
			// fmt.Printf("ct.Merkle[%2d]: %s, amt: %d\n", ct.BestBlock(), ct.MerkleHash(), tt.amt)
		})
	}
}
