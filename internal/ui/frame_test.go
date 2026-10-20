package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
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
			name: "wide characters are split by cells",
			args: args{s: "接続が拒否されました", w: 6},
			want: []string{"接続が", "拒否さ", "れまし", "た"},
		},
		{
			name: "a wide character never splits in half",
			args: args{s: "接続", w: 3},
			want: []string{"接", "続"},
		},
		{
			name: "a line narrower than a wide character still advances",
			args: args{s: "接続", w: 1},
			want: []string{"接", "続"},
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
		{name: "pads on the right", args: args{s: "pg", w: 4}, want: "pg  "},
		{name: "pads on the left", args: args{s: "pg", w: 4, right: true}, want: "  pg"},
		{name: "elides a long text", args: args{s: "postgres", w: 5}, want: "post…"},
		{name: "pads wide characters by cells", args: args{s: "接続", w: 6}, want: "接続  "},
		{name: "elides wide characters by cells", args: args{s: "接続一覧", w: 6}, want: "接続… "},
		{name: "fills the cell a cut wide character leaves", args: args{s: "接続一覧", w: 5}, want: "接続…"},
		{name: "a single cell keeps the first narrow character", args: args{s: "pg", w: 1}, want: "p"},
		{name: "a single cell cannot hold a wide character", args: args{s: "接続", w: 1}, want: " "},
		{name: "no room", args: args{s: "pg", w: 0}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := truncPad(tt.args.s, tt.args.w, tt.args.right)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.args.w, ansi.StringWidth(got))
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
