package ui

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/myjupyter/conm/internal/network"
)

func Test_newConnError_through(t *testing.T) {
	t.Parallel()

	type args struct {
		err error
	}
	opErr := func(code network.ErrorCode, through string) error {
		return &network.OpError{Op: network.PingOperation, Code: code, Through: through, Err: errors.New("connection reset by peer")}
	}

	tests := []struct {
		name     string
		args     args
		wantHint string
	}{
		{
			name:     "a database behind a tunnel that does not answer",
			args:     args{err: opErr(network.UnknownErrorCode, "ssh:test-bastion")},
			wantHint: "the tunnel ssh:test-bastion is up but could not reach the database",
		},
		{
			name:     "a database behind a tunnel that refuses",
			args:     args{err: opErr(network.RefusedErrorCode, "ssh:test-bastion")},
			wantHint: "host and port must be the address as the bastion sees it",
		},
		{
			name:     "a database behind a tunnel that rejects the credentials",
			args:     args{err: opErr(network.AuthErrorCode, "ssh:test-bastion")},
			wantHint: "credentials were rejected",
		},
		{
			name:     "a tunnel whose local port is taken",
			args:     args{err: opErr(network.PortBusyErrorCode, "")},
			wantHint: "give this one its own transport",
		},
		{
			name:     "a direct database that refuses",
			args:     args{err: opErr(network.RefusedErrorCode, "")},
			wantHint: "is the server running",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := newConnError(tt.args.err, network.PingOperation)

			assert.Contains(t, got.hint, tt.wantHint)
			panel := strings.Join(frameErrPanel(tableInner, got), "\n")
			assert.NotContains(t, panel, gEllipsis)
		})
	}
}
