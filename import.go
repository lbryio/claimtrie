package claimtrie

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"path/filepath"
	"strconv"

	"github.com/lbryio/claimtrie/claim"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
	"github.com/syndtr/goleveldb/leveldb"
)

var (
	defaultHomeDir = btcutil.AppDataDir("lbrycrd.go", false)
	defaultDataDir = filepath.Join(defaultHomeDir, "data")
	dbCmdPath      = filepath.Join(defaultDataDir, "dbCmd")
)

type block struct {
	Hash chainhash.Hash
	Cmds []claim.Cmd
}

// Load ...
func Load(ct *ClaimTrie, h claim.Height, chk bool) error {
	db := DefaultRecorder()
	defer db.Close()

	for i := ct.height + 1; i <= h; i++ {
		key := strconv.Itoa(int(i))
		data, err := db.Get([]byte(key), nil)
		if err == leveldb.ErrNotFound {
			continue
		} else if err != nil {
			return err
		}
		var blk block
		if err = gob.NewDecoder(bytes.NewBuffer(data)).Decode(&blk); err != nil {
			return err
		}

		if err = ct.Commit(i - 1); err != nil {
			return err
		}
		for _, cmd := range blk.Cmds {
			if err = execute(ct, &cmd); err != nil {
				fmt.Printf("execute faile: err %s\n", err)
				return err
			}
		}
		if err = ct.Commit(i); err != nil {
			return err
		}
		if !chk {
			continue
		}
		hash, err := ct.MerkleHash()
		if err != nil {
			return err
		}
		if *hash != blk.Hash {
			return fmt.Errorf("block %d hash: got %s, want %s", i, *hash, blk.Hash)
		}
	}
	return ct.Commit(h)
}

func execute(ct *ClaimTrie, c *claim.Cmd) error {
	// Value  []byte
	fmt.Printf("%s\n", c)
	switch c.Cmd {
	case claim.CmdAddClaim:
		return ct.AddClaim(c.Name, c.OP, c.Amt)
	case claim.CmdSpendClaim:
		return ct.SpendClaim(c.Name, c.OP)
	case claim.CmdUpdateClaim:
		return ct.UpdateClaim(c.Name, c.OP, c.Amt, c.ID)
	case claim.CmdAddSupport:
		return ct.AddSupport(c.Name, c.OP, c.Amt, c.ID)
	case claim.CmdSpendSupport:
		return ct.SpendSupport(c.Name, c.OP)
	}
	return nil
}

// Recorder ..
type Recorder struct {
	db *leveldb.DB
}

// Put sets the value for the given key. It overwrites any previous value for that key.
func (r *Recorder) Put(key []byte, data interface{}) error {
	buf := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buf).Encode(data); err != nil {
		return fmt.Errorf("can't encode cmds, err: %s", err)
	}
	if err := r.db.Put(key, buf.Bytes(), nil); err != nil {
		return fmt.Errorf("can't put to db, err: %s", err)

	}
	return nil
}

// Get ...
func (r *Recorder) Get(key []byte, data interface{}) ([]byte, error) {
	return r.db.Get(key, nil)
}

// Close ...
func (r *Recorder) Close() error {
	err := r.db.Close()
	r.db = nil
	return err
}

var recorder Recorder

// DefaultRecorder ...
func DefaultRecorder() *Recorder {
	if recorder.db == nil {
		db, err := leveldb.OpenFile(dbCmdPath, nil)
		if err != nil {
			log.Fatalf("can't open :%s, err: %s\n", dbCmdPath, err)
		}
		fmt.Printf("dbCmds %s opened\n", dbCmdPath)
		recorder.db = db
	}
	return &recorder
}
