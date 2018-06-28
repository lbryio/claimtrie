package claimtrie

import (
	"reflect"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

func newHash(s string) *chainhash.Hash {
	h, _ := chainhash.NewHashFromStr(s)
	return h
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
			name: "test1",
			args: args{op: wire.OutPoint{Hash: *newHash("c73232a755bf015f22eaa611b283ff38100f2a23fb6222e86eca363452ba0c51"), Index: 0}, h: 0},
			want: *newHash("48a312fc5141ad648cb5dca99eaf221f7b1bc4d2fc559e1cde4664a46d8688a4"),
		},
		{
			name: "test2",
			args: args{op: wire.OutPoint{Hash: *newHash("71c7b8d35b9a3d7ad9a1272b68972979bbd18589f1efe6f27b0bf260a6ba78fa"), Index: 1}, h: 1},
			want: *newHash("9132cc5ff95ae67bee79281438e7d00c25c9ec8b526174eb267c1b63a55be67c"),
		},
		{
			name: "test3",
			args: args{op: wire.OutPoint{Hash: *newHash("c4fc0e2ad56562a636a0a237a96a5f250ef53495c2cb5edd531f087a8de83722"), Index: 0x12345678}, h: 0x87654321},
			want: *newHash("023c73b8c9179ffcd75bd0f2ed9784aab2a62647585f4b38e4af1d59cf0665d2"),
		},
		{
			name: "test4",
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
