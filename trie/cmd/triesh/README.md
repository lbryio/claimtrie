# Triesh

An example Key-Value store to excercise the merkletree package

Currently, it's only in-memory.

## Installation

This project requires [Go v1.10](https://golang.org/doc/install) or higher.

``` bash
go get -v github.com/lbryio/trie
```

## Usage

Adding values.

``` bloocks
triesh > u -k alex -v lion
alex=lion
triesh > u -k al -v tiger
al=tiger
triesh > u -k tess -v dolphin
tess=dolphin
triesh > u -k bob -v pig
bob=pig
triesh > u -k ted -v do
ted=do
triesh > u -k ted -v dog
ted=dog
```

Showing Merkle Hash.

``` blocks
triesh > merkle
bfa2927b147161146411b7f6187e1ed0c08c3dc19b200550c3458d44c0032285

triesh > u -k teddy -v bear
teddy=bear

triesh > merkle
94831650b8bf76d579ca4eda1cb35861c6f5c88eb4f5b089f60fe687defe8f3d
```

Showing all values.

``` blocks
triesh > s
[al      ] tiger
[alex    ] lion
[bob     ] pig
[ted     ] dog
[teddy   ] bear
[tess    ] dolphin
```

Showing all values and link nodes.

``` bloocks
triesh > s -a
[a       ]
[al      ] tiger
[ale     ]
[alex    ] lion
[b       ]
[bo      ]
[bob     ] pig
[t       ]
[te      ]
[ted     ] dog
[tedd    ]
[teddy   ] bear
[tes     ]
[tess    ] dolphin
```

Deleting values (setting key to nil / "").

``` blocks
triesh > u -k al
al=
triesh > u -k alex
alex=
```

Updating Values.

``` blocks
triesh > u -k bob -v cat
bob=cat
```

Showing all nodes, include non-pruned link nodes"

``` blocks
triesh > s -a
[a       ]
[al      ]
[ale     ]
[alex    ]
[b       ]
[bo      ]
[bob     ] cat
[t       ]
[te      ]
[ted     ] dog
[tedd    ]
[teddy   ] bear
[tes     ]
[tess    ] dolphin

```

Calculate Merkle Hash.

``` blocks
triesh > merkle
c2fdce68a30e3cabf6efb3b7ebfd32afdaf09f9ebd062743fe91e181f682252b
```

Prune link nodes that do not reach to any values.

``` blocks
triesh > p
pruned
```

Show pruned Trie and caculate the Merkle Hash again.

``` blocks
triesh > s -a
[b       ]
[bo      ]
[bob     ] cat
[t       ]
[te      ]
[ted     ] dog
[tedd    ]
[teddy   ] bear
[tes     ]
[tess    ] dolphin

triesh > merkle
c2fdce68a30e3cabf6efb3b7ebfd32afdaf09f9ebd062743fe91e181f682252b
```

## Running from Source

``` bash
cd $(go env GOPATH)/src/github.com/lbryio/trie
go run cmd/triesh/*.go sh
```

## Contributing

coming soon

## License

This project is MIT licensed.

## Security

We take security seriously. Please contact security@lbry.io regarding any security issues.
Our PGP key is [here](https://keybase.io/lbry/key.asc) if you need it.

## Contact

The primary contact for this project is [@lyoshenka](https://github.com/lyoshenka) (grin@lbry.io)