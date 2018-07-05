package claimtrie

type cmdAddClaim struct {
	node  *Node
	claim *Claim
}

func (c cmdAddClaim) Execute() { c.node.claims[c.claim.op] = c.claim }
func (c cmdAddClaim) Undo()    { delete(c.node.claims, c.claim.op) }

type cmdRemoveClaim struct {
	node  *Node
	claim *Claim
}

func (c cmdRemoveClaim) Execute() { delete(c.node.claims, c.claim.op) }
func (c cmdRemoveClaim) Undo()    { c.node.claims[c.claim.op] = c.claim }

type cmdAddSupport struct {
	node    *Node
	support *Support
}

func (c cmdAddSupport) Execute() { c.node.supports[c.support.op] = c.support }
func (c cmdAddSupport) Undo()    { delete(c.node.supports, c.support.op) }

type cmdRemoveSupport struct {
	node    *Node
	support *Support
}

func (c cmdRemoveSupport) Execute() { delete(c.node.supports, c.support.op) }
func (c cmdRemoveSupport) Undo()    { c.node.supports[c.support.op] = c.support }

type cmdUpdateClaimActiveHeight struct {
	claim *Claim
	old   Height
	new   Height
}

func (c cmdUpdateClaimActiveHeight) Execute() { c.claim.activeAt = c.new }
func (c cmdUpdateClaimActiveHeight) Undo()    { c.claim.activeAt = c.old }

type cmdUpdateSupportActiveHeight struct {
	support *Support
	old     Height
	new     Height
}

func (c cmdUpdateSupportActiveHeight) Execute() { c.support.activeAt = c.new }
func (c cmdUpdateSupportActiveHeight) Undo()    { c.support.activeAt = c.old }

type updateNodeBestClaim struct {
	node   *Node
	height Height
	old    *Claim
	new    *Claim
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
