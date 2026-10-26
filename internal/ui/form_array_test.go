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

func char(r rune) tea.KeyPressMsg { return tea.KeyPressMsg{Code: r, Text: string(r)} }

func TestFormModel_UpdateArray(t *testing.T) {
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

	type args struct {
		tags []string
		keys []tea.KeyPressMsg
	}
	tests := []struct {
		name       string
		args       args
		want       []string
		wantChip   int
		wantInsert bool
	}{
		{
			name:     "a appends a value",
			args:     args{tags: []string{"prod"}, keys: []tea.KeyPressMsg{add, char('e'), char('u'), enter}},
			want:     []string{"prod", "eu"},
			wantChip: 1,
		},
		{
			name: "a adds to an empty list",
			args: args{keys: []tea.KeyPressMsg{add, char('x'), enter}},
			want: []string{"x"},
		},
		{
			name:       "a starts typing a value rather than adding a link",
			args:       args{keys: []tea.KeyPressMsg{add}},
			wantInsert: true,
		},
		{
			name: "esc discards the value being typed",
			args: args{tags: []string{"prod"}, keys: []tea.KeyPressMsg{add, char('x'), esc}},
			want: []string{"prod"},
		},
		{
			name: "a blank value is dropped",
			args: args{tags: []string{"prod"}, keys: []tea.KeyPressMsg{add, char(' '), enter}},
			want: []string{"prod"},
		},
		{
			name:       "a duplicate is refused and stays open",
			args:       args{tags: []string{"prod", "eu"}, keys: []tea.KeyPressMsg{add, char('e'), char('u'), enter}},
			want:       []string{"prod", "eu"},
			wantInsert: true,
		},
		{
			name: "a value edited to itself is kept",
			args: args{tags: []string{"prod", "eu"}, keys: []tea.KeyPressMsg{edit, enter}},
			want: []string{"prod", "eu"},
		},
		{
			name:     "e replaces the focused value",
			args:     args{tags: []string{"prod", "eu"}, keys: []tea.KeyPressMsg{right, edit, back, back, char('u'), char('s'), enter}},
			want:     []string{"prod", "us"},
			wantChip: 1,
		},
		{
			name: "a value edited to blank is removed",
			args: args{tags: []string{"prod", "eu"}, keys: []tea.KeyPressMsg{edit, back, back, back, back, enter}},
			want: []string{"eu"},
		},
		{
			name:     "d removes the focused value",
			args:     args{tags: []string{"prod", "eu", "pci"}, keys: []tea.KeyPressMsg{right, remove}},
			want:     []string{"prod", "pci"},
			wantChip: 1,
		},
		{
			name: "d on the last value keeps the cursor in range",
			args: args{tags: []string{"prod", "eu"}, keys: []tea.KeyPressMsg{right, remove}},
			want: []string{"prod"},
		},
		{
			name: "d on an empty list does nothing",
			args: args{keys: []tea.KeyPressMsg{remove}},
		},
		{
			name:     "moving right stops at the last value",
			args:     args{tags: []string{"prod", "eu"}, keys: []tea.KeyPressMsg{right, right, right}},
			want:     []string{"prod", "eu"},
			wantChip: 1,
		},
		{
			name: "moving left stops at the first value",
			args: args{tags: []string{"prod", "eu"}, keys: []tea.KeyPressMsg{left}},
			want: []string{"prod", "eu"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := typeKeys(t, newTagsForm(t, tt.args.tags), tt.args.keys...)

			assert.Equal(t, tt.want, spec.SplitStringArray(m.vals["tags"]))
			assert.Equal(t, tt.wantChip, m.chipAtCursor(m.fields[m.currentField()]))
			assert.Equal(t, tt.wantInsert, m.insert)
			assert.Zero(t, m.linkCount())
		})
	}
}

func TestFormModel_fieldErrorTags(t *testing.T) {
	type args struct {
		tags []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "distinct tags",
			args: args{tags: []string{"prod", "eu"}},
			want: "",
		},
		{
			name: "a duplicate tag",
			args: args{tags: []string{"prod", "eu", "prod"}},
			want: "tag 'prod' is set more than once",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTagsForm(t, tt.args.tags)

			assert.Equal(t, tt.want, m.fieldError(m.currentField()))
		})
	}
}

func TestFormModel_resultTags(t *testing.T) {
	type args struct {
		tags []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "no tags",
			args: args{tags: nil},
			want: nil,
		},
		{
			name: "tags in order",
			args: args{tags: []string{"prod", "eu"}},
			want: []string{"prod", "eu"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newTagsForm(t, tt.args.tags).result()

			require.NoError(t, err)
			pg, ok := got.(config.Postgres)
			require.True(t, ok)
			assert.Equal(t, tt.want, pg.Metadata.Tags)
		})
	}
}
