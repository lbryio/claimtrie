package claimnode

import (
	"reflect"
	"testing"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/assert"

	"github.com/lbryio/claimtrie/claim"
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
		h  claim.Height
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

// The example on the [whitepaper](https://beta.lbry.tech/whitepaper.html)
func Test_History0(t *testing.T) {

	proportionalDelayFactor = 32
	n := NewNode()

	n.AdjustTo(12)
	claimA, _ := n.AddClaim(*newOutPoint(1), 10) // Claim A
	n.AdjustTo(13)
	// Claim A for 10LBC is accepted. It is the first claim, so it immediately becomes active and controlling.
	assert.Equal(t, claimA, n.BestClaim()) // A(10) is controlling.

	n.AdjustTo(1000)
	n.AddClaim(*newOutPoint(2), 20) // Claim B
	n.AdjustTo(1001)
	// Claim B for 20LBC is accepted. It’s activation height is 1001+min(4032,floor(1001−1332))=1001+30=1031
	assert.Equal(t, claimA, n.BestClaim()) // A(10) is controlling, B(20) is accepted.

	n.AdjustTo(1009)
	n.AddSupport(*newOutPoint(1001), 14, claimA.ID) // Support X
	n.AdjustTo(1010)
	// Support X for 14LBC for claim A is accepted. Since it is a support for the controlling claim, it activates immediately.
	assert.Equal(t, claimA, n.BestClaim()) // A(10+14) is controlling, B(20) is accepted.

	n.AdjustTo(1019)
	n.AddClaim(*newOutPoint(3), 50) // Claim C
	n.AdjustTo(1020)
	// Claim C for 50LBC is accepted. The activation height is 1020+min(4032,floor(1020−1332))=1020+31=1051
	assert.Equal(t, claimA, n.BestClaim()) // A(10+14) is controlling, B(20) is accepted, C(50) is accepted.

	n.AdjustTo(1031)
	// Claim B activates. It has 20LBC, while claim A has 24LBC (10 original + 14 from support X). There is no takeover, and claim A remains controlling.
	assert.Equal(t, claimA, n.BestClaim()) // A(10+14) is controlling, B(20) is active, C(50) is accepted.

	n.AdjustTo(1039)
	claimD, _ := n.AddClaim(*newOutPoint(4), 300) // Claim D
	n.AdjustTo(1040)
	// Claim D for 300LBC is accepted. The activation height is 1040+min(4032,floor(1040−1332))=1040+32=1072
	assert.Equal(t, claimA, n.BestClaim()) //A(10+14) is controlling, B(20) is active, C(50) is accepted, D(300) is accepted.

	n.AdjustTo(1051)
	// Claim C activates. It has 50LBC, while claim A has 24LBC, so a takeover is initiated.
	// The takeover height for this name is set to 1051, and therefore the activation delay for all the claims becomes min(4032, floor(1051−1051/32)) = 0.
	// All the claims become active.
	// The totals for each claim are recalculated, and claim D becomes controlling because it has the highest total.
	assert.Equal(t, claimD, n.BestClaim()) // A(10+14) is active, B(20) is active, C(50) is active, D(300) is controlling.

	// The following are the corresponding commands in the CLI to reproduce the test.
	// Note that when a Toakeover happens, the {TakeoverHeight, BestClaim(OutPoint:Indx)} is pushed to the BestClaims stack.
	// This provides sufficient info to update and backtracking the BestClaim at any height.

	// claimtrie > c -ht 12
	// claimtrie > ac -a 10
	// claimtrie > c -ht 13
	// claimtrie > s
	//
	// <ClaimTrie Height 13>
	// Hello   :  Height 13, 6f5970c9c13f00c77054d98e5b2c50b1a1bb723d91676cc03f984fac763ec6c3 BestClaims: {13, 31},
	//
	//   C-2f9b2ca28b30c97122de584e8d784e784bc7bdfb43b3b4bf9de69fc31196571f:31  amt: 10   effamt: 10   accepted: 13   active: 13   id: ae8b3adc8c8b378c76eae12edf3878357b31c0eb  <B>

	// claimtrie > c -ht 1000
	// claimtrie > ac -a 20
	// claimtrie > c -ht 1001
	// claimtrie > s
	//
	// <ClaimTrie Height 1001>
	// Hello   :  Height 1000, 6f5970c9c13f00c77054d98e5b2c50b1a1bb723d91676cc03f984fac763ec6c3 BestClaims: {13, 31},
	//
	//   C-2f9b2ca28b30c97122de584e8d784e784bc7bdfb43b3b4bf9de69fc31196571f:31  amt: 10   effamt: 10   accepted: 13   active: 13   id: ae8b3adc8c8b378c76eae12edf3878357b31c0eb  <B>
	//   C-74fd7c15b445e7ac2de49a8058ad7dbb070ef31838345feff168dd40c2ef422e:46  amt: 20   effamt: 0    accepted: 1001  active: 1031  id: 2edc6338e9f3654f8b2b878817e029a2b2ecfa9e

	// claimtrie > c -ht 1009
	// claimtrie > as -a 14 -id ae8b3adc8c8b378c76eae12edf3878357b31c0eb
	// claimtrie > c -ht 1010
	// claimtrie > s
	//
	// <ClaimTrie Height 1010>
	// Hello   :  Height 1010, 6f5970c9c13f00c77054d98e5b2c50b1a1bb723d91676cc03f984fac763ec6c3 BestClaims: {13, 31},
	//
	//   C-2f9b2ca28b30c97122de584e8d784e784bc7bdfb43b3b4bf9de69fc31196571f:31  amt: 10   effamt: 24   accepted: 13   active: 13   id: ae8b3adc8c8b378c76eae12edf3878357b31c0eb  <B>
	//   C-74fd7c15b445e7ac2de49a8058ad7dbb070ef31838345feff168dd40c2ef422e:46  amt: 20   effamt: 0    accepted: 1001  active: 1031  id: 2edc6338e9f3654f8b2b878817e029a2b2ecfa9e
	//
	//   S-087acb4f22ab6eb2e6c827624deab7beb02c190056376e0b3a3c3546b79bf216:22  amt: 14                accepted: 1010  active: 1010  id: ae8b3adc8c8b378c76eae12edf3878357b31c0eb

	// claimtrie > c -ht 1019
	// claimtrie > ac -a 50
	// claimtrie > c -ht 1020
	// claimtrie > s
	//
	// <ClaimTrie Height 1020>
	// Hello   :  Height 1019, 6f5970c9c13f00c77054d98e5b2c50b1a1bb723d91676cc03f984fac763ec6c3 BestClaims: {13, 31},
	//
	//   C-2f9b2ca28b30c97122de584e8d784e784bc7bdfb43b3b4bf9de69fc31196571f:31  amt: 10   effamt: 24   accepted: 13   active: 13   id: ae8b3adc8c8b378c76eae12edf3878357b31c0eb  <B>
	//   C-74fd7c15b445e7ac2de49a8058ad7dbb070ef31838345feff168dd40c2ef422e:46  amt: 20   effamt: 0    accepted: 1001  active: 1031  id: 2edc6338e9f3654f8b2b878817e029a2b2ecfa9e
	//   C-864dc683ed1fbe3c072c2387bca7d74e60c2505f1987054fd70955a2e1c9490b:11  amt: 50   effamt: 0    accepted: 1020  active: 1051  id: 0f2ba15a5c66b978df4a27f52525680a6009185e
	//
	//   S-087acb4f22ab6eb2e6c827624deab7beb02c190056376e0b3a3c3546b79bf216:22  amt: 14                accepted: 1010  active: 1010  id: ae8b3adc8c8b378c76eae12edf3878357b31c0eb

	// claimtrie > c -ht 1031
	// claimtrie > s
	//
	// <ClaimTrie Height 1031>
	// Hello   :  Height 1031, 6f5970c9c13f00c77054d98e5b2c50b1a1bb723d91676cc03f984fac763ec6c3 BestClaims: {13, 31},
	//
	//   C-2f9b2ca28b30c97122de584e8d784e784bc7bdfb43b3b4bf9de69fc31196571f:31  amt: 10   effamt: 24   accepted: 13   active: 13   id: ae8b3adc8c8b378c76eae12edf3878357b31c0eb  <B>
	//   C-74fd7c15b445e7ac2de49a8058ad7dbb070ef31838345feff168dd40c2ef422e:46  amt: 20   effamt: 20   accepted: 1001  active: 1031  id: 2edc6338e9f3654f8b2b878817e029a2b2ecfa9e
	//   C-864dc683ed1fbe3c072c2387bca7d74e60c2505f1987054fd70955a2e1c9490b:11  amt: 50   effamt: 0    accepted: 1020  active: 1051  id: 0f2ba15a5c66b978df4a27f52525680a6009185e
	//
	//   S-087acb4f22ab6eb2e6c827624deab7beb02c190056376e0b3a3c3546b79bf216:22  amt: 14                accepted: 1010  active: 1010  id: ae8b3adc8c8b378c76eae12edf3878357b31c0eb

	// claimtrie > c -ht 1039
	// claimtrie > ac -a 300
	// claimtrie > c -ht 1040
	// claimtrie > s
	//
	// <ClaimTrie Height 1040>
	// Hello   :  Height 1039, 6f5970c9c13f00c77054d98e5b2c50b1a1bb723d91676cc03f984fac763ec6c3 BestClaims: {13, 31},
	//
	//   C-2f9b2ca28b30c97122de584e8d784e784bc7bdfb43b3b4bf9de69fc31196571f:31  amt: 10   effamt: 24   accepted: 13   active: 13   id: ae8b3adc8c8b378c76eae12edf3878357b31c0eb  <B>
	//   C-74fd7c15b445e7ac2de49a8058ad7dbb070ef31838345feff168dd40c2ef422e:46  amt: 20   effamt: 20   accepted: 1001  active: 1031  id: 2edc6338e9f3654f8b2b878817e029a2b2ecfa9e
	//   C-864dc683ed1fbe3c072c2387bca7d74e60c2505f1987054fd70955a2e1c9490b:11  amt: 50   effamt: 0    accepted: 1020  active: 1051  id: 0f2ba15a5c66b978df4a27f52525680a6009185e
	//   C-26292b27122d04d08fee4e4cc5a5f94681832204cc29d61039c09af9a5298d16:22  amt: 300  effamt: 0    accepted: 1040  active: 1072  id: 270496c0710e525156510e60e4be2ffa6fe2f507
	//
	//   S-087acb4f22ab6eb2e6c827624deab7beb02c190056376e0b3a3c3546b79bf216:22  amt: 14                accepted: 1010  active: 1010  id: ae8b3adc8c8b378c76eae12edf3878357b31c0eb

	// claimtrie > c -ht 1051
	// claimtrie > s
	//
	// <ClaimTrie Height 1051>
	// Hello   :  Height 1051, 68dff86c9450e3cf96570f31b6ad8f8d35ae0cbce6cdcb3761910e25a815ee8b BestClaims: {13, 31}, {1051, 22},
	//
	//   C-2f9b2ca28b30c97122de584e8d784e784bc7bdfb43b3b4bf9de69fc31196571f:31  amt: 10   effamt: 24   accepted: 13   active: 13   id: ae8b3adc8c8b378c76eae12edf3878357b31c0eb
	//   C-74fd7c15b445e7ac2de49a8058ad7dbb070ef31838345feff168dd40c2ef422e:46  amt: 20   effamt: 20   accepted: 1001  active: 1001  id: 2edc6338e9f3654f8b2b878817e029a2b2ecfa9e
	//   C-864dc683ed1fbe3c072c2387bca7d74e60c2505f1987054fd70955a2e1c9490b:11  amt: 50   effamt: 50   accepted: 1020  active: 1020  id: 0f2ba15a5c66b978df4a27f52525680a6009185e
	//   C-26292b27122d04d08fee4e4cc5a5f94681832204cc29d61039c09af9a5298d16:22  amt: 300  effamt: 300  accepted: 1040  active: 1040  id: 270496c0710e525156510e60e4be2ffa6fe2f507  <B>
	//
	//   S-087acb4f22ab6eb2e6c827624deab7beb02c190056376e0b3a3c3546b79bf216:22  amt: 14                accepted: 1010  active: 1010  id: ae8b3adc8c8b378c76eae12edf3878357b31c0eb
}

var c1, c2, c3, c4, c5, c6, c7, c8, c9, c10 *claim.Claim
var s1, s2, s3, s4, s5, s6, s7, s8, s9, s10 *claim.Support

func Test_History1(t *testing.T) {

	proportionalDelayFactor = 1
	n := NewNode()

	// no competing bids
	test1 := func() {
		c1, _ = n.AddClaim(*newOutPoint(1), 1)
		n.AdjustTo(1)
		assert.Equal(t, c1, n.BestClaim())

		n.AdjustTo(0)
		assert.Nil(t, n.BestClaim())
	}

	// there is a competing bid inserted same height
	test2 := func() {
		n.AddClaim(*newOutPoint(2), 1)
		c3, _ = n.AddClaim(*newOutPoint(3), 2)
		n.AdjustTo(1)
		assert.Equal(t, c3, n.BestClaim())

		n.AdjustTo(0)
		assert.Nil(t, n.BestClaim())

	}
	// make two claims , one older
	test3 := func() {
		c4, _ = n.AddClaim(*newOutPoint(4), 1)
		n.AdjustTo(1)
		assert.Equal(t, c4, n.BestClaim())
		n.AddClaim(*newOutPoint(5), 1)
		n.AdjustTo(2)
		assert.Equal(t, c4, n.BestClaim())
		n.AdjustTo(3)
		assert.Equal(t, c4, n.BestClaim())
		n.AdjustTo(2)
		assert.Equal(t, c4, n.BestClaim())
		n.AdjustTo(1)

		assert.Equal(t, c4, n.BestClaim())
		n.AdjustTo(0)
		assert.Nil(t, n.BestClaim())
	}

	// check claim takeover
	test4 := func() {
		c6, _ = n.AddClaim(*newOutPoint(6), 1)
		n.AdjustTo(10)
		assert.Equal(t, c6, n.BestClaim())

		c7, _ = n.AddClaim(*newOutPoint(7), 2)
		n.AdjustTo(11)
		assert.Equal(t, c6, n.BestClaim())
		n.AdjustTo(21)
		assert.Equal(t, c7, n.BestClaim())

		n.AdjustTo(11)
		assert.Equal(t, c6, n.BestClaim())
		n.AdjustTo(1)
		assert.Equal(t, c6, n.BestClaim())
		n.AdjustTo(0)
		assert.Nil(t, n.BestClaim())
	}

	// spending winning claim will make losing active claim winner
	test5 := func() {
		c1, _ = n.AddClaim(*newOutPoint(1), 2)
		c2, _ = n.AddClaim(*newOutPoint(2), 1)
		n.AdjustTo(1)
		assert.Equal(t, c1, n.BestClaim())
		n.RemoveClaim(c1.OutPoint)
		n.AdjustTo(2)
		assert.Equal(t, c2, n.BestClaim())

		n.AdjustTo(1)
		assert.Equal(t, c1, n.BestClaim())
		n.AdjustTo(0)
		assert.Nil(t, n.BestClaim())
	}

	// spending winning claim will make inactive claim winner
	test6 := func() {
		c3, _ = n.AddClaim(*newOutPoint(3), 2)
		n.AdjustTo(10)
		assert.Equal(t, c3, n.BestClaim())

		c4, _ = n.AddClaim(*newOutPoint(4), 2)
		n.AdjustTo(11)
		assert.Equal(t, c3, n.BestClaim())
		n.RemoveClaim(c3.OutPoint)
		n.AdjustTo(12)
		assert.Equal(t, c4, n.BestClaim())

		n.AdjustTo(11)
		assert.Equal(t, c3, n.BestClaim())
		n.AdjustTo(10)
		assert.Equal(t, c3, n.BestClaim())
		n.AdjustTo(0)
		assert.Nil(t, n.BestClaim())
	}

	// spending winning claim will empty out claim trie
	test7 := func() {
		c5, _ = n.AddClaim(*newOutPoint(5), 2)
		n.AdjustTo(1)
		assert.Equal(t, c5, n.BestClaim())
		n.RemoveClaim(c5.OutPoint)
		n.AdjustTo(2)
		assert.NotEqual(t, c5, n.BestClaim())

		n.AdjustTo(1)
		assert.Equal(t, c5, n.BestClaim())
		n.AdjustTo(0)
		assert.Nil(t, n.BestClaim())
	}

	// check claim with more support wins
	test8 := func() {
		c1, _ = n.AddClaim(*newOutPoint(1), 2)
		c2, _ = n.AddClaim(*newOutPoint(2), 1)
		s1, _ = n.AddSupport(*newOutPoint(11), 1, c1.ID)
		s2, _ = n.AddSupport(*newOutPoint(12), 10, c2.ID)
		n.AdjustTo(1)
		assert.Equal(t, c2, n.BestClaim())
		assert.Equal(t, claim.Amount(11), n.BestClaim().EffAmt)
		n.AdjustTo(0)
		assert.Nil(t, n.BestClaim())
	}
	// check support delay
	test9 := func() {
		c3, _ = n.AddClaim(*newOutPoint(3), 1)
		c4, _ = n.AddClaim(*newOutPoint(4), 2)
		n.AdjustTo(10)
		assert.Equal(t, c4, n.BestClaim())
		assert.Equal(t, claim.Amount(2), n.BestClaim().EffAmt)
		s4, _ = n.AddSupport(*newOutPoint(14), 10, c3.ID)
		n.AdjustTo(20)
		assert.Equal(t, c4, n.BestClaim())
		assert.Equal(t, claim.Amount(2), n.BestClaim().EffAmt)
		n.AdjustTo(21)
		assert.Equal(t, c3, n.BestClaim())
		assert.Equal(t, claim.Amount(11), n.BestClaim().EffAmt)

		n.AdjustTo(20)
		assert.Equal(t, c4, n.BestClaim())
		assert.Equal(t, claim.Amount(2), n.BestClaim().EffAmt)
		n.AdjustTo(10)
		assert.Equal(t, c4, n.BestClaim())
		assert.Equal(t, claim.Amount(2), n.BestClaim().EffAmt)
		n.AdjustTo(0)
		assert.Nil(t, n.BestClaim())
	}

	// supporting and abandoing on the same block will cause it to crash
	test10 := func() {
		c1, _ = n.AddClaim(*newOutPoint(1), 2)
		n.AdjustTo(1)
		s1, _ = n.AddSupport(*newOutPoint(11), 1, c1.ID)
		n.RemoveClaim(c1.OutPoint)
		n.AdjustTo(2)
		assert.NotEqual(t, c1, n.BestClaim())

		n.AdjustTo(1)
		assert.Equal(t, c1, n.BestClaim())
		n.AdjustTo(0)
		assert.Nil(t, n.BestClaim())
	}

	// support on abandon2
	test11 := func() {
		c1, _ = n.AddClaim(*newOutPoint(1), 2)
		s1, _ = n.AddSupport(*newOutPoint(11), 1, c1.ID)
		n.AdjustTo(1)
		assert.Equal(t, c1, n.BestClaim())

		//abandoning a support and abandoing claim on the same block will cause it to crash
		n.RemoveClaim(c1.OutPoint)
		n.RemoveSupport(s1.OutPoint)
		n.AdjustTo(2)
		assert.Nil(t, n.BestClaim())

		n.AdjustTo(1)
		assert.Equal(t, c1, n.BestClaim())
		n.AdjustTo(0)
		assert.Nil(t, n.BestClaim())
	}
	test12 := func() {
		c1, _ = n.AddClaim(*newOutPoint(1), 3)
		c2, _ = n.AddClaim(*newOutPoint(2), 2)
		n.AdjustTo(10)
		// c1 tookover since 1
		assert.Equal(t, c1, n.BestClaim())

		// C3 will takeover at 11 + 11 - 1 = 21
		c3, _ = n.AddClaim(*newOutPoint(3), 5)
		s1, _ = n.AddSupport(*newOutPoint(11), 2, c2.ID)

		n.AdjustTo(20)
		assert.Equal(t, c1, n.BestClaim())

		n.AdjustTo(21)
		assert.Equal(t, c3, n.BestClaim())

		n.RemoveClaim(c3.OutPoint)
		n.AdjustTo(22)

		// c2 (3+4) should bid over c1(5) at 21
		assert.Equal(t, c2, n.BestClaim())

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
		test12,
	}
	for _, tt := range tests {
		tt()
	}
	_ = []func(){test1, test2, test3, test4, test5, test6, test7, test8, test9, test10, test12}
}
