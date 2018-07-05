# ClaimTrie

coming soon

## Installation

coming soon

## Usage

coming soon

## Running from Source

This project requires [Go v1.10](https://golang.org/doc/install) or higher.

``` bash
go get -u -v github.com/lbryio/claimtrie
```

## Examples

Refer to [claimtrie](https://github.com/lbryio/claimtrie/blob/master/cmd/claimtrie) for an interactive CLI tool.

``` bash
NAME:
   claimtrie - A CLI tool for ClaimTrie

USAGE:
   main [global options] command [command options] [arguments...]

VERSION:
   0.0.1

COMMANDS:
     add-claim, ac      Claim a name with specified amount. (outPoint is generated randomly, if unspecified)
     add-support, as    Add support to a specified Claim. (outPoint is generated randomly, if unspecified)
     spend-claim, sc    Spend a specified Claim.
     spend-support, ss  Spend a specified Support.
     show, s            Show the Key-Value pairs of the Stage or specified commit. (links nodes are showed if -a is also specified)
     merkle, m          Show the Merkle Hash of the Stage.
     commit, c          Commit the current Stage to commit database.
     reset, r           Reset the Stage to a specified commit.
     log, l             List the commits in the coommit database.
     shell, sh          Enter interactive mode
     help, h            Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

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

The primary contact for this project is [@roylee17](https://github.com/roylee17) (roylee@lbry.io)