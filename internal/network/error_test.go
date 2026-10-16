package network

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

func dialError(ip string, errno syscall.Errno) error {
	return &net.OpError{
		Op:   "dial",
		Net:  "tcp",
		Addr: &net.TCPAddr{IP: net.ParseIP(ip), Port: 22},
		Err:  os.NewSyscallError("connect", errno),
	}
}

func Test_transportErrorCodeOn(t *testing.T) {
	t.Parallel()

	type args struct {
		goos string
		err  error
	}
	tests := []struct {
		name string
		args args
		want ErrorCode
	}{
		{
			name: "refused",
			args: args{goos: "linux", err: dialError("10.0.0.1", syscall.ECONNREFUSED)},
			want: RefusedErrorCode,
		},
		{
			name: "unreachable public address",
			args: args{goos: "darwin", err: dialError("203.0.113.7", syscall.EHOSTUNREACH)},
			want: UnreachableErrorCode,
		},
		{
			name: "unreachable private address on darwin",
			args: args{goos: "darwin", err: dialError("192.168.1.101", syscall.EHOSTUNREACH)},
			want: LocalNetErrorCode,
		},
		{
			name: "unreachable link-local address on darwin",
			args: args{goos: "darwin", err: dialError("fe80::1", syscall.EHOSTUNREACH)},
			want: LocalNetErrorCode,
		},
		{
			name: "unreachable private address on linux",
			args: args{goos: "linux", err: dialError("192.168.1.101", syscall.EHOSTUNREACH)},
			want: UnreachableErrorCode,
		},
		{
			name: "wrapped unreachable",
			args: args{goos: "darwin", err: fmt.Errorf("connect: %w", dialError("192.168.1.101", syscall.EHOSTUNREACH))},
			want: LocalNetErrorCode,
		},
		{
			name: "unresolved host",
			args: args{goos: "linux", err: &net.DNSError{Err: "no such host", Name: "nope.local", IsNotFound: true}},
			want: UnresolvedErrorCode,
		},
		{
			name: "deadline exceeded",
			args: args{goos: "linux", err: fmt.Errorf("ping: %w", context.DeadlineExceeded)},
			want: TimeoutErrorCode,
		},
		{
			name: "text-only no such host",
			args: args{goos: "linux", err: errors.New("dial tcp: lookup db: no such host")},
			want: UnresolvedErrorCode,
		},
		{
			name: "text-only no route to host",
			args: args{goos: "darwin", err: errors.New("dial tcp 192.168.1.101:22: connect: no route to host")},
			want: UnreachableErrorCode,
		},
		{
			name: "text-only certificate",
			args: args{goos: "linux", err: errors.New("x509: certificate signed by unknown authority")},
			want: TLSErrorCode,
		},
		{
			name: "unknown",
			args: args{goos: "linux", err: errors.New("boom")},
			want: UnknownErrorCode,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := transportErrorCodeOn(tt.args.goos, tt.args.err)
			assert.Equal(t, tt.want, got)
		})
	}
}
