package ui

import (
	"slices"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui/spec"
)

func newTagsForm(t *testing.T, tags []string) formModel {
	t.Helper()

	m := newLinkForm(t, nil)
	m.vals["tags"] = spec.JoinStringArray(tags)
	m.idx = slices.Index(m.sections[m.section].Fields, "tags")
	require.GreaterOrEqual(t, m.idx, 0)
	require.True(t, isArray(m.fields[m.currentField()]))

	return m
}

func typeKeys(t *testing.T, m formModel, keys ...tea.KeyPressMsg) formModel {
	t.Helper()

	for _, k := range keys {
		next, _ := m.Update(k)
		got, ok := next.(formModel)
		require.True(t, ok)
		m = got
	}
	return m
}

func TestFormArrayKeys(t *testing.T) {
	char := func(r rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: r, Text: string(r)} }
	var (
		add    = char('a')
		edit   = char('e')
		remove = char('d')
		right  = char('l')
		left   = tea.KeyPressMsg{Code: tea.KeyLeft}
		enter  = tea.KeyPressMsg{Code: tea.KeyEnter}
		esc    = tea.KeyPressMsg{Code: tea.KeyEscape}
		back   = tea.KeyPressMsg{Code: tea.KeyBackspace}
	)

	tests := map[string]struct {
		tags     []string
		keys     []tea.KeyPressMsg
		want     []string
		wantChip int
	}{
		"add appends": {
			tags: []string{"prod"},
			keys: []tea.KeyPressMsg{add, char('e'), char('u'), enter},
			want: []string{"prod", "eu"}, wantChip: 1,
		},
		"add to none": {
			keys: []tea.KeyPressMsg{add, char('x'), enter},
			want: []string{"x"},
		},
		"esc discards": {
			tags: []string{"prod"},
			keys: []tea.KeyPressMsg{add, char('x'), esc},
			want: []string{"prod"},
		},
		"blank add is dropped": {
			tags: []string{"prod"},
			keys: []tea.KeyPressMsg{add, char(' '), enter},
			want: []string{"prod"},
		},
		"edit replaces the focused value": {
			tags: []string{"prod", "eu"},
			keys: []tea.KeyPressMsg{right, edit, back, back, char('u'), char('s'), enter},
			want: []string{"prod", "us"}, wantChip: 1,
		},
		"edit to blank removes": {
			tags: []string{"prod", "eu"},
			keys: []tea.KeyPressMsg{edit, back, back, back, back, enter},
			want: []string{"eu"},
		},
		"d removes the focused value": {
			tags: []string{"prod", "eu", "pci"},
			keys: []tea.KeyPressMsg{right, remove},
			want: []string{"prod", "pci"}, wantChip: 1,
		},
		"d on the last keeps the cursor in range": {
			tags: []string{"prod", "eu"},
			keys: []tea.KeyPressMsg{right, remove},
			want: []string{"prod"},
		},
		"d on none is a no-op": {
			keys: []tea.KeyPressMsg{remove},
		},
		"moves stop at the ends": {
			tags: []string{"prod", "eu"},
			keys: []tea.KeyPressMsg{right, right, right},
			want: []string{"prod", "eu"}, wantChip: 1,
		},
		"left from the first stays": {
			tags: []string{"prod", "eu"},
			keys: []tea.KeyPressMsg{left},
			want: []string{"prod", "eu"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			m := typeKeys(t, newTagsForm(t, tt.tags), tt.keys...)

			assert.Equal(t, tt.want, spec.SplitStringArray(m.vals["tags"]))
			assert.Equal(t, tt.wantChip, m.chipAtCursor(m.fields[m.currentField()]))
			assert.False(t, m.insert)
		})
	}
}

func TestFormArrayAddDoesNotAddLink(t *testing.T) {
	m := typeKeys(t, newTagsForm(t, nil), tea.KeyPressMsg{Code: 'a', Text: "a"})

	assert.Zero(t, m.linkCount())
	assert.True(t, m.insert)
}

func TestFormArrayBuildsTags(t *testing.T) {
	m := newTagsForm(t, []string{"prod", "eu"})

	got, err := m.result()

	require.NoError(t, err)
	pg, ok := got.(config.Postgres)
	require.True(t, ok)
	assert.Equal(t, []string{"prod", "eu"}, pg.Metadata.Tags)
}
