package claim

import (
	"github.com/lbryio/claimtrie/memento"
)

type nodeBuildable interface {
	build() Node

	setMemento(mem memento.Memento) nodeBuildable
	setBestClaims(claims ...Claim) nodeBuildable
	setClaims(claims ...Claim) nodeBuildable
	setSupports(supports ...Support) nodeBuildable
	setHeight(h Height) nodeBuildable
	setUpdateNext(b bool) nodeBuildable
}

func newNodeBuilder() nodeBuildable {
	return &nodeBuilder{n: NewNode()}
}

type nodeBuilder struct{ n *Node }

func (nb *nodeBuilder) build() Node {
	return *nb.n
}

func (nb *nodeBuilder) setMemento(mem memento.Memento) nodeBuildable {
	nb.n.mem = mem
	return nb
}

func (nb *nodeBuilder) setHeight(h Height) nodeBuildable {
	nb.n.height = h
	return nb
}

func (nb *nodeBuilder) setUpdateNext(b bool) nodeBuildable {
	nb.n.updateNext = b
	return nb
}

func (nb *nodeBuilder) setBestClaims(claims ...Claim) nodeBuildable {
	for i := range claims {
		c := claims[i] // Copy value, instead of holding reference to the slice.
		nb.n.bestClaims[c.ActiveAt] = &c
	}
	return nb
}

func (nb *nodeBuilder) setClaims(claims ...Claim) nodeBuildable {
	for i := range claims {
		c := claims[i] // Copy value, instead of holding reference to the slice.
		nb.n.claims = append(nb.n.claims, &c)
	}
	return nb
}

func (nb *nodeBuilder) setSupports(supports ...Support) nodeBuildable {
	for i := range supports {
		s := supports[i] // Copy value, instead of holding reference to the slice.
		nb.n.supports = append(nb.n.supports, &s)
	}
	return nb
}
