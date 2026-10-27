package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
