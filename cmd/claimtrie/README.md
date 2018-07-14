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

The following are the corresponding commands in the CLI to reproduce the test.
Note that when a Toakeover happens, the {TakeoverHeight, BestClaim(OutPoint:Indx)} is pushed to the BestClaims stack.
This provides sufficient info to update and backtracking the BestClaim at any height.

```block
claimtrie > c -ht 12
claimtrie > ac -a 10
claimtrie > c -ht 13
claimtrie > s

<ClaimTrie Height 13>
Hello   :  Height 13, 6f5970c9c13f00c77054d98e5b2c50b1a1bb723d91676cc03f984fac763ec6c3 BestClaims: {13, 31},

  C-2f9b2ca28b30c97122de584e8d784e784bc7bdfb43b3b4bf9de69fc31196571f:31  amt: 10   effamt: 10   accepted: 13   active: 13   id: ae8b3adc8c8b378c76eae12edf3878357b31c0eb  <B>

claimtrie > c -ht 1000
claimtrie > ac -a 20
claimtrie > c -ht 1001
claimtrie > s

<ClaimTrie Height 1001>
Hello   :  Height 1000, 6f5970c9c13f00c77054d98e5b2c50b1a1bb723d91676cc03f984fac763ec6c3 BestClaims: {13, 31},

  C-2f9b2ca28b30c97122de584e8d784e784bc7bdfb43b3b4bf9de69fc31196571f:31  amt: 10   effamt: 10   accepted: 13   active: 13   id: ae8b3adc8c8b378c76eae12edf3878357b31c0eb  <B>
  C-74fd7c15b445e7ac2de49a8058ad7dbb070ef31838345feff168dd40c2ef422e:46  amt: 20   effamt: 0    accepted: 1001  active: 1031  id: 2edc6338e9f3654f8b2b878817e029a2b2ecfa9e

claimtrie > c -ht 1009
claimtrie > as -a 14 -id ae8b3adc8c8b378c76eae12edf3878357b31c0eb
claimtrie > c -ht 1010
claimtrie > s

<ClaimTrie Height 1010>
Hello   :  Height 1010, 6f5970c9c13f00c77054d98e5b2c50b1a1bb723d91676cc03f984fac763ec6c3 BestClaims: {13, 31},

  C-2f9b2ca28b30c97122de584e8d784e784bc7bdfb43b3b4bf9de69fc31196571f:31  amt: 10   effamt: 24   accepted: 13   active: 13   id: ae8b3adc8c8b378c76eae12edf3878357b31c0eb  <B>
  C-74fd7c15b445e7ac2de49a8058ad7dbb070ef31838345feff168dd40c2ef422e:46  amt: 20   effamt: 0    accepted: 1001  active: 1031  id: 2edc6338e9f3654f8b2b878817e029a2b2ecfa9e

  S-087acb4f22ab6eb2e6c827624deab7beb02c190056376e0b3a3c3546b79bf216:22  amt: 14                accepted: 1010  active: 1010  id: ae8b3adc8c8b378c76eae12edf3878357b31c0eb

claimtrie > c -ht 1019
claimtrie > ac -a 50
claimtrie > c -ht 1020
claimtrie > s

<ClaimTrie Height 1020>
Hello   :  Height 1019, 6f5970c9c13f00c77054d98e5b2c50b1a1bb723d91676cc03f984fac763ec6c3 BestClaims: {13, 31},

  C-2f9b2ca28b30c97122de584e8d784e784bc7bdfb43b3b4bf9de69fc31196571f:31  amt: 10   effamt: 24   accepted: 13   active: 13   id: ae8b3adc8c8b378c76eae12edf3878357b31c0eb  <B>
  C-74fd7c15b445e7ac2de49a8058ad7dbb070ef31838345feff168dd40c2ef422e:46  amt: 20   effamt: 0    accepted: 1001  active: 1031  id: 2edc6338e9f3654f8b2b878817e029a2b2ecfa9e
  C-864dc683ed1fbe3c072c2387bca7d74e60c2505f1987054fd70955a2e1c9490b:11  amt: 50   effamt: 0    accepted: 1020  active: 1051  id: 0f2ba15a5c66b978df4a27f52525680a6009185e

  S-087acb4f22ab6eb2e6c827624deab7beb02c190056376e0b3a3c3546b79bf216:22  amt: 14                accepted: 1010  active: 1010  id: ae8b3adc8c8b378c76eae12edf3878357b31c0eb

claimtrie > c -ht 1031
claimtrie > s

<ClaimTrie Height 1031>
Hello   :  Height 1031, 6f5970c9c13f00c77054d98e5b2c50b1a1bb723d91676cc03f984fac763ec6c3 BestClaims: {13, 31},

  C-2f9b2ca28b30c97122de584e8d784e784bc7bdfb43b3b4bf9de69fc31196571f:31  amt: 10   effamt: 24   accepted: 13   active: 13   id: ae8b3adc8c8b378c76eae12edf3878357b31c0eb  <B>
  C-74fd7c15b445e7ac2de49a8058ad7dbb070ef31838345feff168dd40c2ef422e:46  amt: 20   effamt: 20   accepted: 1001  active: 1031  id: 2edc6338e9f3654f8b2b878817e029a2b2ecfa9e
  C-864dc683ed1fbe3c072c2387bca7d74e60c2505f1987054fd70955a2e1c9490b:11  amt: 50   effamt: 0    accepted: 1020  active: 1051  id: 0f2ba15a5c66b978df4a27f52525680a6009185e

  S-087acb4f22ab6eb2e6c827624deab7beb02c190056376e0b3a3c3546b79bf216:22  amt: 14                accepted: 1010  active: 1010  id: ae8b3adc8c8b378c76eae12edf3878357b31c0eb

claimtrie > c -ht 1039
claimtrie > ac -a 300
claimtrie > c -ht 1040
claimtrie > s

<ClaimTrie Height 1040>
Hello   :  Height 1039, 6f5970c9c13f00c77054d98e5b2c50b1a1bb723d91676cc03f984fac763ec6c3 BestClaims: {13, 31},

  C-2f9b2ca28b30c97122de584e8d784e784bc7bdfb43b3b4bf9de69fc31196571f:31  amt: 10   effamt: 24   accepted: 13   active: 13   id: ae8b3adc8c8b378c76eae12edf3878357b31c0eb  <B>
  C-74fd7c15b445e7ac2de49a8058ad7dbb070ef31838345feff168dd40c2ef422e:46  amt: 20   effamt: 20   accepted: 1001  active: 1031  id: 2edc6338e9f3654f8b2b878817e029a2b2ecfa9e
  C-864dc683ed1fbe3c072c2387bca7d74e60c2505f1987054fd70955a2e1c9490b:11  amt: 50   effamt: 0    accepted: 1020  active: 1051  id: 0f2ba15a5c66b978df4a27f52525680a6009185e
  C-26292b27122d04d08fee4e4cc5a5f94681832204cc29d61039c09af9a5298d16:22  amt: 300  effamt: 0    accepted: 1040  active: 1072  id: 270496c0710e525156510e60e4be2ffa6fe2f507

  S-087acb4f22ab6eb2e6c827624deab7beb02c190056376e0b3a3c3546b79bf216:22  amt: 14                accepted: 1010  active: 1010  id: ae8b3adc8c8b378c76eae12edf3878357b31c0eb

claimtrie > c -ht 1051
claimtrie > s

<ClaimTrie Height 1051>
Hello   :  Height 1051, 68dff86c9450e3cf96570f31b6ad8f8d35ae0cbce6cdcb3761910e25a815ee8b BestClaims: {13, 31}, {1051, 22},

  C-2f9b2ca28b30c97122de584e8d784e784bc7bdfb43b3b4bf9de69fc31196571f:31  amt: 10   effamt: 24   accepted: 13   active: 13   id: ae8b3adc8c8b378c76eae12edf3878357b31c0eb
  C-74fd7c15b445e7ac2de49a8058ad7dbb070ef31838345feff168dd40c2ef422e:46  amt: 20   effamt: 20   accepted: 1001  active: 1001  id: 2edc6338e9f3654f8b2b878817e029a2b2ecfa9e
  C-864dc683ed1fbe3c072c2387bca7d74e60c2505f1987054fd70955a2e1c9490b:11  amt: 50   effamt: 50   accepted: 1020  active: 1020  id: 0f2ba15a5c66b978df4a27f52525680a6009185e
  C-26292b27122d04d08fee4e4cc5a5f94681832204cc29d61039c09af9a5298d16:22  amt: 300  effamt: 300  accepted: 1040  active: 1040  id: 270496c0710e525156510e60e4be2ffa6fe2f507  <B>

  S-087acb4f22ab6eb2e6c827624deab7beb02c190056376e0b3a3c3546b79bf216:22  amt: 14                accepted: 1010  active: 1010  id: ae8b3adc8c8b378c76eae12edf3878357b31c0eb
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