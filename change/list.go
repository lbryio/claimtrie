package change

import (
	"bytes"
	"encoding/gob"
	"fmt"

	"github.com/lbryio/claimtrie/claim"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

// List ...
type List struct {
	db   *leveldb.DB
	name string
	chgs []*Change
	err  error
}

// NewChangeList ...
func NewChangeList(db *leveldb.DB, name string) *List {
	return &List{db: db, name: name}
}

// Changes returns the Changes in the list.
func (cl *List) Changes() []*Change {
	return cl.chgs
}

// Load loads Changes from database.
func (cl *List) Load() *List {
	if cl.err == nil {
		cl.chgs, cl.err = loadChanges(cl.db, cl.name)
	}
	return cl
}

// Save saves Changes to database.
func (cl *List) Save() *List {
	if cl.err == nil {
		cl.err = saveChanges(cl.db, cl.name, cl.chgs)
	}
	return cl
}

// Append appenda a Change to the Changes in the list.
func (cl *List) Append(chg *Change) *List {
	cl.chgs = append(cl.chgs, chg)
	return cl
}

// Truncate truncates Changes that has Heiht larger than ht.
func (cl *List) Truncate(ht claim.Height) *List {
	for i, chg := range cl.chgs {
		if chg.Height > ht {
			cl.chgs = cl.chgs[:i]
			break
		}
	}
	return cl
}

// Dump prints out the Changes in the list. (Debugging only.)
func (cl *List) Dump() *List {
	for i, chg := range cl.chgs {
		fmt.Printf("chgs[%d] %s\n", i, chg)
	}
	return cl
}

func loadChanges(db *leveldb.DB, name string) ([]*Change, error) {
	data, err := db.Get([]byte(name), nil)
	if err == leveldb.ErrNotFound {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrapf(err, "db.Get(%s)", name)
	}
	var chgs []*Change
	if err = gob.NewDecoder(bytes.NewBuffer(data)).Decode(&chgs); err != nil {
		return nil, errors.Wrapf(err, "gob.Decode(&blk)")
	}
	return chgs, nil
}

func saveChanges(db *leveldb.DB, name string, chgs []*Change) error {
	buf := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buf).Encode(&chgs); err != nil {
		return errors.Wrapf(err, "gob.Decode(&blk)")
	}
	return errors.Wrapf(db.Put([]byte(name), buf.Bytes(), nil), "db.put(%s, buf)", name)
}
