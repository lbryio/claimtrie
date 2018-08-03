package main

import (
	"bufio"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/lbryio/claimtrie"
	"github.com/lbryio/claimtrie/claim"
	"github.com/lbryio/claimtrie/trie"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcutil"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/urfave/cli"
)

var (
	ct *claimtrie.ClaimTrie

	defaultHomeDir = btcutil.AppDataDir("lbrycrd.go", false)
	defaultDataDir = filepath.Join(defaultHomeDir, "data")
	dbTriePath     = filepath.Join(defaultDataDir, "dbTrie")
)

var (
	all    bool
	chk    bool
	name   string
	height claim.Height
	amt    claim.Amount
	op     claim.OutPoint
	id     claim.ID
)

var (
	flagAll      = cli.BoolFlag{Name: "all, a", Usage: "Show all nodes", Destination: &all}
	flagCheck    = cli.BoolFlag{Name: "chk, c", Usage: "Check Merkle Hash during importing", Destination: &chk}
	flagAmount   = cli.Int64Flag{Name: "amount, a", Usage: "Amount", Destination: (*int64)(&amt)}
	flagHeight   = cli.Int64Flag{Name: "height, ht", Usage: "Height"}
	flagName     = cli.StringFlag{Name: "name, n", Value: "Hello", Usage: "Name", Destination: &name}
	flagID       = cli.StringFlag{Name: "id", Usage: "Claim ID"}
	flagOutPoint = cli.StringFlag{Name: "outpoint, op", Usage: "Outpoint. (HASH:INDEX)"}
)

var (
	errNotImplemented = errors.New("not implemented")
	errHeight         = errors.New("invalid height")
)

func main() {
	app := cli.NewApp()
	app.Name = "claimtrie"
	app.Usage = "A CLI tool for LBRY ClaimTrie"
	app.Version = "0.0.1"
	app.Action = cli.ShowAppHelp
	app.Commands = []cli.Command{
		{
			Name:    "add-claim",
			Aliases: []string{"ac"},
			Usage:   "Claim a name.",
			Before:  parseArgs,
			Action:  cmdAddClaim,
			Flags:   []cli.Flag{flagName, flagOutPoint, flagAmount},
		},
		{
			Name:    "spend-claim",
			Aliases: []string{"sc"},
			Usage:   "Spend a Claim.",
			Before:  parseArgs,
			Action:  cmdSpendClaim,
			Flags:   []cli.Flag{flagName, flagOutPoint},
		},
		{
			Name:    "update-claim",
			Aliases: []string{"uc"},
			Usage:   "Update a Claim.",
			Before:  parseArgs,
			Action:  cmdUpdateClaim,
			Flags:   []cli.Flag{flagName, flagOutPoint, flagAmount, flagID},
		},
		{
			Name:    "add-support",
			Aliases: []string{"as"},
			Usage:   "Support a Claim.",
			Before:  parseArgs,
			Action:  cmdAddSupport,
			Flags:   []cli.Flag{flagName, flagOutPoint, flagAmount, flagID},
		},
		{
			Name:    "spend-support",
			Aliases: []string{"ss"},
			Usage:   "Spend a specified Support.",
			Before:  parseArgs,
			Action:  cmdSpendSupport,
			Flags:   []cli.Flag{flagName, flagOutPoint},
		},
		{
			Name:    "show",
			Aliases: []string{"s"},
			Usage:   "Show the status of Stage)",
			Before:  parseArgs,
			Action:  cmdShow,
			Flags:   []cli.Flag{flagAll, flagName, flagHeight},
		},
		{
			Name:    "merkle",
			Aliases: []string{"m"},
			Usage:   "Show the Merkle Hash of the Stage.",
			Before:  parseArgs,
			Action:  cmdMerkle,
		},
		{
			Name:    "commit",
			Aliases: []string{"c"},
			Usage:   "Commit the current Stage to database.",
			Before:  parseArgs,
			Action:  cmdCommit,
			Flags:   []cli.Flag{flagHeight},
		},
		{
			Name:    "reset",
			Aliases: []string{"r"},
			Usage:   "Reset the Head commit and Stage to a specified commit.",
			Before:  parseArgs,
			Action:  cmdReset,
			Flags:   []cli.Flag{flagHeight},
		},
		{
			Name:    "log",
			Aliases: []string{"l"},
			Usage:   "List the commits in the coommit database.",
			Before:  parseArgs,
			Action:  cmdLog,
		},
		{
			Name:    "load",
			Aliases: []string{"ld"},
			Usage:   "Load prerecorded command from datbase.",
			Before:  parseArgs,
			Action:  cmdLoad,
			Flags:   []cli.Flag{flagHeight, flagCheck},
		},
		{
			Name:    "shell",
			Aliases: []string{"sh"},
			Usage:   "Enter interactive mode",
			Before:  parseArgs,
			Action:  func(c *cli.Context) { cmdShell(app) },
		},
	}

	dbTrie, err := leveldb.OpenFile(dbTriePath, nil)
	if err != nil {
		log.Fatalf("can't open dbTrie at %s, err: %s\n", dbTriePath, err)
	}
	fmt.Printf("dbTriePath: %q\n", dbTriePath)
	ct = claimtrie.New(dbTrie, nil)
	if err := app.Run(os.Args); err != nil {
		fmt.Printf("error: %s\n", err)
	}
}

func cmdAddClaim(c *cli.Context) error {
	return ct.AddClaim(name, op, amt)
}

func cmdSpendClaim(c *cli.Context) error {
	return ct.SpendClaim(name, op)
}

func cmdUpdateClaim(c *cli.Context) error {
	if !c.IsSet("id") {
		return fmt.Errorf("flag id is required")
	}
	return ct.UpdateClaim(name, op, amt, id)
}

func cmdAddSupport(c *cli.Context) error {
	if !c.IsSet("id") {
		return fmt.Errorf("flag id is required")
	}
	return ct.AddSupport(name, op, amt, id)
}

func cmdSpendSupport(c *cli.Context) error {
	return ct.SpendSupport(name, op)
}

func cmdShow(c *cli.Context) error {
	fmt.Printf("\n<ClaimTrie Height %d (Nodes) >\n\n", ct.Height())
	if all {
		name = ""
	}
	return ct.NodeMgr().Show(name)

	// fmt.Printf("\n<ClaimTrie Height %d (Stage) >\n\n", ct.Height())
	// return ct.Traverse(showNode())
}

func cmdMerkle(c *cli.Context) error {
	h, err := ct.MerkleHash()
	if err != nil {
		return err
	}
	fmt.Printf("%s at %d\n", h, ct.Height())
	return nil
}

func cmdCommit(c *cli.Context) error {
	if !c.IsSet("height") {
		height = ct.Height() + 1
	}
	return ct.Commit(height)
}

func cmdReset(c *cli.Context) error {
	return ct.Reset(height)
}

func cmdLog(c *cli.Context) error {
	visit := func(c *claimtrie.Commit) {
		meta := c.Meta
		fmt.Printf("%s at %d\n", c.MerkleRoot, meta.Height)
	}
	return claimtrie.Log(ct.Head(), visit)
}

func cmdLoad(c *cli.Context) error {
	return claimtrie.Load(ct, height, chk)
}

func cmdShell(app *cli.App) {
	cli.OsExiter = func(c int) {}
	reader := bufio.NewReader(os.Stdin)
	sigs := make(chan os.Signal, 1)
	go func() {
		for range sigs {
			fmt.Printf("\n(type quit or q to exit)\n\n")
			fmt.Printf("%s > ", app.Name)
		}
	}()
	defer close(sigs)
	// signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		fmt.Printf("%s > ", app.Name)
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("error: %s\n", err)
		}
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		if text == "quit" || text == "q" {
			break
		}
		err = app.Run(append(os.Args[1:], strings.Split(text, " ")...))
		if err != nil {
			fmt.Printf("error: %s\n", err)
		}
	}
	signal.Stop(sigs)
}

func parseArgs(c *cli.Context) error {
	parsers := []func(*cli.Context) error{
		parseOP,
		parseOP,
		parseAmt,
		parseHeight,
		parseID,
	}
	for _, p := range parsers {
		if err := p(c); err != nil {
			return err
		}
	}
	return nil
}

func randInt(min, max int64) int64 {
	i, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		panic(err)
	}
	return min + i.Int64()
}

func parseHeight(c *cli.Context) error {
	height = claim.Height(c.Int("height"))
	return nil
}

// parseOP generates random OutPoint for the ease of testing.
func parseOP(c *cli.Context) error {
	var err error
	h := &chainhash.Hash{}
	idx := randInt(0, 256)
	if _, err = rand.Read(h[:]); err != nil {
		return err
	}
	var sh string
	if c.IsSet("outpoint") {
		if _, err = fmt.Sscanf(c.String("outpoint"), "%64s:%d", &sh, &idx); err != nil {
			return err
		}
		if h, err = chainhash.NewHashFromStr(sh); err != nil {
			return err
		}
	}
	op = *claim.NewOutPoint(h, uint32(idx))
	return nil
}

func parseAmt(c *cli.Context) error {
	if !c.IsSet("amount") {
		amt = claim.Amount(randInt(1, 500))
	}
	return nil
}

func parseID(c *cli.Context) error {
	if !c.IsSet("id") {
		return nil
	}
	var err error
	id, err = claim.NewIDFromString(c.String("id"))
	return err
}

var showNode = func() trie.Visit {
	return func(prefix trie.Key, val trie.Value) error {
		if val == nil || val.Hash() == nil {
			fmt.Printf("%-8s:\n", prefix)
			return nil
		}
		fmt.Printf("%-8s: %v\n", prefix, val)
		return nil
	}
}

var recall = func(h claim.Height, visit trie.Visit) trie.Visit {
	return func(prefix trie.Key, val trie.Value) error {
		n := val.(*claim.Node)
		old := n.Height()
		err := n.Recall(h)
		if err == nil {
			err = visit(prefix, val)
		}
		if err == nil {
			err = n.AdjustTo(old)
		}
		return err
	}
}
