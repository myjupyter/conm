package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/ui/view"
)

func TestModel_startsChord(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "the first key of settings", args: args{key: "s"}, want: true},
		{name: "the first key of the connection type picker", args: args{key: "a"}, want: true},
		{name: "a key no sequence starts with", args: args{key: "e"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Model{}.startsChord(tt.args.key))
		})
	}
}

func TestModel_chordKey(t *testing.T) {
	type args struct {
		first  string
		second string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "a whole sequence", args: args{first: "s", second: "s"}, want: "s s"},
		{name: "another whole sequence", args: args{first: "a", second: "a"}, want: "a a"},
		{name: "a second key that makes no sequence", args: args{first: "s", second: "x"}, want: "s"},
		{name: "the window ran out", args: args{first: "a", second: ""}, want: "a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Model{}.chordKey(tt.args.first, tt.args.second))
		})
	}
}

func TestModel_finishChord(t *testing.T) {
	type args struct {
		first  string
		second tea.KeyPressMsg
	}
	tests := []struct {
		name          string
		args          args
		wantSearching bool
		wantChord     string
	}{
		{name: "a second key that makes no sequence is handled too", args: args{first: "e", second: tea.KeyPressMsg{Code: '/', Text: "/"}}, wantSearching: true},
		{name: "a second key may start the next sequence", args: args{first: "e", second: tea.KeyPressMsg{Code: 's', Text: "s"}}, wantChord: "s"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(view.NewConnections(), nil, nil)
			m.chord = tt.args.first

			next, _ := m.finishChord(tt.args.second)
			got, ok := next.(Model)
			require.True(t, ok)

			assert.Equal(t, tt.wantSearching, got.conns.Searching())
			assert.Equal(t, tt.wantChord, got.chord)
		})
	}
}
