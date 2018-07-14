package claim

import (
	"testing"

	"github.com/lbryio/claimtrie/memento"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/assert"
)

var (
	opA = wire.OutPoint{Hash: newHash("0000000000000000000000000000000011111111111111111111111111111111"), Index: 1}
	opB = wire.OutPoint{Hash: newHash("0000000000000000000000000000000022222222222222222222222222222222"), Index: 2}
	opC = wire.OutPoint{Hash: newHash("0000000000000000000000000000000033333333333333333333333333333333"), Index: 3}
	opD = wire.OutPoint{Hash: newHash("0000000000000000000000000000000044444444444444444444444444444444"), Index: 4}
	opE = wire.OutPoint{Hash: newHash("0000000000000000000000000000000555555555555555555555555555555555"), Index: 5}
	opF = wire.OutPoint{Hash: newHash("0000000000000000000000000000000666666666666666666666666666666666"), Index: 6}
	opX = wire.OutPoint{Hash: newHash("0000000000000000000000000000000777777777777777777777777777777777"), Index: 7}
	opY = wire.OutPoint{Hash: newHash("0000000000000000000000000000000888888888888888888888888888888888"), Index: 8}
	opZ = wire.OutPoint{Hash: newHash("0000000000000000000000000000000999999999999999999999999999999999"), Index: 9}

	cA = New(opA, 0)
	cB = New(opB, 0)
	cC = New(opC, 0)
	cD = New(opD, 0)
	cE = New(opE, 0)
	sX = NewSupport(opX, 0, ID{})
	sY = NewSupport(opY, 0, ID{})
	sZ = NewSupport(opZ, 0, ID{})
)

func newHash(s string) chainhash.Hash {
	h, _ := chainhash.NewHashFromStr(s)
	return *h
}

// The example on the [whitepaper](https://beta.lbry.tech/whitepaper.html)
func Test_BestClaimExample(t *testing.T) {

	SetParams(ActiveDelayFactor(32))
	defer SetParams(ResetParams())

	n := NewNode()
	at := func(h Height) *Node {
		if err := n.AdjustTo(h - 1); err != nil {
			panic(err)
		}
		return n
	}
	bestAt := func(at Height) *Claim {
		if len(n.mem.Executed()) != 0 {
			n.Increment()
			n.Decrement()
		}
		for n.height < at {
			if err := n.Redo(); err == memento.ErrCommandStackEmpty {
				n.Increment()
			}
		}
		for n.height > at {
			n.Decrement()
		}
		return n.BestClaim()
	}

	sX.SetAmt(14).SetClaimID(cA.ID)

	at(13).AddClaim(cA.SetAmt(10))    // Claim A for 10LBC is accepted. It is the first claim, so it immediately becomes active and controlling.
	at(1001).AddClaim(cB.SetAmt(20))  // Claim B for 20LBC is accepted. It’s activation height is 1001+min(4032,floor(1001−1332))=1001+30=1031
	at(1010).AddSupport(sX)           // Support X for 14LBC for claim A is accepted. Since it is a support for the controlling claim, it activates immediately.
	at(1020).AddClaim(cC.SetAmt(50))  // Claim C for 50LBC is accepted. The activation height is 1020+min(4032,floor(1020−1332))=1020+31=1051
	at(1040).AddClaim(cD.SetAmt(300)) // Claim D for 300LBC is accepted. The activation height is 1040+min(4032,floor(1040−1332))=1040+32=1072

	assert.Equal(t, cA, bestAt(13))   // A(10) is controlling.
	assert.Equal(t, cA, bestAt(1001)) // A(10) is controlling, B(20) is accepted.
	assert.Equal(t, cA, bestAt(1010)) // A(10+14) is controlling, B(20) is accepted.
	assert.Equal(t, cA, bestAt(1020)) // A(10+14) is controlling, B(20) is accepted, C(50) is accepted.

	// Claim B activates. It has 20LBC, while claim A has 24LBC (10 original + 14 from support X). There is no takeover, and claim A remains controlling.
	assert.Equal(t, cA, bestAt(1031)) // A(10+14) is controlling, B(20) is active, C(50) is accepted.
	assert.Equal(t, cA, bestAt(1040)) //A(10+14) is controlling, B(20) is active, C(50) is accepted, D(300) is accepted.

	// Claim C activates. It has 50LBC, while claim A has 24LBC, so a takeover is initiated.
	// The takeover height for this name is set to 1051, and therefore the activation delay for all the claims becomes min(4032, floor(1051−1051/32)) = 0.
	// All the claims become active.
	// The totals for each claim are recalculated, and claim D becomes controlling because it has the highest total.
	assert.Equal(t, cD, bestAt(1051)) // A(10+14) is active, B(20) is active, C(50) is active, D(300) is controlling.
}

func Test_BestClaim(t *testing.T) {

	SetParams(ActiveDelayFactor(1))
	defer SetParams(ResetParams())

	n := NewNode()
	at := func(h Height) *Node {
		if err := n.AdjustTo(h - 1); err != nil {
			panic(err)
		}
		return n
	}

	bestAt := func(at Height) *Claim {
		if len(n.mem.Executed()) != 0 {
			n.Increment()
			n.Decrement()
		}
		for n.height < at {
			if err := n.Redo(); err == memento.ErrCommandStackEmpty {
				n.Increment()
			}
		}
		for n.height > at {
			n.Decrement()
		}
		return n.BestClaim()
	}

	tests := []func(t *testing.T){
		// No competing bids.
		func(t *testing.T) {
			at(1).AddClaim(cA.SetAmt(1))

			assert.Equal(t, cA, bestAt(1))
			assert.Nil(t, bestAt(0))
		},
		// Competing bids inserted at the same height.
		func(t *testing.T) {
			at(1).AddClaim(cA.SetAmt(1))
			at(1).AddClaim(cB.SetAmt(2))

			assert.Equal(t, cB, bestAt(1))
			assert.Nil(t, bestAt(0))

		},
		// Two claims with the same amount. The older one wins.
		func(t *testing.T) {
			at(1).AddClaim(cA.SetAmt(1))
			at(2).AddClaim(cB.SetAmt(1))

			assert.Equal(t, cA, bestAt(1))
			assert.Equal(t, cA, bestAt(2))
			assert.Equal(t, cA, bestAt(3))
			assert.Equal(t, cA, bestAt(2))
			assert.Equal(t, cA, bestAt(1))
			assert.Nil(t, bestAt(0))
		},
		// Check claim takeover.
		func(t *testing.T) {
			at(1).AddClaim(cA.SetAmt(1))
			at(10).AddClaim(cB.SetAmt(2))

			assert.Equal(t, cA, bestAt(10))
			assert.Equal(t, cA, bestAt(11))
			assert.Equal(t, cB, bestAt(21))
			assert.Equal(t, cA, bestAt(11))
			assert.Equal(t, cA, bestAt(1))
			assert.Nil(t, bestAt(0))
		},
		// Spending winning claim will make losing active claim winner.
		func(t *testing.T) {
			at(1).AddClaim(cA.SetAmt(2))
			at(1).AddClaim(cB.SetAmt(1))
			at(2).RemoveClaim(cA.OutPoint)

			assert.Equal(t, cA, bestAt(1))
			assert.Equal(t, cB, bestAt(2))
			assert.Equal(t, cA, bestAt(1))
			assert.Nil(t, bestAt(0))
		},
		// spending winning claim will make inactive claim winner
		func(t *testing.T) {
			at(1).AddClaim(cA.SetAmt(2))
			at(11).AddClaim(cB.SetAmt(1))
			at(12).RemoveClaim(cA.OutPoint)

			assert.Equal(t, cA, bestAt(10))
			assert.Equal(t, cA, bestAt(11))
			assert.Equal(t, cB, bestAt(12))
			assert.Equal(t, cA, bestAt(11))
			assert.Equal(t, cA, bestAt(10))
			assert.Nil(t, bestAt(0))
		},
		// spending winning claim will empty out claim trie
		func(t *testing.T) {
			at(1).AddClaim(cA.SetAmt(2))
			at(2).RemoveClaim(cA.OutPoint)

			assert.Equal(t, cA, bestAt(1))
			assert.NotEqual(t, cA, bestAt(2))
			assert.Equal(t, cA, bestAt(1))
			assert.Nil(t, bestAt(0))
		},
		// check claim with more support wins
		func(t *testing.T) {
			at(1).AddClaim(cA.SetAmt(2))
			at(1).AddClaim(cB.SetAmt(1))
			at(1).AddSupport(sX.SetAmt(1).SetClaimID(cA.ID))
			at(1).AddSupport(sY.SetAmt(10).SetClaimID(cB.ID))

			assert.Equal(t, cB, bestAt(1))
			assert.Equal(t, Amount(11), bestAt(1).EffAmt)
			assert.Nil(t, bestAt(0))
		},
		// check support delay
		func(t *testing.T) {
			at(1).AddClaim(cA.SetAmt(1))
			at(1).AddClaim(cB.SetAmt(2))
			at(11).AddSupport(sX.SetAmt(10).SetClaimID(cA.ID))

			assert.Equal(t, cB, bestAt(10))
			assert.Equal(t, Amount(2), bestAt(10).EffAmt)
			assert.Equal(t, cB, bestAt(20))
			assert.Equal(t, Amount(2), bestAt(20).EffAmt)
			assert.Equal(t, cA, bestAt(21))
			assert.Equal(t, Amount(11), bestAt(21).EffAmt)
			assert.Equal(t, cB, bestAt(20))
			assert.Equal(t, Amount(2), bestAt(20).EffAmt)
			assert.Equal(t, cB, bestAt(10))
			assert.Equal(t, Amount(2), bestAt(10).EffAmt)
			assert.Nil(t, bestAt(0))
		},
		// supporting and abandoing on the same block will cause it to crash
		func(t *testing.T) {
			at(1).AddClaim(cA.SetAmt(2))
			at(2).AddSupport(sX.SetAmt(1).SetClaimID(cA.ID))
			at(2).RemoveClaim(cA.OutPoint)

			assert.NotEqual(t, cA, bestAt(2))
			assert.Equal(t, cA, bestAt(1))
			assert.Nil(t, bestAt(0))
		},
		// support on abandon2
		func(t *testing.T) {
			at(1).AddClaim(cA.SetAmt(2))
			at(1).AddSupport(sX.SetAmt(1).SetClaimID(cA.ID))

			// abandoning a support and abandoing claim on the same block will cause it to crash
			at(2).RemoveClaim(cA.OutPoint)
			at(2).RemoveSupport(sX.OutPoint)

			assert.Equal(t, cA, bestAt(1))
			assert.Nil(t, bestAt(2))
			assert.Equal(t, cA, bestAt(1))
			assert.Nil(t, bestAt(0))
		},
		// expiration
		func(t *testing.T) {
			SetParams(OriginalClaimExpirationTime(5))
			defer SetParams(OriginalClaimExpirationTime(DefaultOriginalClaimExpirationTime))

			at(1).AddClaim(cA.SetAmt(2))
			at(5).AddClaim(cB.SetAmt(1))

			assert.Equal(t, cA, bestAt(1))
			assert.Equal(t, cA, bestAt(5))
			assert.Equal(t, cB, bestAt(6))
			assert.Equal(t, cB, bestAt(7))
			assert.Equal(t, cB, bestAt(6))
			assert.Equal(t, cA, bestAt(5))
			assert.Equal(t, cA, bestAt(1))
			assert.Nil(t, bestAt(0))
		},

		// check claims expire and is not updateable (may be changed in future soft fork)
		//     CMutableTransaction tx3 = fixture.MakeClaim(fixture.GetCoinbase(),"test","one",2);
		//     fixture.IncrementBlocks(1);
		//     BOOST_CHECK(is_best_claim("test",tx3));
		//     fixture.IncrementBlocks(pclaimTrie->nExpirationTime);
		//     CMutableTransaction u1 = fixture.MakeUpdate(tx3,"test","two",ClaimIdHash(tx3.GetHash(),0) ,2);
		//     BOOST_CHECK(!is_best_claim("test",u1));

		//     fixture.DecrementBlocks(pclaimTrie->nExpirationTime);
		//     BOOST_CHECK(is_best_claim("test",tx3));
		//     fixture.DecrementBlocks(1);

		// check supports expire and can cause supported bid to lose claim
		//     CMutableTransaction tx4 = fixture.MakeClaim(fixture.GetCoinbase(),"test","one",1);
		//     CMutableTransaction tx5 = fixture.MakeClaim(fixture.GetCoinbase(),"test","one",2);
		//     CMutableTransaction s1 = fixture.MakeSupport(fixture.GetCoinbase(),tx4,"test",2);
		//     fixture.IncrementBlocks(1);
		//     BOOST_CHECK(is_best_claim("test",tx4));
		//     CMutableTransaction u2 = fixture.MakeUpdate(tx4,"test","two",ClaimIdHash(tx4.GetHash(),0) ,1);
		//     CMutableTransaction u3 = fixture.MakeUpdate(tx5,"test","two",ClaimIdHash(tx5.GetHash(),0) ,2);
		//     fixture.IncrementBlocks(pclaimTrie->nExpirationTime);
		//     BOOST_CHECK(is_best_claim("test",u3));
		//     fixture.DecrementBlocks(pclaimTrie->nExpirationTime);
		//     BOOST_CHECK(is_best_claim("test",tx4));
		//     fixture.DecrementBlocks(1);

		// check updated claims will extend expiration
		//     CMutableTransaction tx6 = fixture.MakeClaim(fixture.GetCoinbase(),"test","one",2);
		//     fixture.IncrementBlocks(1);
		//     BOOST_CHECK(is_best_claim("test",tx6));
		//     CMutableTransaction u4 = fixture.MakeUpdate(tx6,"test","two",ClaimIdHash(tx6.GetHash(),0) ,2);
		//     fixture.IncrementBlocks(1);
		//     BOOST_CHECK(is_best_claim("test",u4));
		//     fixture.IncrementBlocks(pclaimTrie->nExpirationTime-1);
		//     BOOST_CHECK(is_best_claim("test",u4));
		//     fixture.IncrementBlocks(1);
		//     BOOST_CHECK(!is_best_claim("test",u4));
		//     fixture.DecrementBlocks(1);
		//     BOOST_CHECK(is_best_claim("test",u4));
		//     fixture.DecrementBlocks(pclaimTrie->nExpirationTime);
		//     BOOST_CHECK(is_best_claim("test",tx6));
	}
	for _, tt := range tests {
		t.Run("BestClaim", tt)
	}
}
