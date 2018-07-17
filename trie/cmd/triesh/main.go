package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/lbryio/claimtrie/trie"
	"github.com/urfave/cli"
)

var (
	flgKey     = cli.StringFlag{Name: "key, k", Usage: "Key"}
	flgValue   = cli.StringFlag{Name: "value, v", Usage: "Value"}
	flgAll     = cli.BoolFlag{Name: "all, a", Usage: "Apply to non-value nodes"}
	flgMessage = cli.StringFlag{Name: "message, m", Usage: "Commit Message"}
	flgID      = cli.StringFlag{Name: "id", Usage: "Commit ID"}
)

var (
	// ErrNotImplemented is returned when a function is not implemented yet.
	ErrNotImplemented = fmt.Errorf("not implemented")
)

func main() {
	app := cli.NewApp()

	app.Name = "triesh"
	app.Usage = "A CLI tool for Merkle MerkleTrie"
	app.Version = "0.0.1"
	app.Action = cli.ShowAppHelp

	app.Commands = []cli.Command{
		{
			Name:    "update",
			Aliases: []string{"u"},
			Usage:   "Update Value for Key",
			Action:  cmdUpdate,
			Flags:   []cli.Flag{flgKey, flgValue},
		},
		{
			Name:    "get",
			Aliases: []string{"g"},
			Usage:   "Get Value for specified Key",
			Action:  cmdGet,
			Flags:   []cli.Flag{flgKey, flgID},
		},
		{
			Name:    "show",
			Aliases: []string{"s"},
			Usage:   "Show Key-Value pairs of specified commit",
			Action:  cmdShow,
			Flags:   []cli.Flag{flgAll, flgID},
		},
		{
			Name:    "merkle",
			Aliases: []string{"m"},
			Usage:   "Show Merkle Hash of stage",
			Action:  cmdMerkle,
		},
		{
			Name:    "prune",
			Aliases: []string{"p"},
			Usage:   "Prune link nodes that doesn't reach to any value",
			Action:  cmdPrune,
		},
		{
			Name:    "commit",
			Aliases: []string{"c"},
			Usage:   "Commit current stage to database",
			Action:  cmdCommit,
			Flags:   []cli.Flag{flgMessage},
		},
		{
			Name:    "reset",
			Aliases: []string{"r"},
			Usage:   "Reset HEAD & Stage to specified commit",
			Action:  cmdReset,
			Flags:   []cli.Flag{flgAll, flgID},
		},
		{
			Name:    "log",
			Aliases: []string{"l"},
			Usage:   "Show commit logs",
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

type strValue string

func (s strValue) Hash() *chainhash.Hash {
	h := chainhash.DoubleHashH([]byte(s))
	return &h
}

var (
	mt   = trie.New()
	head = trie.NewCommit(nil, "initial", mt)
	stg  = trie.NewStage(mt)
)

func commitVisit(c *trie.Commit) {
	fmt.Printf("commit %s\n\n", c.MerkleTrie.MerkleHash())
	fmt.Printf("\t%s\n\n", c.Meta.(string))
}

func cmdUpdate(c *cli.Context) error {
	key, value := c.String("key"), c.String("value")
	fmt.Printf("%s=%s\n", key, value)
	if len(value) == 0 {
		return stg.Update(trie.Key(key), nil)
	}
	return stg.Update(trie.Key(key), strValue(value))
}

func cmdGet(c *cli.Context) error {
	key := c.String("key")
	value, err := stg.Get(trie.Key(key))
	if err != nil {
		return err
	}
	if str, ok := value.(strValue); ok {
		fmt.Printf("[%s]\n", str)
	}
	return nil
}

func cmdShow(c *cli.Context) error {
	dump := func(prefix trie.Key, val trie.Value) error {
		if val == nil {
			fmt.Printf("[%-8s]\n", prefix)
			return nil
		}
		fmt.Printf("[%-8s] %v\n", prefix, val)
		return nil
	}
	id := c.String("id")
	if len(id) == 0 {
		return stg.Traverse(dump, false, !c.Bool("all"))
	}
	for commit := head; commit != nil; commit = commit.Prev {
		if commit.MerkleTrie.MerkleHash().String() == id {
			return commit.MerkleTrie.Traverse(dump, false, true)
		}

	}
	return fmt.Errorf("commit noot found")
}

func cmdMerkle(c *cli.Context) error {
	fmt.Printf("%s\n", stg.MerkleHash())
	return nil
}

func cmdPrune(c *cli.Context) error {
	stg.Prune()
	fmt.Printf("pruned\n")
	return nil
}

func cmdCommit(c *cli.Context) error {
	msg := c.String("message")
	if len(msg) == 0 {
		return fmt.Errorf("no message specified")
	}
	h, err := stg.Commit(head, msg)
	if err != nil {
		return err
	}
	head = h
	return nil
}

func cmdReset(c *cli.Context) error {
	id := c.String("id")
	for commit := head; commit != nil; commit = commit.Prev {
		if commit.MerkleTrie.MerkleHash().String() != id {
			continue
		}
		head = commit
		stg = trie.NewStage(head.MerkleTrie)
		return nil
	}
	return fmt.Errorf("commit noot found")
}

func cmdLog(c *cli.Context) error {
	commitVisit := func(c *trie.Commit) {
		fmt.Printf("commit %s\n\n", c.MerkleTrie.MerkleHash())
		fmt.Printf("\t%s\n\n", c.Meta.(string))
	}

	trie.Log(head, commitVisit)
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
		if err := app.Run(append(os.Args[1:], strings.Split(text, " ")...)); err != nil {
			fmt.Printf("errot: %s\n", err)
		}

	}
	signal.Stop(sigs)
}
