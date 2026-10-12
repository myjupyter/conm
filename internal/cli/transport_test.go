package cli

import (
	"context"
	"net"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
)

func Test_sshTransportCommand(t *testing.T) {
	t.Parallel()

	type args struct {
		cfg      config.Connection
		tr       config.Transport
		password string
		target   config.Connection
	}
	tests := []struct {
		name        string
		args        args
		wantArgs    []string
		wantAskpass bool
		wantErr     error
	}{
		{
			name: "session options of the bastion stay out of the transport",
			args: args{
				cfg: config.SSH{
					Hostname: "bastion.local", PortNumber: 2222, User: "me", Auth: config.SSHAuthKey,
					Jump: "me@jump.local", ForwardAgent: true, KeepAlive: 99, LocalForward: "5432:localhost:5432",
				},
				tr:       config.Transport{Kind: config.SSHTransportKind, LocalPort: 40000, ConnectTimeout: 5, KeepAlive: 30},
				password: "~/.ssh/id_ed25519",
				target:   config.Postgres{Hostname: "db.internal", PortNumber: 5432},
			},
			wantArgs: []string{
				"-p", "2222", "-i", "~/.ssh/id_ed25519", "-J", "me@jump.local",
				"-N", "-o", "ExitOnForwardFailure=yes", "-o", "BatchMode=yes",
				"-o", "ConnectTimeout=5", "-o", "ServerAliveInterval=30",
				"-L", "127.0.0.1:40000:db.internal:5432", "me@bastion.local",
			},
		},
		{
			name: "password auth answers through askpass, not batch mode",
			args: args{
				cfg:      config.SSH{Hostname: "bastion.local", PortNumber: 22, User: "me", Auth: config.SSHAuthPassword},
				tr:       config.Transport{Kind: config.SSHTransportKind, LocalPort: 40000},
				password: "s3cret",
				target:   config.Postgres{Hostname: "db.internal", PortNumber: 5432},
			},
			wantArgs: append(append([]string{"-p", "22"}, sshPasswordFlags...),
				"-N", "-o", "ExitOnForwardFailure=yes",
				"-L", "127.0.0.1:40000:db.internal:5432", "me@bastion.local",
			),
			wantAskpass: true,
		},
		{
			name: "an ipv6 target is bracketed",
			args: args{
				cfg:    config.SSH{Hostname: "bastion.local", User: "me", Auth: config.SSHAuthAgent},
				tr:     config.Transport{Kind: config.SSHTransportKind, LocalPort: 40000},
				target: config.Postgres{Hostname: "fd00::5", PortNumber: 5432},
			},
			wantArgs: []string{
				"-N", "-o", "ExitOnForwardFailure=yes", "-o", "BatchMode=yes",
				"-L", "127.0.0.1:40000:[fd00::5]:5432", "me@bastion.local",
			},
		},
		{
			name: "a local port must be picked before ssh is built",
			args: args{
				cfg:    config.SSH{Hostname: "bastion.local", User: "me", Auth: config.SSHAuthAgent},
				tr:     config.Transport{Kind: config.SSHTransportKind},
				target: config.Postgres{Hostname: "db.internal", PortNumber: 5432},
			},
			wantErr: ErrTransport,
		},
		{
			name: "a database is not a bastion",
			args: args{
				cfg:    config.Postgres{Hostname: "db.local", PortNumber: 5432},
				tr:     config.Transport{Kind: config.SSHTransportKind, LocalPort: 40000},
				target: config.Postgres{Hostname: "db.internal", PortNumber: 5432},
			},
			wantErr: ErrConnType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := sshTransportCommand(tt.args.cfg, tt.args.tr, tt.args.password, tt.args.target)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			if got.cleanup != nil {
				t.Cleanup(got.cleanup)
			}

			is := assert.New(t)
			is.Equal(tt.wantArgs, got.args)
			if !tt.wantAskpass {
				is.Nil(got.env)
				return
			}
			is.Contains(got.env, "SSH_ASKPASS_REQUIRE=force")
		})
	}
}

func TestTransportLauncher_Run(t *testing.T) {
	_, lookErr := exec.LookPath(SSH)
	sshInstalled := lookErr == nil

	type args struct {
		cfg    config.Connection
		tr     config.Transport
		target config.Connection
	}
	sshThrough := func(bastionPort int) config.Connection {
		return config.SSH{Hostname: "127.0.0.1", PortNumber: bastionPort, User: "me", Auth: config.SSHAuthAgent}
	}
	sshTransport := config.Transport{Kind: config.SSHTransportKind, LocalPort: 40000, ConnectTimeout: 2}
	target := config.Postgres{Hostname: "db.internal", PortNumber: 5432}
	refusingPort := closedPort(t)

	tests := []struct {
		name         string
		args         args
		needsSSH     bool
		cancelAfter  time.Duration
		wantErr      error
		wantErrPart  string
		wantStopFast bool
	}{
		{
			name: "an unknown transport kind is refused",
			args: args{
				cfg:    config.SSH{Hostname: "bastion.local", User: "me", Auth: config.SSHAuthAgent},
				tr:     config.Transport{Kind: "vpn", LocalPort: 40000},
				target: target,
			},
			wantErr: ErrUnknownTransport,
		},
		{
			name:        "reports why ssh exited",
			args:        args{cfg: sshThrough(refusingPort), tr: sshTransport, target: target},
			needsSSH:    true,
			wantErr:     ErrTransport,
			wantErrPart: strconv.Itoa(refusingPort),
		},
		{
			name:         "a cancelled context stops it without an error",
			args:         args{cfg: sshThrough(silentServer(t)), tr: sshTransport, target: target},
			needsSSH:     true,
			cancelAfter:  300 * time.Millisecond,
			wantStopFast: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.needsSSH && !sshInstalled {
				t.Skip("ssh is not installed")
			}

			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			defer cancel()
			if tt.cancelAfter > 0 {
				time.AfterFunc(tt.cancelAfter, cancel)
			}

			start := time.Now()
			err := TransportLauncher{}.Run(ctx, tt.args.cfg, tt.args.tr, "", tt.args.target)

			if tt.wantStopFast {
				assert.Less(t, time.Since(start), transportStopGrace)
			}
			if tt.wantErr == nil {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, tt.wantErr)
			assert.Contains(t, err.Error(), tt.wantErrPart)
		})
	}
}

func closedPort(t *testing.T) int {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr, ok := ln.Addr().(*net.TCPAddr)
	require.True(t, ok)
	require.NoError(t, ln.Close())

	return addr.Port
}

func silentServer(t *testing.T) int {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { ln.Close() })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			t.Cleanup(func() { conn.Close() })
		}
	}()

	addr, ok := ln.Addr().(*net.TCPAddr)
	require.True(t, ok)
	return addr.Port
}
