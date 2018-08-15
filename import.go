package claimtrie

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"strconv"

	"github.com/lbryio/claimtrie/change"
	"github.com/lbryio/claimtrie/claim"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

// Load ...
func Load(db *leveldb.DB, ct *ClaimTrie, ht claim.Height, verbose, chk bool) error {
	for i := ct.Height() + 1; i <= ht; i++ {
		blk, err := getBlock(db, i)
		if err != nil {
			return errors.Wrapf(err, "getBlock(db, %d)", i)
		}
		if blk == nil {
			continue
		}
		ct.Commit(i - 1)
		for _, chg := range blk.Changes {
			if err = apply(ct, chg, verbose); err != nil {
				return errors.Wrapf(err, "apply(%s)", chg)
			}
		}
		ct.Commit(i)

		if !chk {
			continue
		}
		if hash := ct.MerkleHash(); *hash != blk.Hash {
			return fmt.Errorf("blk %d hash: got %s, want %s", i, *hash, blk.Hash)
		}
	}
	ct.Commit(ht)
	return nil
}

func getBlock(db *leveldb.DB, ht claim.Height) (*change.Block, error) {
	key := strconv.Itoa(int(ht))
	data, err := db.Get([]byte(key), nil)
	if err == leveldb.ErrNotFound {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrapf(err, "db.Get(%s)", key)
	}

	var blk change.Block
	if err = gob.NewDecoder(bytes.NewBuffer(data)).Decode(&blk); err != nil {
		return nil, errors.Wrapf(err, "gob.Decode(&blk)")
	}
	return &blk, nil
}

func apply(ct *ClaimTrie, c *change.Change, verbose bool) error {
	if verbose {
		log.Printf("%s", c)
	}
	var err error
	switch c.Cmd {
	case change.AddClaim:
		err = ct.AddClaim(c.Name, c.OP, c.Amt, c.Value)
	case change.SpendClaim:
		err = ct.SpendClaim(c.Name, c.OP)
	case change.UpdateClaim:
		err = ct.UpdateClaim(c.Name, c.OP, c.Amt, c.ID, c.Value)
	case change.AddSupport:
		err = ct.AddSupport(c.Name, c.OP, c.Amt, c.ID)
	case change.SpendSupport:
		err = ct.SpendSupport(c.Name, c.OP)
	}
	return errors.Wrapf(err, "exec %s", c)
}
