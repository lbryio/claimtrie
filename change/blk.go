package change

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// Block ...
type Block struct {
	Hash    chainhash.Hash
	Changes []*Change
}
