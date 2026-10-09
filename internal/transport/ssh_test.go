package transport

import (
	"context"
	"errors"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

var errPermissionDenied = errors.New("permission denied (publickey)")

type fakeTransportLauncher struct {
	mx       sync.Mutex
	runs     []config.Transport
	listen   bool
	exitWith error
	died     chan struct{}
}

func (f *fakeTransportLauncher) Run(
	ctx context.Context,
	_ config.Connection,
	tr config.Transport,
	_ string,
	_ config.Connection,
) error {
	f.mx.Lock()
	f.runs = append(f.runs, tr)
	died := f.died
	f.mx.Unlock()

	if f.exitWith != nil {
		return f.exitWith
	}
	if f.listen {
		ln, err := net.Listen("tcp", net.JoinHostPort(sshLocalHost, strconv.Itoa(tr.LocalPort)))
		if err != nil {
			return err
		}
		defer ln.Close()
	}

	select {
	case <-ctx.Done():
		return nil
	case <-died:
		return errors.New("connection to bastion lost")
	}
}

func (f *fakeTransportLauncher) dropRunning() {
	f.mx.Lock()
	defer f.mx.Unlock()

	close(f.died)
	f.died = make(chan struct{})
}

func (f *fakeTransportLauncher) runCount() int {
	f.mx.Lock()
	defer f.mx.Unlock()
	return len(f.runs)
}

func newTestSSH(launcher *fakeTransportLauncher, tr config.Transport) *SSH {
	if launcher.died == nil {
		launcher.died = make(chan struct{})
	}
	if tr.Name == "" {
		tr = config.Transport{Name: "prod", Kind: config.SSHTransportKind, Via: "bastion", ConnectTimeout: 1}
	}

	return NewSSH(
		launcher,
		config.SSH{Hostname: "bastion.local", PortNumber: 22, User: "me", Auth: config.SSHAuthAgent},
		tr,
		config.Postgres{Hostname: "db.internal", PortNumber: 5432},
		secret.Default(),
	)
}

func localPortOpen(port int) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(sshLocalHost, strconv.Itoa(port)), time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func requireNamesTheRoute(t *testing.T, err error) {
	t.Helper()

	require.Error(t, err)
	assert.Contains(t, err.Error(), `transport "prod" to db.internal:5432 via bastion.local:22`)
}

func TestSSHSharesOneRunBetweenOverlappingUsers(t *testing.T) {
	launcher := &fakeTransportLauncher{listen: true}
	s := newTestSSH(launcher, config.Transport{})
	ctx := t.Context()

	host, port, err := s.Open(ctx)
	require.NoError(t, err)
	_, again, err := s.Open(ctx)
	require.NoError(t, err)

	is := assert.New(t)
	is.Equal(sshLocalHost, host)
	is.Equal(port, again)
	is.Equal(1, launcher.runCount())
	is.Equal(port, launcher.runs[0].LocalPort)
	is.True(localPortOpen(port))

	require.NoError(t, s.Close(ctx))
	is.True(localPortOpen(port), "a user still holds the transport")
	require.NoError(t, s.Close(ctx))
	is.False(localPortOpen(port))
	require.NoError(t, s.Close(ctx), "closing more than opened is a no-op")
}

func TestSSHKeepsAFixedLocalPort(t *testing.T) {
	ln, err := net.Listen("tcp", net.JoinHostPort(sshLocalHost, "0"))
	require.NoError(t, err)
	addr, ok := ln.Addr().(*net.TCPAddr)
	require.True(t, ok)
	require.NoError(t, ln.Close())

	launcher := &fakeTransportLauncher{listen: true}
	s := newTestSSH(launcher, config.Transport{
		Name: "prod", Kind: config.SSHTransportKind, Via: "bastion", LocalPort: addr.Port, ConnectTimeout: 1,
	})

	_, port, err := s.Open(t.Context())
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close(context.Background()) })

	assert.Equal(t, addr.Port, port)
}

func TestSSHRefusesABusyFixedPort(t *testing.T) {
	ln, err := net.Listen("tcp", net.JoinHostPort(sshLocalHost, "0"))
	require.NoError(t, err)
	t.Cleanup(func() { ln.Close() })
	addr, ok := ln.Addr().(*net.TCPAddr)
	require.True(t, ok)

	launcher := &fakeTransportLauncher{listen: true}
	s := newTestSSH(launcher, config.Transport{
		Name: "prod", Kind: config.SSHTransportKind, Via: "bastion", LocalPort: addr.Port, ConnectTimeout: 1,
	})

	_, _, err = s.Open(t.Context())

	requireNamesTheRoute(t, err)
	require.ErrorIs(t, err, ErrLocalPortBusy)
	assert.Zero(t, launcher.runCount(), "nothing is launched onto a port somebody else holds")
}

func TestSSHOpenFailures(t *testing.T) {
	tests := map[string]struct {
		launcher *fakeTransportLauncher
		wantErr  error
	}{
		"the launcher exits before the port is up": {
			launcher: &fakeTransportLauncher{exitWith: errPermissionDenied},
			wantErr:  errPermissionDenied,
		},
		"the port never comes up": {
			launcher: &fakeTransportLauncher{},
			wantErr:  ErrNotReady,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := newTestSSH(tt.launcher, config.Transport{})

			_, _, err := s.Open(t.Context())

			requireNamesTheRoute(t, err)
			require.ErrorIs(t, err, tt.wantErr)
			require.NoError(t, s.Close(t.Context()), "a failed open leaves nothing to close")
		})
	}
}

func TestSSHReplacesARunThatDied(t *testing.T) {
	launcher := &fakeTransportLauncher{listen: true}
	s := newTestSSH(launcher, config.Transport{})
	ctx := t.Context()

	_, first, err := s.Open(ctx)
	require.NoError(t, err)

	launcher.dropRunning()
	require.Eventually(t, func() bool { return !localPortOpen(first) }, time.Second, 10*time.Millisecond)

	_, second, err := s.Open(ctx)
	require.NoError(t, err)

	is := assert.New(t)
	is.Equal(2, launcher.runCount())
	is.True(localPortOpen(second))

	require.NoError(t, s.Close(ctx))
	is.True(localPortOpen(second), "the first user has not closed yet")
	require.NoError(t, s.Close(ctx))
	is.False(localPortOpen(second))
}
