package claimtrie

import (
	"reflect"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"

	"github.com/stretchr/testify/assert"
)

func newHash(s string) *chainhash.Hash {
	h, _ := chainhash.NewHashFromStr(s)
	return h
}

func newOutPoint(idx int) *wire.OutPoint {
	// var h chainhash.Hash
	// if _, err := rand.Read(h[:]); err != nil {
	// 	return nil
	// }
	// return wire.NewOutPoint(&h, uint32(idx))
	return wire.NewOutPoint(new(chainhash.Hash), uint32(idx))
}

func Test_calNodeHash(t *testing.T) {
	type args struct {
		op wire.OutPoint
		h  Height
	}
	tests := []struct {
		name string
		args args
		want chainhash.Hash
	}{
		{
			name: "0-1",
			args: args{op: wire.OutPoint{Hash: *newHash("c73232a755bf015f22eaa611b283ff38100f2a23fb6222e86eca363452ba0c51"), Index: 0}, h: 0},
			want: *newHash("48a312fc5141ad648cb5dca99eaf221f7b1bc4d2fc559e1cde4664a46d8688a4"),
		},
		{
			name: "0-2",
			args: args{op: wire.OutPoint{Hash: *newHash("71c7b8d35b9a3d7ad9a1272b68972979bbd18589f1efe6f27b0bf260a6ba78fa"), Index: 1}, h: 1},
			want: *newHash("9132cc5ff95ae67bee79281438e7d00c25c9ec8b526174eb267c1b63a55be67c"),
		},
		{
			name: "0-3",
			args: args{op: wire.OutPoint{Hash: *newHash("c4fc0e2ad56562a636a0a237a96a5f250ef53495c2cb5edd531f087a8de83722"), Index: 0x12345678}, h: 0x87654321},
			want: *newHash("023c73b8c9179ffcd75bd0f2ed9784aab2a62647585f4b38e4af1d59cf0665d2"),
		},
		{
			name: "0-4",
			args: args{op: wire.OutPoint{Hash: *newHash("baf52472bd7da19fe1e35116cfb3bd180d8770ffbe3ae9243df1fb58a14b0975"), Index: 0x11223344}, h: 0x88776655},
			want: *newHash("6a2d40f37cb2afea3b38dea24e1532e18cade5d1dc9c2f8bd635aca2bc4ac980"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calNodeHash(tt.args.op, tt.args.h); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calNodeHash() = %X, want %X", got, tt.want)
			}
		})
	}
}

var c1, c2, c3, c4, c5, c6, c7, c8, c9, c10 *Claim
var s1, s2, s3, s4, s5, s6, s7, s8, s9, s10 *Support

func Test_History1(t *testing.T) {

	proportionalDelayFactor = 1
	n := NewNode()

	// no competing bids
	test1 := func() {
		c1, _ = n.addClaim(*newOutPoint(1), 1)
		n.IncrementBlock(1)
		assert.Equal(t, c1, n.BestClaim())

		n.DecrementBlock(1)
		assert.Nil(t, n.BestClaim())
	}

	// there is a competing bid inserted same height
	test2 := func() {
		n.addClaim(*newOutPoint(2), 1)
		c3, _ = n.addClaim(*newOutPoint(3), 2)
		n.IncrementBlock(1)
		assert.Equal(t, c3, n.BestClaim())

		n.DecrementBlock(1)
		assert.Nil(t, n.BestClaim())

	}
	// make two claims , one older
	test3 := func() {
		c4, _ = n.addClaim(*newOutPoint(4), 1)
		n.IncrementBlock(1)
		assert.Equal(t, c4, n.BestClaim())
		n.addClaim(*newOutPoint(5), 1)
		n.IncrementBlock(1)
		assert.Equal(t, c4, n.BestClaim())
		n.IncrementBlock(1)
		assert.Equal(t, c4, n.BestClaim())
		n.DecrementBlock(1)
		assert.Equal(t, c4, n.BestClaim())
		n.DecrementBlock(1)

		assert.Equal(t, c4, n.BestClaim())
		n.DecrementBlock(1)
		assert.Nil(t, n.BestClaim())
	}

	// check claim takeover
	test4 := func() {
		c6, _ = n.addClaim(*newOutPoint(6), 1)
		n.IncrementBlock(10)
		assert.Equal(t, c6, n.BestClaim())

		c7, _ = n.addClaim(*newOutPoint(7), 2)
		n.IncrementBlock(1)
		assert.Equal(t, c6, n.BestClaim())
		n.IncrementBlock(10)
		assert.Equal(t, c7, n.BestClaim())

		n.DecrementBlock(10)
		assert.Equal(t, c6, n.BestClaim())
		n.DecrementBlock(10)
		assert.Equal(t, c6, n.BestClaim())
		n.DecrementBlock(1)
		assert.Nil(t, n.BestClaim())
	}

	// spending winning claim will make losing active claim winner
	test5 := func() {
		c1, _ = n.addClaim(*newOutPoint(1), 2)
		c2, _ = n.addClaim(*newOutPoint(2), 1)
		n.IncrementBlock(1)
		assert.Equal(t, c1, n.BestClaim())
		n.removeClaim(c1.op)
		n.IncrementBlock(1)
		assert.Equal(t, c2, n.BestClaim())

		n.DecrementBlock(1)
		assert.Equal(t, c1, n.BestClaim())
		n.DecrementBlock(1)
		assert.Nil(t, n.BestClaim())
	}

	// spending winning claim will make inactive claim winner
	test6 := func() {
		c3, _ = n.addClaim(*newOutPoint(3), 2)
		n.IncrementBlock(10)
		assert.Equal(t, c3, n.BestClaim())

		c4, _ = n.addClaim(*newOutPoint(4), 2)
		n.IncrementBlock(1)
		assert.Equal(t, c3, n.BestClaim())
		n.removeClaim(c3.op)
		n.IncrementBlock(1)
		assert.Equal(t, c4, n.BestClaim())

		n.DecrementBlock(1)
		assert.Equal(t, c3, n.BestClaim())
		n.DecrementBlock(1)
		assert.Equal(t, c3, n.BestClaim())
		n.DecrementBlock(10)
		assert.Nil(t, n.BestClaim())
	}

	// spending winning claim will empty out claim trie
	test7 := func() {
		c5, _ = n.addClaim(*newOutPoint(5), 2)
		n.IncrementBlock(1)
		assert.Equal(t, c5, n.BestClaim())
		n.removeClaim(c5.op)
		n.IncrementBlock(1)
		assert.NotEqual(t, c5, n.BestClaim())

		n.DecrementBlock(1)
		assert.Equal(t, c5, n.BestClaim())
		n.DecrementBlock(1)
		assert.Nil(t, n.BestClaim())
	}

	// check claim with more support wins
	test8 := func() {
		c1, _ = n.addClaim(*newOutPoint(1), 2)
		c2, _ = n.addClaim(*newOutPoint(2), 1)
		s1, _ = n.addSupport(*newOutPoint(11), 1, c1.id)
		s2, _ = n.addSupport(*newOutPoint(12), 10, c2.id)
		n.IncrementBlock(1)
		assert.Equal(t, c2, n.BestClaim())
		assert.Equal(t, Amount(11), n.BestClaim().effAmt)
		n.DecrementBlock(1)
		assert.Nil(t, n.BestClaim())
	}
	// check support delay
	test9 := func() {
		c3, _ = n.addClaim(*newOutPoint(3), 1)
		c4, _ = n.addClaim(*newOutPoint(4), 2)
		n.IncrementBlock(10)
		assert.Equal(t, c4, n.BestClaim())
		assert.Equal(t, Amount(2), n.BestClaim().effAmt)
		s4, _ = n.addSupport(*newOutPoint(14), 10, c3.id)
		n.IncrementBlock(10)
		assert.Equal(t, c4, n.BestClaim())
		assert.Equal(t, Amount(2), n.BestClaim().effAmt)
		n.IncrementBlock(1)
		assert.Equal(t, c3, n.BestClaim())
		assert.Equal(t, Amount(11), n.BestClaim().effAmt)

		n.DecrementBlock(1)
		assert.Equal(t, c4, n.BestClaim())
		assert.Equal(t, Amount(2), n.BestClaim().effAmt)
		n.DecrementBlock(10)
		assert.Equal(t, c4, n.BestClaim())
		assert.Equal(t, Amount(2), n.BestClaim().effAmt)
		n.DecrementBlock(10)
		assert.Nil(t, n.BestClaim())
	}

	// supporting and abandoing on the same block will cause it to crash
	test10 := func() {
		c1, _ = n.addClaim(*newOutPoint(1), 2)
		n.IncrementBlock(1)
		s1, _ = n.addSupport(*newOutPoint(11), 1, c1.id)
		n.removeClaim(c1.op)
		n.IncrementBlock(1)
		assert.NotEqual(t, c1, n.BestClaim())

		n.DecrementBlock(1)
		assert.Equal(t, c1, n.BestClaim())
		n.DecrementBlock(1)
		assert.Nil(t, n.BestClaim())
	}

	// support on abandon2
	test11 := func() {
		c1, _ = n.addClaim(*newOutPoint(1), 2)
		s1, _ = n.addSupport(*newOutPoint(11), 1, c1.id)
		n.IncrementBlock(1)
		assert.Equal(t, c1, n.BestClaim())

		//abandoning a support and abandoing claim on the same block will cause it to crash
		n.removeClaim(c1.op)
		n.removeSupport(s1.op)
		n.IncrementBlock(1)
		assert.Nil(t, n.BestClaim())

		n.DecrementBlock(1)
		assert.Equal(t, c1, n.BestClaim())
		n.DecrementBlock(1)
		assert.Nil(t, n.BestClaim())
	}
	tests := []func(){
		test1,
		test2,
		test3,
		test4,
		test5,
		test6,
		test7,
		test8,
		test9,
		test10,
		test11,
	}
	for _, tt := range tests {
		tt()
	}
	_ = []func(){test1, test2, test3, test4, test5, test6, test7, test8, test9, test10, test11}
}
