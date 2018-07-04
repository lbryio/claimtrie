package claimtrie

import (
	"crypto/rand"
	"reflect"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

func newHash(s string) *chainhash.Hash {
	h, _ := chainhash.NewHashFromStr(s)
	return h
}

func newOutPoint(idx int) *wire.OutPoint {
	var h chainhash.Hash
	if _, err := rand.Read(h[:]); err != nil {
		return nil
	}
	return wire.NewOutPoint(&h, uint32(idx))
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
func Test_BestClaim(t *testing.T) {
	cA := newClaim(*newOutPoint(1), 0, 0)
	cB := newClaim(*newOutPoint(2), 0, 0)
	cC := newClaim(*newOutPoint(3), 0, 0)
	cD := newClaim(*newOutPoint(4), 0, 0)

	s1 := newSupport(*newOutPoint(91), 0, 0, cA.id)

	var n *node
	type operation int
	const (
		opReset = 1 << iota
		opAddClaim
		opRemoveClaim
		opAddSupport
		opRemoveSupport
		opCheck
	)
	tests := []struct {
		name    string
		op      operation
		claim   *claim
		support *support
		amount  Amount
		curr    Height
		want    *claim
	}{
		{name: "0-0", op: opReset},
		{name: "0-1", op: opAddClaim, claim: cA, amount: 10, curr: 13, want: cA},                  // A(10) is controlling
		{name: "0-2", op: opAddClaim, claim: cB, amount: 20, curr: 1001, want: cA},                // A(10) is controlling, B(20) is accepted. Act(B) = 1001 + (1001-13)/32 = 1031
		{name: "0-3", op: opAddSupport, claim: cA, support: s1, amount: 14, curr: 1010, want: cA}, // A(10+14) is controlling, B(20) is accepted.
		{name: "0-4", op: opAddClaim, claim: cC, amount: 50, curr: 1020, want: cA},                // A(10+14) is controlling, B(20) is accepted, C(50) is accepted. Act(C) = 1020 + (1020-13)/32 = 1051
		{name: "0-5", op: opCheck, curr: 1031, want: cA},                                          // A(10+14) is controlling, B(20) is active, C(50) is accepted.
		{name: "0-6", op: opAddClaim, claim: cD, amount: 300, curr: 1040, want: cA},               // A(10+14) is controlling, B(20) is active, C(50) is accepted, D(300) is accepted. Act(C) = 1040 + (1040-13)/32 = 1072
		{name: "0-7", op: opCheck, curr: 1051, want: cD},                                          // A(10+14) is active, B(20) is active, C(50) is active, D(300) is controlling.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			switch tt.op {
			case opReset:
				n = newNode()
			case opAddClaim:
				tt.claim.amt = tt.amount
				tt.claim.accepted = tt.curr
				err = n.addClaim(tt.claim)
			case opRemoveClaim:
				err = n.removeClaim(tt.claim.op)
			case opAddSupport:
				tt.support.accepted = tt.curr
				tt.support.amt = tt.amount
				tt.support.supportedID = tt.claim.id
				err = n.addSupport(tt.support)
			case opRemoveSupport:
			}
			if err != nil {
				t.Errorf("BestClaim() failed, err: %s", err)
			}
			n.updateBestClaim(tt.curr)
			got := n.bestClaim
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BestClaim() = %d, want %d", got.op.Index, tt.want.op.Index)
			}
		})

	}
}
