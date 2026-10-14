package ui

import (
	"strings"
	"testing"

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
