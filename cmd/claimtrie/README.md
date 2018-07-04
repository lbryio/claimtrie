# ClaimTrie

coming soon

## Installation

coming soon

## Usage

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

## Running from Source

This project requires [Go v1.10](https://golang.org/doc/install) or higher.

``` bash
go get -u -v github.com/lbryio/claimtrie/cmd/claimtrie
go run ${GOPATH}/src/github.com/lbryio/claimtrie/cmd/claimtrie/main.go sh
```

## Examples

Adding claims.

``` bash
claimtrie > add-claim

claimtrie > show
<BestBlock: 0>
Hello   : {
  "Hash": "91185db0db792a6f6ad60e01e99e27f5263b8c3225137ff3b33bd5d3ebe197bd",
  "Tookover": 0,
  "BestClaim": {
    "OutPoint": "5fed5c7a39d47b4432b55d172b312df87142995e83ef28bf070859e94c916f30:48",
    "ClaimID": "911771619d4e6656bc0f08e1dfab5827756ed39d",
    "Amount": 44,
    "EffectiveAmount": 44,
    "Accepted": 0,
    "ActiveAt": 0
  },
  "Claims": [
    {
      "OutPoint": "5fed5c7a39d47b4432b55d172b312df87142995e83ef28bf070859e94c916f30:48",
      "ClaimID": "911771619d4e6656bc0f08e1dfab5827756ed39d",
      "Amount": 44,
      "EffectiveAmount": 44,
      "Accepted": 0,
      "ActiveAt": 0
    },
    {
      "OutPoint": "db993d544e0cbea83cff2465d3a8615cfc0750d39aa904a60e8fafab7c315a50:80",
      "ClaimID": "013df2835b1b8256ed019cc71df2dfb61fdce63c",
      "Amount": 10,
      "EffectiveAmount": 10,
      "Accepted": 0,
      "ActiveAt": 0
    }
  ],
  "Supports": []
}
claimtrie > commit
```

Commit another claim.

```bash
claimtrie > add-claim --amount 100
claimtrie > commit
```

Show logs

``` bash
claimtrie > log

height: 2, commit 9e2a2cf0e7f2a60e195ce46b261d6a953a3cbb68ef6b3274543ec8fdbf8a171b
height: 1, commit ce548249c28d61920d69ac759b82f53b5da52fa611f055c4f44c2d94703667a1
height: 0, commit 0000000000000000000000000000000000000000000000000000000000000001
```

Show current status.

```bash
claimtrie > show
<BestBlock: 2>
Hello   : {
  "Hash": "82629d2e9fb1eb8cc78e9d6712f217d5322f1cd9a3cdd15bf3923ee2d9376e94",
  "Tookover": 1,
  "BestClaim": {
    "OutPoint": "0f2fb103891bdf34344d34a64403537653d344558ebd3138b45e770585950d6e:110",
    "ClaimID": "128f9a84dddc87afdb747e04c6ce22726d2a90e7",
    "Amount": 100,
    "EffectiveAmount": 100,
    "Accepted": 1,
    "ActiveAt": 1
  },
  "Claims": [
    {
      "OutPoint": "5fed5c7a39d47b4432b55d172b312df87142995e83ef28bf070859e94c916f30:48",
      "ClaimID": "911771619d4e6656bc0f08e1dfab5827756ed39d",
      "Amount": 44,
      "EffectiveAmount": 44,
      "Accepted": 0,
      "ActiveAt": 0
    },
    {
      "OutPoint": "0f2fb103891bdf34344d34a64403537653d344558ebd3138b45e770585950d6e:110",
      "ClaimID": "128f9a84dddc87afdb747e04c6ce22726d2a90e7",
      "Amount": 100,
      "EffectiveAmount": 100,
      "Accepted": 1,
      "ActiveAt": 1
    },
    {
      "OutPoint": "db993d544e0cbea83cff2465d3a8615cfc0750d39aa904a60e8fafab7c315a50:80",
      "ClaimID": "013df2835b1b8256ed019cc71df2dfb61fdce63c",
      "Amount": 10,
      "EffectiveAmount": 10,
      "Accepted": 0,
      "ActiveAt": 0
    }
  ],
  "Supports": []
}

```

Reset the history to height 1.

``` bash
claimtrie > reset --height 1

claimtrie > show
<BestBlock: 1>
Hello   : {
  "Hash": "91185db0db792a6f6ad60e01e99e27f5263b8c3225137ff3b33bd5d3ebe197bd",
  "Tookover": 0,
  "BestClaim": {
    "OutPoint": "5fed5c7a39d47b4432b55d172b312df87142995e83ef28bf070859e94c916f30:48",
    "ClaimID": "911771619d4e6656bc0f08e1dfab5827756ed39d",
    "Amount": 44,
    "EffectiveAmount": 44,
    "Accepted": 0,
    "ActiveAt": 0
  },
  "Claims": [
    {
      "OutPoint": "5fed5c7a39d47b4432b55d172b312df87142995e83ef28bf070859e94c916f30:48",
      "ClaimID": "911771619d4e6656bc0f08e1dfab5827756ed39d",
      "Amount": 44,
      "EffectiveAmount": 44,
      "Accepted": 0,
      "ActiveAt": 0
    },
    {
      "OutPoint": "db993d544e0cbea83cff2465d3a8615cfc0750d39aa904a60e8fafab7c315a50:80",
      "ClaimID": "013df2835b1b8256ed019cc71df2dfb61fdce63c",
      "Amount": 10,
      "EffectiveAmount": 10,
      "Accepted": 0,
      "ActiveAt": 0
    }
  ],
  "Supports": []
}
claimtrie >
```

## Contributing

coming soon

## License

This project is MIT licensed.

## Security

We take security seriously. Please contact security@lbry.io regarding any security issues.
Our PGP key is [here](https://keybase.io/lbry/key.asc) if you need it.

## Contact

The primary contact for this project is [@roylee17](https://github.com/roylee) (roylee@lbry.io)