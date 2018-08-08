# ClaimTrie

coming soon

## Installation

``` bash
go get -u github.com/lbryio/claimtrie/cmd/claimtrie
```

## Usage

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

``` bash
go run ${GOPATH}/src/github.com/lbryio/claimtrie/cmd/claimtrie/main.go sh
```

Or build and run as executable

``` bash
go build ${GOPATH}/src/github.com/lbryio/claimtrie/cmd/claimtrie/main.go
./claimtrie
```

## Examples

Import (replay) claim scripts saved from the LBRY mainnet.
We need to run the lbrycrd.go to produce the dataset.

Let it run for 15 - 30 mins until we sync 400K blocks of claim scripts to test with.
(It's currently hard coded to exit one 400k blocks are collected.)

``` block
go get -u github.com/btcsuite/btcd
cd $GOPATH/src/github.com/btcsuite/btcd
git remote add lbryio git@github.com:lbryio/lbrycrd.go
git fetch --all
git checkout -b dev -m lbryio/dev
go build -i && ./btcd
```

Enter interactive mode

``` bash
go run ${GOPATH}/src/github.com/lbryio/claimtrie/cmd/claimtrie/main.go sh

opened "/Users/roylee17/Library/Application Support/Lbrycrd.go/data/trie.db"
opened "/Users/roylee17/Library/Application Support/Lbrycrd.go/data/nm.db"
opened "/Users/roylee17/Library/Application Support/Lbrycrd.go/data/commit.db"

claimtrie >
```

Note: the storage part was just brought up. In case it missed something, or the program was shutdown cleanly, try erase the database to clean the status.
(It doesn't modify the ClaimScript we saved from BTCD.)

``` block
claimtrie > erase
```

Import claim scripts saved from lbrycrd.go (-v with verbose output)

``` block
claimtrie > i -ht 400000 -v -c

2018/08/07 19:43:14    102 +C 67ad533eb2676c9d36bfa100092af5358de747e08ef928c0c54a8b3891c2b76b:1 bdb4df1b86ada117a61e1737c6d2604e940f1fb4     50000000 [mindblown]
2018/08/07 19:43:14    125 +C 9b4afb7edf206f7d2fbd353add4a471887c92dba97145ee550ac06a4fa73bcd1:1 b48341bb3a5470abe3b661e40ad187ac61c3301a    300000000 [mindblown]
2018/08/07 19:43:14    984 +C 720ad1906617112924845a775e28e5176d491ca289992dc64b9a38d34b04ce08:0 982b7b306122e3c38c826549ca1f2370d3113ca1    100000000 [keynesvhayek]
2018/08/07 19:43:14    986 +C 90f3eb8682ee620bc5fcd7f37a5a3c83b38863c316797c36ec5022b81fc32331:1 61c91810c863394bd9b436c6cbfd19a5d8b9f3c4    100000000 [keynesvhayek]

  ...

```

show the commits

``` block
claimtrie >log

abbb24f008ec4a6b9c8a972cdaa779fb2ee10d763191c824848460ec3a5e3f21 at 400000
abbb24f008ec4a6b9c8a972cdaa779fb2ee10d763191c824848460ec3a5e3f21 at 399997
a4fe8f67ca7510daaf68a63bd2de3693bf5fd5504a38c600d52bb6e57547ce18 at 399996
a4fe8f67ca7510daaf68a63bd2de3693bf5fd5504a38c600d52bb6e57547ce18 at 399993
134672de8c8ef824c9ed76efcae2924ac20dc0627a123d0c746c10d0efa89203 at 399992
134672de8c8ef824c9ed76efcae2924ac20dc0627a123d0c746c10d0efa89203 at 399885
134672de8c8ef824c9ed76efcae2924ac20dc0627a123d0c746c10d0efa89203 at 399884
134672de8c8ef824c9ed76efcae2924ac20dc0627a123d0c746c10d0efa89203 at 399883
134672de8c8ef824c9ed76efcae2924ac20dc0627a123d0c746c10d0efa89203 at 399882
134672de8c8ef824c9ed76efcae2924ac20dc0627a123d0c746c10d0efa89203 at 399881
...
```

Show merkle hash of current commit.

``` block
claimtrie> merkle

abbb24f008ec4a6b9c8a972cdaa779fb2ee10d763191c824848460ec3a5e3f21 at 400000
```

Show particular nodes at specific height.

``` block

claimtrie > s -n @jack -ht 300000

<ClaimTrie Height 400000 >

[@jack]  Height 300000, d0fa172e082e2c7d45df5e514a213a812812be04fb5d42c14b2711f88d65fc8d Tookover: 156270 Next: 415690

  C cc3ad268eb9662046585975c2da4c214a0b7d11d78f5bd154ce9e74e5a868633:0   id: b8e04a32cdb977e96b80ff24324b1771c198038c accepted: 152716  active: 152716, amt:     99983800  effamt:     99983800
  C 4c40d17a13e21661821b746e0c8b3771101fe7ad70444fb295feaa653b066f11:0   id: 413677be053754e735fea95dd8a1db65d63dcc01 accepted: 157311  active: 157343, amt:     99983800  effamt:     99983800
  C dced82824e8d8cafc6d0e8ab58e80b5cacb8ff8317aadeb0ae0f34b6fac0432f:0   id: 251305ca93d4dbedb50dceb282ebcb7b07b7ac65 accepted: 191763  active: 156270, amt:    999967600  effamt: 1001311967600  <B>
  C a67db7518d32e001343ffa1d83ca4bf57b20584c8faaa14b4dba3ae99d6a4caa:0   id: dc4c3f7f3f80a13349d0205c177bd940e83c8a42 accepted: 198608  active: 199931, amt:     10000000  effamt:     10000000
  C 0f56714baafa0c3639f60d197927f8efca06a6bace0d5b2b11c13f77a26db968:0   id: 112d4619c2ec119105ce25737e8b3bdc1efc29df accepted: 200056  active: 201424, amt:      1000000  effamt:      1000000
  C 00e8ee6d1fc3d234f95e7f532f27a6336019cff6e5d9a156725b6d37c5c95f46:0   id: e12bb13d7e78f2a573c96c4229ac1b54b53f2186 accepted: 240389  active: 243017, amt:     10000000  effamt:     10000000

  S 46c4fcb0f30d05e7780382a4fe999f69211a208318df592f162cef0837cf3903:0   id: 251305ca93d4dbedb50dceb282ebcb7b07b7ac65 accepted: 156479  active: 156479, amt: 500000000000  effamt:            0
  S 361b37b3d16c049dd7efeda5e02b90e0085ad03b49c97306810ac018eeeeafa3:0   id: 251305ca93d4dbedb50dceb282ebcb7b07b7ac65 accepted: 156483  active: 156483, amt: 500000000000  effamt:            0
  S 8b20b97b88ac17a24fbb7736f479f92c0ac727cda946bf76444f194bbeebada6:0   id: 251305ca93d4dbedb50dceb282ebcb7b07b7ac65 accepted: 202938  active: 202938, amt:    100000000  effamt:            0
  S 7d1951cac00feac7d7ea7569270a0c392e48937d23cf3a395c074b5d8c724785:0   id: 251305ca93d4dbedb50dceb282ebcb7b07b7ac65 accepted: 216968  active: 216968, amt:      1000000  effamt:            0
  S d050fb0e5a7366ab648c5b0dea057fc889b7a37c305567f17174241441d20895:0   id: 251305ca93d4dbedb50dceb282ebcb7b07b7ac65 accepted: 216972  active: 216972, amt:      1000000  effamt:            0
  S 7710f9d6b69c67002a8ea390823d773d5c875f4a1471a90c75ee00b10102e867:0   id: 251305ca93d4dbedb50dceb282ebcb7b07b7ac65 accepted: 223846  active: 223846, amt:    100000000  effamt:            0
  S 10e3f76735f483bd9f912080cf4952f7d512b16c20e242d73bd87b159cf6bb7b:0   id: 251305ca93d4dbedb50dceb282ebcb7b07b7ac65 accepted: 228125  active: 228125, amt:    100000000  effamt:            0
  S fe53337b37c1b1ac31eddec3cdc08e7467642529c18d92a18629f74ed1d6afae:0   id: 251305ca93d4dbedb50dceb282ebcb7b07b7ac65 accepted: 228163  active: 228163, amt:     10000000  effamt:            0
```

Show a node at specified height, and also dump the changes replayed to produce it.

``` block
claimtrie > s -n @jack -ht 200000 -d

<ClaimTrie Height 400000 >

[@jack]  Height 200000, d0fa172e082e2c7d45df5e514a213a812812be04fb5d42c14b2711f88d65fc8d Tookover: 156270 Next: 415690

  C cc3ad268eb9662046585975c2da4c214a0b7d11d78f5bd154ce9e74e5a868633:0   id: b8e04a32cdb977e96b80ff24324b1771c198038c accepted: 152716  active: 152716, amt:     99983800  effamt:     99983800
  C 4c40d17a13e21661821b746e0c8b3771101fe7ad70444fb295feaa653b066f11:0   id: 413677be053754e735fea95dd8a1db65d63dcc01 accepted: 157311  active: 157343, amt:     99983800  effamt:     99983800
  C dced82824e8d8cafc6d0e8ab58e80b5cacb8ff8317aadeb0ae0f34b6fac0432f:0   id: 251305ca93d4dbedb50dceb282ebcb7b07b7ac65 accepted: 191763  active: 156270, amt:    999967600  effamt: 1000999967600  <B>
  C a67db7518d32e001343ffa1d83ca4bf57b20584c8faaa14b4dba3ae99d6a4caa:0   id: dc4c3f7f3f80a13349d0205c177bd940e83c8a42 accepted: 198608  active: 199931, amt:     10000000  effamt:     10000000

  S 46c4fcb0f30d05e7780382a4fe999f69211a208318df592f162cef0837cf3903:0   id: 251305ca93d4dbedb50dceb282ebcb7b07b7ac65 accepted: 156479  active: 156479, amt: 500000000000  effamt:            0
  S 361b37b3d16c049dd7efeda5e02b90e0085ad03b49c97306810ac018eeeeafa3:0   id: 251305ca93d4dbedb50dceb282ebcb7b07b7ac65 accepted: 156483  active: 156483, amt: 500000000000  effamt:            0

chgs[0] 152510 +C c768895f52035a4fafee54ff30795d1e0ceb5ed0f46dae4c3da6253b9d11c525:0 0000000000000000000000000000000000000000    100000000 [@jack]
chgs[1] 152716 -C c768895f52035a4fafee54ff30795d1e0ceb5ed0f46dae4c3da6253b9d11c525:0 0000000000000000000000000000000000000000            0 [@jack]
chgs[2] 152716 +U cc3ad268eb9662046585975c2da4c214a0b7d11d78f5bd154ce9e74e5a868633:0 b8e04a32cdb977e96b80ff24324b1771c198038c     99983800 [@jack]
chgs[3] 156157 +C 04a5901f845c7f21911538159d5fe541d2b4480336bb29115b805bcbab032bff:0 0000000000000000000000000000000000000000   1000000000 [@jack]
chgs[4] 156479 +S 46c4fcb0f30d05e7780382a4fe999f69211a208318df592f162cef0837cf3903:0 251305ca93d4dbedb50dceb282ebcb7b07b7ac65 500000000000 [@jack]
chgs[5] 156483 +S 361b37b3d16c049dd7efeda5e02b90e0085ad03b49c97306810ac018eeeeafa3:0 251305ca93d4dbedb50dceb282ebcb7b07b7ac65 500000000000 [@jack]
chgs[6] 157311 +C 3f275be2b58bc9a66b43d79d5b819059c1ea4c3132c57aef0b89a022aac2f09b:0 0000000000000000000000000000000000000000    100000000 [@jack]
chgs[7] 157311 -C 3f275be2b58bc9a66b43d79d5b819059c1ea4c3132c57aef0b89a022aac2f09b:0 0000000000000000000000000000000000000000            0 [@jack]
chgs[8] 157311 +U 4c40d17a13e21661821b746e0c8b3771101fe7ad70444fb295feaa653b066f11:0 413677be053754e735fea95dd8a1db65d63dcc01     99983800 [@jack]
chgs[9] 191752 -C 04a5901f845c7f21911538159d5fe541d2b4480336bb29115b805bcbab032bff:0 0000000000000000000000000000000000000000            0 [@jack]
chgs[10] 191752 +U 9e4588172b15cfc922aeed4a426acc17914552224c399e6179bb4d896ba77b4b:0 251305ca93d4dbedb50dceb282ebcb7b07b7ac65    999983800 [@jack]
chgs[11] 191763 -C 9e4588172b15cfc922aeed4a426acc17914552224c399e6179bb4d896ba77b4b:0 0000000000000000000000000000000000000000            0 [@jack]
chgs[12] 191763 +U dced82824e8d8cafc6d0e8ab58e80b5cacb8ff8317aadeb0ae0f34b6fac0432f:0 251305ca93d4dbedb50dceb282ebcb7b07b7ac65    999967600 [@jack]
chgs[13] 198607 +C 439e17b2370bed8df0f00dba4a662f0634653a136e6b8e7c6b76ba071870f013:0 0000000000000000000000000000000000000000     10000000 [@jack]
chgs[14] 198608 -C 439e17b2370bed8df0f00dba4a662f0634653a136e6b8e7c6b76ba071870f013:0 0000000000000000000000000000000000000000            0 [@jack]
chgs[15] 198608 +U a67db7518d32e001343ffa1d83ca4bf57b20584c8faaa14b4dba3ae99d6a4caa:0 dc4c3f7f3f80a13349d0205c177bd940e83c8a42     10000000 [@jack]
```

Show all nodes (try with lower height, or it takes time to print) at specific height.

``` block
s -a -ht 100000
```

Reset the ClaimTrie to an earlier height.

``` block
r -ht 150000
```

Save the state to db.

``` block
sv
```

quit

``` block
q
```

Enter interactive mode.

``` bash
go run ${GOPATH}/src/github.com/lbryio/claimtrie/cmd/claimtrie/main.go sh
```

Load the state back.

``` block
claimtrie > ld
4891 of commits loaded. Head: 150000
194171 of nodes loaded.
Trie root: 1527b999092f171186ac6690f550bad939fc44e8b3128d1ea11ed6c3b74d8628.
```

The following are the corresponding commands in the CLI to reproduce the test.
Note that when a Toakeover happens, the {TakeoverHeight, BestClaim(OutPoint:Indx)} is pushed to the BestClaims stack.
This provides sufficient info to update and backtracking the BestClaim at any height.

``` block
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

The primary contact for this project is [@roylee17](https://github.com/roylee17) (roylee@lbry.io) or [@lyoshenka](https://github.com/lyoshenka) (grin@lbry.io)