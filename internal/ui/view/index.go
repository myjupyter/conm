package view

import (
	"slices"
	"strings"
)

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
	query   string
	visible []Ref[K]
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
	x.refilter()
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

func (x *index[K]) Searching() bool {
	return len(searchTerms(x.query)) > 0
}

func (x *index[K]) Search(query string) {
	x.query = query
	x.refilter()
}

func (x *index[K]) ClearSearch() {
	x.Search("")
}

func (x *index[K]) Focus(i int) (int, bool) {
	ref, ok := x.RefAt(i)
	if !ok {
		return 0, false
	}

	x.query = ""
	x.active = ref.Kind
	x.refilter()

	row := slices.Index(x.visible, ref)
	if row < 0 {
		return 0, false
	}

	return row, true
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
	terms := searchTerms(x.query)

	x.visible = x.visible[:0]
	for _, e := range x.entries {
		if len(terms) == 0 {
			if e.ref.Kind == x.active {
				x.visible = append(x.visible, e.ref)
			}

			continue
		}

		if e.matches(terms) {
			x.visible = append(x.visible, e.ref)
		}
	}
}

func (e entry[K]) matches(terms []string) bool {
	for _, term := range terms {
		found := slices.ContainsFunc(e.fields, func(field string) bool {
			return strings.Contains(field, term)
		})
		if !found {
			return false
		}
	}

	return true
}

func searchTerms(query string) []string {
	return strings.Fields(strings.ToLower(query))
}

func searchable(values ...string) []string {
	fields := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.ToLower(strings.TrimSpace(value)); value != "" {
			fields = append(fields, value)
		}
	}

	return fields
}
