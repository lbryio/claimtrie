package claim

type List []*Claim

type Comparator func(c *Claim) bool

func ByOP(op OutPoint) Comparator {
	return func(c *Claim) bool {
		return c.OutPoint == op
	}
}

func ByID(id ID) Comparator {
	return func(c *Claim) bool {
		return c.ID == id
	}
}

func Remove(l List, cmp Comparator) (List, *Claim) {
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

func Find(cmp Comparator, lists ...List) *Claim {
	for _, l := range lists {
		for _, v := range l {
			if cmp(v) {
				return v
			}
		}
	}
	return nil
}
