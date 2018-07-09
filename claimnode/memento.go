package claimnode

import "github.com/lbryio/claimtrie/claim"

type cmdAddClaim struct {
	node  *Node
	claim *claim.Claim
}

func (c cmdAddClaim) Execute() { c.node.claims[c.claim.OutPoint] = c.claim }
func (c cmdAddClaim) Undo()    { delete(c.node.claims, c.claim.OutPoint) }

type cmdRemoveClaim struct {
	node  *Node
	claim *claim.Claim
}

func (c cmdRemoveClaim) Execute() { delete(c.node.claims, c.claim.OutPoint) }
func (c cmdRemoveClaim) Undo()    { c.node.claims[c.claim.OutPoint] = c.claim }

type cmdAddSupport struct {
	node    *Node
	support *claim.Support
}

func (c cmdAddSupport) Execute() { c.node.supports[c.support.OutPoint] = c.support }
func (c cmdAddSupport) Undo()    { delete(c.node.supports, c.support.OutPoint) }

type cmdRemoveSupport struct {
	node    *Node
	support *claim.Support
}

func (c cmdRemoveSupport) Execute() { delete(c.node.supports, c.support.OutPoint) }
func (c cmdRemoveSupport) Undo()    { c.node.supports[c.support.OutPoint] = c.support }

type cmdUpdateClaimActiveHeight struct {
	claim *claim.Claim
	old   claim.Height
	new   claim.Height
}

func (c cmdUpdateClaimActiveHeight) Execute() { c.claim.ActiveAt = c.new }
func (c cmdUpdateClaimActiveHeight) Undo()    { c.claim.ActiveAt = c.old }

type cmdUpdateSupportActiveHeight struct {
	support *claim.Support
	old     claim.Height
	new     claim.Height
}

func (c cmdUpdateSupportActiveHeight) Execute() { c.support.ActiveAt = c.new }
func (c cmdUpdateSupportActiveHeight) Undo()    { c.support.ActiveAt = c.old }

type updateNodeBestClaim struct {
	node   *Node
	height claim.Height
	old    *claim.Claim
	new    *claim.Claim
}

func (c updateNodeBestClaim) Execute() {
	c.node.bestClaims[c.height] = c.new
	if c.node.bestClaims[c.height] == nil {
		delete(c.node.bestClaims, c.height)
	}
}

func (c updateNodeBestClaim) Undo() {
	c.node.bestClaims[c.height] = c.old
	if c.node.bestClaims[c.height] == nil {
		delete(c.node.bestClaims, c.height)
	}
}
