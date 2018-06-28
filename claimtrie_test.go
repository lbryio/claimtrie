package claimtrie

import (
	"testing"

	"github.com/btcsuite/btcd/wire"

	"github.com/lbryio/merkletrie"
)

func TestClaimTrie_AddClaim(t *testing.T) {
	type fields struct {
		stg *merkletrie.Stage
	}
	type args struct {
		name     string
		outPoint wire.OutPoint
		value    Amount
		height   Height
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct := &ClaimTrie{
				stg: tt.fields.stg,
			}
			if err := ct.AddClaim(tt.args.name, tt.args.outPoint, tt.args.value, tt.args.height); (err != nil) != tt.wantErr {
				t.Errorf("ClaimTrie.AddClaim() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
