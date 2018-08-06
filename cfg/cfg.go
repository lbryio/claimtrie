package cfg

import (
	"path/filepath"

	"github.com/btcsuite/btcutil"
)

// Index ...
type Index int

// ...
const (
	TrieDB Index = 1 << iota
	CommitDB
	NodeDB
	ClaimScriptDB
)

var (
	defaultHomeDir = btcutil.AppDataDir("lbrycrd.go", false)
	defaultDataDir = filepath.Join(defaultHomeDir, "data")
)

// Config ...
type Config struct {
	path string
}

var datastores = map[Index]string{
	ClaimScriptDB: "cs.db", // Exported from BTCD

	CommitDB: "commit.db",
	TrieDB:   "trie.db",
	NodeDB:   "nm.db",
}

// DefaultConfig ...
func DefaultConfig(idx Index) string {
	return filepath.Join(defaultDataDir, datastores[idx])
}
