package main

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"

	"github.com/lbryio/claimtrie"
	"github.com/lbryio/merkletrie"

	"github.com/urfave/cli"
)

var (
	flagAll      = cli.BoolFlag{Name: "all, a", Usage: "apply to non-value nodes"}
	flagAmount   = cli.Int64Flag{Name: "amount, a", Usage: "Amount"}
	flagHeight   = cli.Int64Flag{Name: "height, ht", Usage: "Height"}
	flagName     = cli.StringFlag{Name: "name, n", Value: "Hello", Usage: "Name"}
	flagID       = cli.StringFlag{Name: "id", Usage: "Claim ID"}
	flagOutPoint = cli.StringFlag{Name: "outpoint, op", Usage: "Outpoint. (HASH:INDEX) "}
)

var (
	errNotImplemented = fmt.Errorf("not implemented")
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
			Flags:   []cli.Flag{flagAll, flagHeight},
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

func cmdAddClaim(c *cli.Context) error {
	amount := claimtrie.Amount(c.Int64("amount"))
	if !c.IsSet("amount") {
		i, err := rand.Int(rand.Reader, big.NewInt(100))
		if err != nil {
			return err
		}
		amount = 1 + claimtrie.Amount(i.Int64())
	}

	height := claimtrie.Height(c.Int64("height"))
	if !c.IsSet("height") {
		height = ct.BestBlock()
	}

	outPoint, err := newOutPoint(c.String("outpoint"))
	if err != nil {
		return nil
	}
	return ct.AddClaim(c.String("name"), *outPoint, amount, height)
}

func cmdAddSupport(c *cli.Context) error {
	amount := claimtrie.Amount(c.Int64("amount"))
	if !c.IsSet("amount") {
		i, err := rand.Int(rand.Reader, big.NewInt(100))
		if err != nil {
			return err
		}
		amount = 1 + claimtrie.Amount(i.Int64())
	}

	height := claimtrie.Height(c.Int64("height"))
	if !c.IsSet("height") {
		height = ct.BestBlock()
	}

	outPoint, err := newOutPoint(c.String("outpoint"))
	if err != nil {
		return err
	}

	if !c.IsSet("id") {
		return fmt.Errorf("flag -id is required")
	}
	cid, err := claimtrie.NewClaimIDFromString(c.String("id"))
	if err != nil {
		return err
	}
	return ct.AddSupport(c.String("name"), *outPoint, amount, height, cid)
}

func cmdSpendClaim(c *cli.Context) error {
	outPoint, err := newOutPoint(c.String("outpoint"))
	if err != nil {
		return err
	}
	return ct.SpendClaim(c.String("name"), *outPoint)
}
func cmdSpendSupport(c *cli.Context) error {
	outPoint, err := newOutPoint(c.String("outpoint"))
	if err != nil {
		return err
	}
	return ct.SpendSupport(c.String("name"), *outPoint)
}

func cmdShow(c *cli.Context) error {
	dump := func(prefix merkletrie.Key, val merkletrie.Value) error {
		if val == nil {
			fmt.Printf("%-8s:\n", prefix)
			return nil
		}
		fmt.Printf("%-8s: %v\n", prefix, val)
		return nil
	}
	height := claimtrie.Height(c.Int64("height"))
	if !c.IsSet("height") {
		fmt.Printf("<ClaimTrie Height %d>\n", ct.BestBlock())
		return ct.Traverse(dump, false, !c.Bool("all"))
	}
	fmt.Printf("NOTE: peeking to the past is broken for now. Try RESET command instead\n")
	for commit := ct.Head(); commit != nil; commit = commit.Prev {
		meta := commit.Meta.(claimtrie.CommitMeta)
		fmt.Printf("HEAD: %d/%d\n", height, meta.Height)
		if height == meta.Height {
			return commit.MerkleTrie.Traverse(dump, false, !c.Bool("all"))
		}
	}

	return fmt.Errorf("commit not found")
}

func cmdMerkle(c *cli.Context) error {
	fmt.Printf("%s\n", (ct.MerkleHash()))
	return nil
}

func cmdCommit(c *cli.Context) error {
	height := claimtrie.Height(c.Int64("height"))
	if !c.IsSet("height") {
		height = ct.BestBlock() + 1
	}
	return ct.Commit(height)
}

func cmdReset(c *cli.Context) error {
	height := claimtrie.Height(c.Int64("height"))
	return ct.Reset(height)
}

func cmdLog(c *cli.Context) error {
	commitVisit := func(c *merkletrie.Commit) {
		meta := c.Meta.(claimtrie.CommitMeta)
		fmt.Printf("height: %d, commit %s\n", meta.Height, c.MerkleTrie.MerkleHash())
	}

	fmt.Printf("\n")
	merkletrie.Log(ct.Head(), commitVisit)
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
