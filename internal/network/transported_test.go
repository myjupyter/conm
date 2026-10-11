package network

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
	"github.com/myjupyter/conm/internal/transport"
)

type fakeTransport struct {
	host    string
	port    int
	openErr error
	opened  int
	closed  int
}

func (f *fakeTransport) Open(context.Context) (string, int, error) {
	if f.openErr != nil {
		return "", 0, f.openErr
	}
	f.opened++
	return f.host, f.port, nil
}

func (f *fakeTransport) Close(context.Context) error {
	f.closed++
	return nil
}

func closedLocalPort(t *testing.T) int {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr, ok := ln.Addr().(*net.TCPAddr)
	require.True(t, ok)
	require.NoError(t, ln.Close())

	return addr.Port
}

func behindTransport() config.Postgres {
	return config.Postgres{
		Hostname: "db.internal", PortNumber: 5432, User: "svc", DBName: "orders",
		SSLMode: config.PostgresSSLModeDisable, Transport: "ssh:prod",
	}
}

type unroutable struct {
	config.Postgres
}

func (unroutable) TransportRef() {}

func TestNewTransported(t *testing.T) {
	t.Parallel()

	type args struct {
		cfg config.DBConnection
	}
	tests := []struct {
		name    string
		args    args
		wantErr string
	}{
		{
			name: "a connection with a transport reference",
			args: args{cfg: behindTransport()},
		},
		{
			name:    "a connection without a transport reference",
			args:    args{cfg: unroutable{behindTransport()}},
			wantErr: "does not support transports",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NewTransported(cli.Launcher{}, tt.args.cfg, &fakeTransport{}, secret.Default())
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, got)
		})
	}
}

func TestTransported_Ping(t *testing.T) {
	t.Parallel()

	type fields struct {
		transport *fakeTransport
	}
	tests := []struct {
		name       string
		fields     fields
		wantCode   ErrorCode
		wantTarget string
		wantOpened int
	}{
		{
			name:       "the database behind the transport refuses",
			fields:     fields{transport: &fakeTransport{host: "127.0.0.1", port: closedLocalPort(t)}},
			wantCode:   RefusedErrorCode,
			wantTarget: "postgres://svc@db.internal:5432/orders",
			wantOpened: 1,
		},
		{
			name:       "the transport does not come up",
			fields:     fields{transport: &fakeTransport{openErr: errors.Join(errors.New("prod"), transport.ErrNotReady)}},
			wantCode:   TransportErrorCode,
			wantTarget: "ssh:prod",
		},
		{
			name:       "the bastion secret is unreadable",
			fields:     fields{transport: &fakeTransport{openErr: errors.Join(errors.New("prod"), transport.ErrSecret)}},
			wantCode:   SecretErrorCode,
			wantTarget: "ssh:prod",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			conn, err := NewTransported(cli.Launcher{}, behindTransport(), tt.fields.transport, secret.Default())
			require.NoError(t, err)

			_, err = conn.Ping(t.Context())

			var opErr *OpError
			require.ErrorAs(t, err, &opErr)
			is := assert.New(t)
			is.Equal(tt.wantCode, opErr.Code)
			is.Equal(tt.wantTarget, opErr.Target)
			is.Contains(opErr.During, "db.internal:5432")
			is.NotContains(opErr.During, "127.0.0.1")
			is.Equal(tt.wantOpened, tt.fields.transport.opened)
			is.Equal(tt.wantOpened, tt.fields.transport.closed)
		})
	}
}

func TestTransported_Run(t *testing.T) {
	t.Parallel()

	type fields struct {
		transport *fakeTransport
	}
	tests := []struct {
		name     string
		fields   fields
		wantCode ErrorCode
	}{
		{
			name:     "the transport does not come up",
			fields:   fields{transport: &fakeTransport{openErr: transport.ErrNotReady}},
			wantCode: TransportErrorCode,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			conn, err := NewTransported(cli.Launcher{}, behindTransport(), tt.fields.transport, secret.Default())
			require.NoError(t, err)

			err = conn.Run(t.Context())

			var opErr *OpError
			require.ErrorAs(t, err, &opErr)
			assert.Equal(t, tt.wantCode, opErr.Code)
			assert.Zero(t, tt.fields.transport.closed)
		})
	}
}
