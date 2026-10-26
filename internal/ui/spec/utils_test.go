package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitStringArray(t *testing.T) {
	type args struct {
		raw FormFieldValue
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "empty",
			args: args{raw: ""},
			want: nil,
		},
		{
			name: "one value",
			args: args{raw: "prod"},
			want: []string{"prod"},
		},
		{
			name: "whitespace is trimmed",
			args: args{raw: " prod ,  eu-west "},
			want: []string{"prod", "eu-west"},
		},
		{
			name: "blank values are dropped",
			args: args{raw: "prod,, ,eu-west,"},
			want: []string{"prod", "eu-west"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, SplitStringArray(tt.args.raw))
		})
	}
}

func TestJoinStringArray(t *testing.T) {
	type args struct {
		items []string
	}
	tests := []struct {
		name string
		args args
		want FormFieldValue
	}{
		{
			name: "none",
			args: args{items: nil},
			want: "",
		},
		{
			name: "several values",
			args: args{items: []string{"prod", "eu-west", "analytics"}},
			want: "prod, eu-west, analytics",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := JoinStringArray(tt.args.items)

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.args.items, SplitStringArray(got))
		})
	}
}
