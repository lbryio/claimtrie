package main

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"

	"github.com/urfave/cli"

	"github.com/lbryio/claimtrie"
	"github.com/lbryio/claimtrie/claim"
	"github.com/lbryio/claimtrie/trie"
)

var (
	flagAll      = cli.BoolFlag{Name: "all, a", Usage: "apply to non-value nodes"}
	flagAmount   = cli.Int64Flag{Name: "amount, a", Usage: "Amount"}
	flagHeight   = cli.Int64Flag{Name: "height, ht", Usage: "Height"}
	flagName     = cli.StringFlag{Name: "name, n", Value: "Hello", Usage: "Name"}
	flagID       = cli.StringFlag{Name: "id", Usage: "Claim ID"}
	flagOutPoint = cli.StringFlag{Name: "outpoint, op", Usage: "Outpoint. (HASH:INDEX)"}
	flagJSON     = cli.BoolFlag{Name: "json, j", Usage: "Show Claim / Support in JSON format."}
)

var (
	errNotImplemented = errors.New("not implemented")
	errInvalidHeight  = errors.New("invalid height")
	errCommitNotFound = errors.New("commit not found")
)

func main() {
	app := cli.NewApp()

	app.Name = "claimtrie"
	app.Usage = "A CLI tool for ClaimTrie"
	app.Version = "0.0.1"
	app.Action = cli.ShowAppHelp
	app.Commands = []cli.Command{
		{
			Name:    "add-claim",
			Aliases: []string{"ac"},
			Usage:   "Claim a name with specified amount. (outPoint is generated randomly, if unspecified)",
			Action:  cmdAddClaim,
			Flags:   []cli.Flag{flagName, flagOutPoint, flagAmount, flagHeight},
		},
		{
			Name:    "add-support",
			Aliases: []string{"as"},
			Usage:   "Add support to a specified Claim. (outPoint is generated randomly, if unspecified)",
			Action:  cmdAddSupport,
			Flags:   []cli.Flag{flagName, flagOutPoint, flagAmount, flagHeight, flagID},
		},
		{
			Name:    "spend-claim",
			Aliases: []string{"sc"},
			Usage:   "Spend a specified Claim.",
			Action:  cmdSpendClaim,
			Flags:   []cli.Flag{flagName, flagOutPoint},
		},
		{
			Name:    "spend-support",
			Aliases: []string{"ss"},
			Usage:   "Spend a specified Support.",
			Action:  cmdSpendSupport,
			Flags:   []cli.Flag{flagName, flagOutPoint},
		},
		{
			Name:    "show",
			Aliases: []string{"s"},
			Usage:   "Show the Key-Value pairs of the Stage or specified commit. (links nodes are showed if -a is also specified)",
			Action:  cmdShow,
			Flags:   []cli.Flag{flagAll, flagJSON, flagHeight},
		},
		{
			Name:    "merkle",
			Aliases: []string{"m"},
			Usage:   "Show the Merkle Hash of the Stage.",
			Action:  cmdMerkle,
		},
		{
			Name:    "commit",
			Aliases: []string{"c"},
			Usage:   "Commit the current Stage to commit database.",
			Action:  cmdCommit,
			Flags:   []cli.Flag{flagHeight},
		},
		{
			Name:    "reset",
			Aliases: []string{"r"},
			Usage:   "Reset the Stage to a specified commit.",
			Action:  cmdReset,
			Flags:   []cli.Flag{flagHeight},
		},
		{
			Name:    "log",
			Aliases: []string{"l"},
			Usage:   "List the commits in the coommit database.",
			Action:  cmdLog,
		},
		{
			Name:    "shell",
			Aliases: []string{"sh"},
			Usage:   "Enter interactive mode",
			Action:  func(c *cli.Context) { cmdShell(app) },
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("error: %s\n", err)
	}
}

func randInt(min, max int64) int64 {
	i, err := rand.Int(rand.Reader, big.NewInt(100))
	if err != nil {
		panic(err)
	}
	return min + i.Int64()
}

var ct = claimtrie.New()

// newOutPoint generates random OutPoint for the ease of testing.
func newOutPoint(s string) (*wire.OutPoint, error) {
	if len(s) == 0 {
		var h chainhash.Hash
		if _, err := rand.Read(h[:]); err != nil {
			return nil, err
		}
		return wire.NewOutPoint(&h, uint32(h[0])), nil
	}
	fields := strings.Split(s, ":")
	if len(fields) != 2 {
		return nil, fmt.Errorf("invalid outpoint format (HASH:INDEX)")
	}
	h, err := chainhash.NewHashFromStr(fields[0])
	if err != nil {
		return nil, err
	}
	idx, err := strconv.Atoi(fields[1])
	if err != nil {
		return nil, err
	}
	return wire.NewOutPoint(h, uint32(idx)), nil
}

type args struct {
	*cli.Context
	err error
}

func (a *args) amount() claim.Amount {
	if a.err != nil {
		return 0
	}
	amt := a.Int64("amount")
	if !a.IsSet("amount") {
		amt = randInt(1, 500)
	}
	return claim.Amount(amt)
}

func (a *args) outPoint() wire.OutPoint {
	if a.err != nil {
		return wire.OutPoint{}
	}
	op, err := newOutPoint(a.String("outpoint"))
	a.err = err

	return *op
}

func (a *args) name() (name string) {
	if a.err != nil {
		return
	}
	return a.String("name")
}

func (a *args) id() (id claim.ID) {
	if a.err != nil {
		return
	}
	if !a.IsSet("id") {
		a.err = fmt.Errorf("flag -id is required")
		return
	}
	id, a.err = claim.NewIDFromString(a.String("id"))
	return
}

func (a *args) height() (h claim.Height, ok bool) {
	if a.err != nil {
		return 0, false
	}
	return claim.Height(a.Int64("height")), a.IsSet("height")
}

func (a *args) json() bool {
	if a.err != nil {
		return false
	}
	return a.IsSet("json")
}

func (a *args) all() bool {
	if a.err != nil {
		return false
	}
	return a.Bool("all")
}

var showNode = func(showJSON bool) trie.Visit {
	return func(prefix trie.Key, val trie.Value) error {
		if val == nil || val.Hash() == nil {
			fmt.Printf("%-8s:\n", prefix)
			return nil
		}
		if !showJSON {
			fmt.Printf("%-8s: %v\n", prefix, val)
			return nil
		}
		b, err := json.MarshalIndent(val, "", "  ")
		if err != nil {
			return err
		}
		fmt.Printf("%-8s: %s\n", prefix, b)
		return nil
	}
}

var recall = func(h claim.Height, visit trie.Visit) trie.Visit {
	return func(prefix trie.Key, val trie.Value) (err error) {
		n := val.(*claim.Node)
		for err == nil && n.Height() > h {
			err = n.Decrement()
		}
		if err == nil {
			err = visit(prefix, val)
		}
		for err == nil && n.Height() < ct.Height() {
			err = n.Redo()
		}
		return err
	}
}

func cmdAddClaim(c *cli.Context) error {
	a := args{Context: c}
	amt := a.amount()
	op := a.outPoint()
	name := a.name()
	if a.err != nil {
		return a.err
	}
	return ct.AddClaim(name, op, amt)
}

func cmdAddSupport(c *cli.Context) error {
	a := args{Context: c}
	name := a.name()
	amt := a.amount()
	op := a.outPoint()
	id := a.id()
	if a.err != nil {
		return a.err
	}
	return ct.AddSupport(name, op, amt, id)
}

func cmdSpendClaim(c *cli.Context) error {
	a := args{Context: c}
	name := a.name()
	op := a.outPoint()
	if a.err != nil {
		return a.err
	}
	return ct.SpendClaim(name, op)
}

func cmdSpendSupport(c *cli.Context) error {
	a := args{Context: c}
	name := a.name()
	op := a.outPoint()
	if a.err != nil {
		return a.err
	}
	return ct.SpendSupport(name, op)
}

func cmdShow(c *cli.Context) error {
	a := args{Context: c}
	h, setHeight := a.height()
	setJSON := a.json()
	setAll := a.all()
	if a.err != nil {
		return a.err
	}
	if h > ct.Height() {
		return errInvalidHeight
	}
	visit := showNode(setJSON)
	if !setHeight {
		fmt.Printf("\n<ClaimTrie Height %d (Stage) >\n\n", ct.Height())
		return ct.Traverse(visit, false, !setAll)
	}

	visit = recall(h, visit)
	for commit := ct.Head(); commit != nil; commit = commit.Prev {
		meta := commit.Meta.(claimtrie.CommitMeta)
		if h == meta.Height {
			fmt.Printf("\n<ClaimTrie Height %d>\n\n", h)
			return commit.MerkleTrie.Traverse(visit, false, !setAll)
		}
	}

	return errCommitNotFound
}

func cmdMerkle(c *cli.Context) error {
	fmt.Printf("%s\n", (ct.MerkleHash()))
	return nil
}

func cmdCommit(c *cli.Context) error {
	h := claim.Height(c.Int64("height"))
	if !c.IsSet("height") {
		h = ct.Height() + 1
	}
	return ct.Commit(h)
}

func cmdReset(c *cli.Context) error {
	h := claim.Height(c.Int64("height"))
	return ct.Reset(h)
}

func cmdLog(c *cli.Context) error {
	commitVisit := func(c *trie.Commit) {
		meta := c.Meta.(claimtrie.CommitMeta)
		fmt.Printf("height: %d, commit %s\n", meta.Height, c.MerkleTrie.MerkleHash())
	}

	fmt.Printf("\n")
	trie.Log(ct.Head(), commitVisit)
	return nil
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
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
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
