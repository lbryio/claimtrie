# ClaimTrie

coming soon

## Installation

``` bash
go get -u -v github.com/lbryio/claimtrie
```

## Usage

Refer to [claimtrie](https://github.com/lbryio/claimtrie/blob/master/cmd/claimtrie) for an interactive CLI tool as example.

``` block
NAME:
   claimtrie - A CLI tool for LBRY ClaimTrie

USAGE:
   main [global options] command [command options] [arguments...]

VERSION:
   0.0.1

COMMANDS:
     add-claim, ac      Claim a name.
     spend-claim, sc    Spend a Claim.
     update-claim, uc   Update a Claim.
     add-support, as    Support a Claim.
     spend-support, ss  Spend a specified Support.
     show, s            Show the status of nodes)
     merkle, m          Show the Merkle Hash of the ClaimTrie.
     commit, c          Commit the current changes to database.
     reset, r           Reset the Head commit and a specified commit (by Height).
     log, l             List the commits in the coommit database.
     ipmort, i          Import changes from datbase.
     load, ld           Load nodes from datbase.
     save, sv           Save nodes to datbase.
     erase              Erase datbase
     shell, sh          Enter interactive mode
     help, h            Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

## Running from Source

This project requires [Go v1.10](https://golang.org/doc/install) or higher.

## Examples


## Testing

``` bash
go test -v github.com/lbryio/claimtrie
gocov test -v github.com/lbryio/claimtrie 1>/dev/null
```

## Contributing

coming soon

## License

This project is MIT licensed.

## Security

We take security seriously. Please contact security@lbry.io regarding any security issues.
Our PGP key is [here](https://keybase.io/lbry/key.asc) if you need it.

## Contact

The primary contact for this project is [@roylee17](https://github.com/roylee17) (roylee@lbry.io) or [@lyoshenka](https://github.com/lyoshenka) (grin@lbry.io)