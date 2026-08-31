package view

import "slices"

type Ref[K comparable] struct {
	Kind  K
	Index int
}

type entry[K comparable] struct {
	ref    Ref[K]
	fields []string
}

type index[K comparable] struct {
	order   []K
	active  K
	entries []entry[K]
	search  searcher

	query    string
	filtered bool
	visible  []Ref[K]
}

func (x *index[K]) register(kind K) bool {
	if slices.Contains(x.order, kind) {
		return false
	}

	if len(x.order) == 0 {
		x.active = kind
	}
	x.order = append(x.order, kind)

	return true
}

func (x *index[K]) reset(entries []entry[K]) {
	x.entries = entries

	docs := make([]document, 0, len(entries))
	for i, e := range entries {
		docs = append(docs, document{id: i, fields: e.fields})
	}
	x.searcher().Reindex(docs)

	x.refilter()
}

func (x *index[K]) searcher() searcher {
	if x.search == nil {
		x.search = newSearcher()
	}

	return x.search
}

func (x *index[K]) Kinds() []K {
	return slices.Clone(x.order)
}

func (x *index[K]) Active() K {
	return x.active
}

func (x *index[K]) SetActive(kind K) bool {
	if !slices.Contains(x.order, kind) {
		return false
	}

	x.active = kind
	x.refilter()

	return true
}

func (x *index[K]) NextKind() bool {
	return x.stepKind(1)
}

func (x *index[K]) PrevKind() bool {
	return x.stepKind(-1)
}

func (x *index[K]) Total() int {
	return len(x.entries)
}

func (x *index[K]) Len() int {
	return len(x.visible)
}

func (x *index[K]) RefAt(i int) (Ref[K], bool) {
	if i < 0 || i >= len(x.visible) {
		var zero Ref[K]
		return zero, false
	}

	return x.visible[i], true
}

func (x *index[K]) Query() string {
	return x.query
}

func (x *index[K]) Filtered() bool {
	return x.filtered
}

func (x *index[K]) Search(query string) {
	x.query = query
	x.refilter()
}

func (x *index[K]) ClearSearch() {
	x.Search("")
}

func (x *index[K]) stepKind(step int) bool {
	if len(x.order) < 2 {
		return false
	}

	i := slices.Index(x.order, x.active)
	if i < 0 {
		return false
	}

	x.active = x.order[(i+step+len(x.order))%len(x.order)]
	x.refilter()

	return true
}

func (x *index[K]) refilter() {
	hits, filtered := x.searcher().Search(x.query)
	x.filtered = filtered
	x.visible = x.visible[:0]

	if !filtered {
		for _, e := range x.entries {
			if e.ref.Kind == x.active {
				x.visible = append(x.visible, e.ref)
			}
		}

		return
	}

	for _, id := range hits {
		if id < 0 || id >= len(x.entries) {
			continue
		}

		if e := x.entries[id]; e.ref.Kind == x.active {
			x.visible = append(x.visible, e.ref)
		}
	}
}
