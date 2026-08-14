package secret

import "sort"

type Usage struct {
	Spec string
	Refs int
}

type usageSet struct {
	refs map[string]int
}

func (u *usageSet) track(spec string) {
	if u.refs == nil {
		u.refs = make(map[string]int)
	}
	u.refs[spec]++
}

func (u *usageSet) untrack(spec string) {
	if u.refs[spec] <= 1 {
		delete(u.refs, spec)
		return
	}
	u.refs[spec]--
}

func (u *usageSet) list() []Usage {
	usages := make([]Usage, 0, len(u.refs))
	for spec, refs := range u.refs {
		usages = append(usages, Usage{Spec: spec, Refs: refs})
	}

	sort.Slice(usages, func(i, j int) bool { return usages[i].Spec < usages[j].Spec })

	return usages
}
