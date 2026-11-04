package ui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_wrapText(t *testing.T) {
	t.Parallel()

	type args struct {
		s string
		w int
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "a short error stays on one line",
			args: args{s: "connection refused", w: 30},
			want: []string{"connection refused"},
		},
		{
			name: "a long error wraps at word boundaries",
			args: args{s: "dial tcp 127.0.0.1:40123: connect: connection refused", w: 20},
			want: []string{"dial tcp", "127.0.0.1:40123:", "connect: connection", "refused"},
		},
		{
			name: "line breaks in the error are kept",
			args: args{s: "Host key verification failed.\nPermission denied", w: 40},
			want: []string{"Host key verification failed.", "Permission denied"},
		},
		{
			name: "a word longer than the line is split",
			args: args{s: "abcdefghij", w: 4},
			want: []string{"abcd", "efgh", "ij"},
		},
		{
			name: "an empty error still takes a line",
			args: args{s: "", w: 10},
			want: []string{""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, wrapText(tt.args.s, tt.args.w))
		})
	}
}

func Test_frameErrPanel(t *testing.T) {
	t.Parallel()

	type args struct {
		detail string
	}
	long := "ssh exited: @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@\n" +
		"@    WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED!     @\n" +
		"Host key for [localhost]:2222 has changed and you have requested strict checking.\n" +
		"Host key verification failed."

	tests := []struct {
		name      string
		args      args
		wantParts []string
	}{
		{
			name:      "every part of a long error is on screen",
			args:      args{detail: long},
			wantParts: []string{"REMOTE HOST IDENTIFICATION HAS CHANGED", "requested strict", "checking.", "Host key verification failed."},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := strings.Join(frameErrPanel(tableInner, &connError{op: "ping", code: "ETRANSPORT", detail: tt.args.detail}), "\n")

			require.NotContains(t, got, gEllipsis)
			for _, part := range tt.wantParts {
				assert.Contains(t, got, part)
			}
		})
	}
}

func Test_truncPad(t *testing.T) {
	t.Parallel()

	type args struct {
		s     string
		w     int
		right bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "pads on the right", args: args{s: "ab", w: 4}, want: "ab  "},
		{name: "pads on the left", args: args{s: "ab", w: 4, right: true}, want: "  ab"},
		{name: "elides what is too long", args: args{s: "abcdef", w: 4}, want: "abc" + gEllipsis},
		{name: "one cell has no room for an ellipsis", args: args{s: "abc", w: 1}, want: "a"},
		{name: "wide characters count two cells", args: args{s: "数据", w: 6}, want: "数据  "},
		{name: "a wide character is never split", args: args{s: "数据库", w: 4}, want: "数" + gEllipsis + " "},
		{name: "no room is empty", args: args{s: "abc", w: 0}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, truncPad(tt.args.s, tt.args.w, tt.args.right))
		})
	}
}

func Test_fitNote(t *testing.T) {
	t.Parallel()

	type args struct {
		note string
		w    int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "a note that fits is untouched", args: args{note: " 3 rows · 1 below ", w: 30}, want: " 3 rows · 1 below "},
		{name: "the leading part goes first", args: args{note: " 3 rows · 1 below ", w: 10}, want: " 1 below "},
		{name: "a single part is cut from the left", args: args{note: " 1 below ", w: 4}, want: "low "},
		{name: "no room leaves nothing", args: args{note: " 1 below ", w: 0}, want: ""},
		{name: "negative room leaves nothing", args: args{note: " 1 below ", w: -3}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, fitNote(tt.args.note, tt.args.w))
		})
	}
}

func Test_frameLine_clipsToTheFrame(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		spans []span
	}{
		{name: "one span too long", spans: []span{{text: strings.Repeat("x", 40)}}},
		{name: "a later span too long", spans: []span{{text: "abc"}, {text: strings.Repeat("x", 40)}}},
		{name: "spans past the frame are dropped", spans: []span{{text: strings.Repeat("x", 10)}, {text: "ab"}, {text: "cd"}}},
		{name: "a wide character at the edge", spans: []span{{text: strings.Repeat("数", 40)}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, 12, lipgloss.Width(frameLine(10, tt.spans, nil)))
		})
	}
}
