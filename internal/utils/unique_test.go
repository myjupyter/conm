package utils

import (
	"slices"
	"testing"
)

type row struct {
	id   int
	name string
}

func TestFirstDuplicate(t *testing.T) {
	byName := func(r row) string { return r.name }

	tests := []struct {
		name  string
		rows  []row
		of    row
		want  row
		found bool
	}{
		{
			name: "an empty sequence has no duplicate",
			of:   row{id: 1, name: "a"},
		},
		{
			name: "no match",
			rows: []row{{id: 1, name: "a"}, {id: 2, name: "b"}},
			of:   row{id: 3, name: "c"},
		},
		{
			name:  "a match is returned whole, not just its key",
			rows:  []row{{id: 1, name: "a"}, {id: 2, name: "b"}},
			of:    row{id: 3, name: "b"},
			want:  row{id: 2, name: "b"},
			found: true,
		},
		{
			name:  "the first match wins",
			rows:  []row{{id: 1, name: "x"}, {id: 2, name: "x"}},
			of:    row{id: 3, name: "x"},
			want:  row{id: 1, name: "x"},
			found: true,
		},
		{
			name:  "the zero key is a key like any other",
			rows:  []row{{id: 1, name: ""}},
			of:    row{id: 2, name: ""},
			want:  row{id: 1, name: ""},
			found: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, found := FirstDuplicate(slices.Values(test.rows), byName, test.of)
			if found != test.found {
				t.Fatalf("found = %t, want %t", found, test.found)
			}
			if got != test.want {
				t.Errorf("duplicate = %+v, want %+v", got, test.want)
			}
		})
	}
}

func TestFirstDuplicateStopsAtTheMatch(t *testing.T) {
	visited := 0
	counted := func(yield func(row) bool) {
		for _, r := range []row{{id: 1, name: "a"}, {id: 2, name: "b"}, {id: 3, name: "c"}} {
			visited++
			if !yield(r) {
				return
			}
		}
	}

	if _, found := FirstDuplicate(counted, func(r row) string { return r.name }, row{name: "b"}); !found {
		t.Fatal("FirstDuplicate found nothing, want the second row")
	}
	if visited != 2 {
		t.Errorf("walked %d rows, want 2 — the sequence must stop at the match", visited)
	}
}
