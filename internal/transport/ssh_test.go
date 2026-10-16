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
		ln, err := net.Listen("tcp", net.JoinHostPort(tr.LocalHost(), strconv.Itoa(tr.LocalPort)))
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

func prodTransport(localPort int) config.Transport {
	return config.Transport{
		Name: "prod", Kind: config.SSHTransportKind, Via: "bastion", LocalPort: localPort, ConnectTimeout: 1,
	}
}

func newTestSSH(launcher *fakeTransportLauncher, tr config.Transport) *SSH {
	if launcher.died == nil {
		launcher.died = make(chan struct{})
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
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(config.Transport{}.LocalHost(), strconv.Itoa(port)), time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func freeLocalPort(t *testing.T) int {
	t.Helper()

	ln, err := net.Listen("tcp", net.JoinHostPort(config.Transport{}.LocalHost(), "0"))
	require.NoError(t, err)
	addr, ok := ln.Addr().(*net.TCPAddr)
	require.True(t, ok)
	require.NoError(t, ln.Close())

	return addr.Port
}

func busyLocalPort(t *testing.T) int {
	t.Helper()

	ln, err := net.Listen("tcp", net.JoinHostPort(config.Transport{}.LocalHost(), "0"))
	require.NoError(t, err)
	t.Cleanup(func() { ln.Close() })
	addr, ok := ln.Addr().(*net.TCPAddr)
	require.True(t, ok)

	return addr.Port
}

func TestSSH_Open(t *testing.T) {
	type fields struct {
		launcher *fakeTransportLauncher
		tr       config.Transport
	}
	tests := []struct {
		name     string
		fields   fields
		prepare  func(t *testing.T, s *SSH, launcher *fakeTransportLauncher)
		wantRuns int
		wantErr  error
	}{
		{
			name:     "starts a tunnel on a free local port",
			fields:   fields{launcher: &fakeTransportLauncher{listen: true}, tr: prodTransport(0)},
			wantRuns: 1,
		},
		{
			name:     "keeps a fixed local port",
			fields:   fields{launcher: &fakeTransportLauncher{listen: true}, tr: prodTransport(freeLocalPort(t))},
			wantRuns: 1,
		},
		{
			name:   "reuses the running tunnel",
			fields: fields{launcher: &fakeTransportLauncher{listen: true}, tr: prodTransport(0)},
			prepare: func(t *testing.T, s *SSH, _ *fakeTransportLauncher) {
				t.Helper()

				_, _, err := s.Open(t.Context())
				require.NoError(t, err)
			},
			wantRuns: 1,
		},
		{
			name:   "replaces a tunnel that died",
			fields: fields{launcher: &fakeTransportLauncher{listen: true}, tr: prodTransport(0)},
			prepare: func(t *testing.T, s *SSH, launcher *fakeTransportLauncher) {
				t.Helper()

				_, port, err := s.Open(t.Context())
				require.NoError(t, err)
				launcher.dropRunning()
				require.Eventually(t, func() bool { return !localPortOpen(port) }, time.Second, 10*time.Millisecond)
			},
			wantRuns: 2,
		},
		{
			name:    "refuses a busy fixed port",
			fields:  fields{launcher: &fakeTransportLauncher{listen: true}, tr: prodTransport(busyLocalPort(t))},
			wantErr: ErrLocalPortBusy,
		},
		{
			name:     "fails when the launcher exits before the port is up",
			fields:   fields{launcher: &fakeTransportLauncher{exitWith: errPermissionDenied}, tr: prodTransport(0)},
			wantRuns: 1,
			wantErr:  errPermissionDenied,
		},
		{
			name:     "fails when the port never comes up",
			fields:   fields{launcher: &fakeTransportLauncher{}, tr: prodTransport(0)},
			wantRuns: 1,
			wantErr:  ErrNotReady,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestSSH(tt.fields.launcher, tt.fields.tr)
			t.Cleanup(func() {
				for range tt.wantRuns + 1 {
					_ = s.Close(context.Background())
				}
			})
			if tt.prepare != nil {
				tt.prepare(t, s, tt.fields.launcher)
			}

			got, got1, err := s.Open(t.Context())

			is := assert.New(t)
			is.Equal(tt.wantRuns, tt.fields.launcher.runCount())
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				is.Contains(err.Error(), `transport "prod" to db.internal:5432 via bastion.local:22`)
				return
			}
			require.NoError(t, err)
			is.Equal(config.Transport{}.LocalHost(), got)
			is.True(localPortOpen(got1))
			if tt.fields.tr.LocalPort != 0 {
				is.Equal(tt.fields.tr.LocalPort, got1)
			}
		})
	}
}

func TestSSH_Close(t *testing.T) {
	type args struct {
		opens  int
		closes int
	}
	tests := []struct {
		name     string
		args     args
		wantOpen bool
	}{
		{
			name:     "keeps the tunnel while a user still holds it",
			args:     args{opens: 2, closes: 1},
			wantOpen: true,
		},
		{
			name: "stops the tunnel with the last user",
			args: args{opens: 2, closes: 2},
		},
		{
			name: "closing more than opened is a no-op",
			args: args{opens: 1, closes: 3},
		},
		{
			name: "closing without opening is a no-op",
			args: args{opens: 0, closes: 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestSSH(&fakeTransportLauncher{listen: true}, prodTransport(0))
			t.Cleanup(func() {
				for range tt.args.opens {
					_ = s.Close(context.Background())
				}
			})

			var port int
			for range tt.args.opens {
				_, got, err := s.Open(t.Context())
				require.NoError(t, err)
				port = got
			}

			for range tt.args.closes {
				require.NoError(t, s.Close(t.Context()))
			}

			if port != 0 {
				assert.Equal(t, tt.wantOpen, localPortOpen(port))
			}
		})
	}
}
