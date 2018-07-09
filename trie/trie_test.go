package trie

import (
	"reflect"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

func TestTrie_Update(t *testing.T) {
	mt := buildTrie(New(), pairs1())
	m := buildMap(newMap(), pairs1())

	for k := range m {
		v, _ := mt.Get(Key(k))
		if m[k] != v {
			t.Errorf("exp %s got %s", m[k], v)
		}
	}
}

func TestTrie_Hash(t *testing.T) {
	tr1 := buildTrie(New(), pairs1())
	// tr2 := func() *MerkleTrie {
	// 	p1 := wire.OutPoint{Hash: *newHashFromStr("627ecfee2110b28fbc4b012944cadf66a72f394ad9fa9bb18fec30789e26c9ac"), Index: 0}
	// 	p2 := wire.OutPoint{Hash: *newHashFromStr("c31bd469112abf04930879c6b6007d2b23224e042785d404bbeff1932dd94880"), Index: 0}

	// 	n1 := claim.NewNode(&claim.Claim{OutPoint: p1, ClaimID: nil, Amount: 50, Height: 100, ValidAtHeight: 200})
	// 	n2 := claim.NewNode(&claim.Claim{OutPoint: p2, ClaimID: nil, Amount: 50, Height: 100, ValidAtHeight: 200})

	// 	pairs := []pair{
	// 		{"test", n1},
	// 		{"test2", n2},
	// 	}
	// 	return buildTrie(New(), pairs)
	// }()
	tests := []struct {
		name string
		mt   *MerkleTrie
		want chainhash.Hash
	}{
		{"empty", New(), *newHashFromStr("0000000000000000000000000000000000000000000000000000000000000001")},
		{"test1", tr1, *newHashFromStr("c2fdce68a30e3cabf6efb3b7ebfd32afdaf09f9ebd062743fe91e181f682252b")},
		// {"test2", tr2, *newHashFromStr("71c7b8d35b9a3d7ad9a1272b68972979bbd18589f1efe6f27b0bf260a6ba78fa")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := tt.mt
			if got := mt.MerkleHash(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("trie.MerkleHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrie_Size(t *testing.T) {
	mt1 := buildTrie(New(), pairs1())
	map1 := buildMap(newMap(), pairs1())

	tests := []struct {
		name string
		mt   *MerkleTrie
		want int
	}{
		{"test1", mt1, len(map1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := tt.mt
			if got := mt.Size(); got != tt.want {
				t.Errorf("trie.Size() = %v, want %v", got, tt.want)
			}
		})
	}
}
