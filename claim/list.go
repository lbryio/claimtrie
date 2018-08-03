package claim

type list []*Claim

type comparator func(c *Claim) bool

func byOP(op OutPoint) comparator {
	return func(c *Claim) bool {
		return c.OutPoint == op
	}
}

func byID(id ID) comparator {
	return func(c *Claim) bool {
		return c.ID == id
	}
}

func remove(l list, cmp comparator) (list, *Claim) {
	last := len(l) - 1
	for i, v := range l {
		if !cmp(v) {
			continue
		}
		removed := l[i]
		l[i] = l[last]
		l[last] = nil
		return l[:last], removed
	}
	return l, nil
}

func find(cmp comparator, lists ...list) *Claim {
	for _, l := range lists {
		for _, v := range l {
			if cmp(v) {
				return v
			}
		}
	}
	return nil
}
