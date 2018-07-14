package claim

type cmdAddClaim struct {
	node  *Node
	claim *Claim
}

func (c cmdAddClaim) Execute() { c.node.claims = append(c.node.claims, c.claim) }
func (c cmdAddClaim) Undo()    { c.node.claims = c.node.claims.remove(c.claim.OutPoint) }

type cmdRemoveClaim struct {
	node  *Node
	claim *Claim
}

func (c cmdRemoveClaim) Execute() { c.node.claims = c.node.claims.remove(c.claim.OutPoint) }
func (c cmdRemoveClaim) Undo()    { c.node.claims = append(c.node.claims, c.claim) }

type cmdAddSupport struct {
	node    *Node
	support *Support
}

func (c cmdAddSupport) Execute() { c.node.supports = append(c.node.supports, c.support) }
func (c cmdAddSupport) Undo()    { c.node.supports = c.node.supports.remove(c.support.OutPoint) }

type cmdRemoveSupport struct {
	node    *Node
	support *Support
}

func (c cmdRemoveSupport) Execute() {
	c.node.supports = c.node.supports.remove(c.support.OutPoint)
}
func (c cmdRemoveSupport) Undo() { c.node.supports = append(c.node.supports, c.support) }

type cmdUpdateClaimActiveHeight struct {
	claim *Claim
	old   Height
	new   Height
}

func (c cmdUpdateClaimActiveHeight) Execute() { c.claim.ActiveAt = c.new }
func (c cmdUpdateClaimActiveHeight) Undo()    { c.claim.ActiveAt = c.old }

type cmdUpdateSupportActiveHeight struct {
	support *Support
	old     Height
	new     Height
}

func (c cmdUpdateSupportActiveHeight) Execute() { c.support.ActiveAt = c.new }
func (c cmdUpdateSupportActiveHeight) Undo()    { c.support.ActiveAt = c.old }

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
