package trie

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func Test_update(t *testing.T) {
	res1 := buildNode(newNode(nil), pairs1())
	tests := []struct {
		name string
		res  *node
		exp  *node
	}{
		{"test1", res1, unprunedNode()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.res, tt.exp) {
				traverse(tt.res, Key{}, dump)
				fmt.Println("")
				traverse(tt.exp, Key{}, dump)
				t.Errorf("update() = %v, want %v", tt.res, tt.exp)
			}
		})
	}
}

func Test_nullify(t *testing.T) {
	tests := []struct {
		name string
		res  *node
		exp  *node
	}{
		{"test1", prune(unprunedNode()), prunedNode()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.res, tt.exp) {
				t.Errorf("traverse() = %v, want %v", tt.res, tt.exp)
			}
		})
	}
}

func Test_traverse(t *testing.T) {
	res1 := []pair{}
	fn := func(prefix Key, value Value) error {
		res1 = append(res1, pair{string(prefix), value})
		return nil
	}
	traverse(unprunedNode(), Key{}, fn)
	exp1 := []pair{
		{"a", nil},
		{"al", nil},
		{"ale", nil},
		{"alex", nil},
		{"b", nil},
		{"bo", nil},
		{"bob", strValue("cat")},
		{"t", nil},
		{"te", nil},
		{"ted", strValue("dog")},
		{"tedd", nil},
		{"teddy", strValue("bear")},
		{"tes", nil},
		{"tess", strValue("dolphin")},
	}

	res2 := []pair{}
	fn2 := func(prefix Key, value Value) error {
		res2 = append(res2, pair{string(prefix), value})
		return nil
	}
	traverse(prunedNode(), Key{}, fn2)
	exp2 := []pair{
		{"b", nil},
		{"bo", nil},
		{"bob", strValue("cat")},
		{"t", nil},
		{"te", nil},
		{"ted", strValue("dog")},
		{"tedd", nil},
		{"teddy", strValue("bear")},
		{"tes", nil},
		{"tess", strValue("dolphin")},
	}

	tests := []struct {
		name string
		res  []pair
		exp  []pair
	}{
		{"test1", res1, exp1},
		{"test2", res2, exp2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.res, tt.exp) {
				t.Errorf("traverse() = %v, want %v", tt.res, tt.exp)
			}
		})
	}
}

func Test_merkle(t *testing.T) {
	n1 := buildNode(newNode(nil), pairs1())
	// n2 := func() *node {
	// 	p1 := wire.OutPoint{Hash: *newHashFromStr("627ecfee2110b28fbc4b012944cadf66a72f394ad9fa9bb18fec30789e26c9ac"), Index: 0}
	// 	p2 := wire.OutPoint{Hash: *newHashFromStr("c31bd469112abf04930879c6b6007d2b23224e042785d404bbeff1932dd94880"), Index: 0}

	// 	n1 := claim.NewNode(&claim.Claim{OutPoint: p1, ClaimID: nil, Amount: 50, Height: 100, ValidAtHeight: 200})
	// 	n2 := claim.NewNode(&claim.Claim{OutPoint: p2, ClaimID: nil, Amount: 50, Height: 100, ValidAtHeight: 200})

	// 	pairs := []pair{
	// 		{"test", n1},
	// 		{"test2", n2},
	// 	}
	// 	return buildNode(newNode(nil), pairs)
	// }()
	tests := []struct {
		name string
		n    *node
		want *chainhash.Hash
	}{
		{"test1", n1, newHashFromStr("c2fdce68a30e3cabf6efb3b7ebfd32afdaf09f9ebd062743fe91e181f682252b")},
		// {"test2", n2, newHashFromStr("71c7b8d35b9a3d7ad9a1272b68972979bbd18589f1efe6f27b0bf260a6ba78fa")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := merkle(tt.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("merkle() = %v, want %v", got, tt.want)
			}
		})
	}
}
