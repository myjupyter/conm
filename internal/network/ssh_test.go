package network

import (
	"context"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
)

func sshServer(t *testing.T, banner string) config.SSH {
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
			if banner == "" {
				time.AfterFunc(time.Second, func() { conn.Close() })
				continue
			}
			_, _ = conn.Write([]byte(banner))
			conn.Close()
		}
	}()

	host, port, err := net.SplitHostPort(ln.Addr().String())
	require.NoError(t, err)
	portNumber, err := strconv.Atoi(port)
	require.NoError(t, err)

	return config.SSH{Hostname: host, PortNumber: portNumber, User: "me", Auth: config.SSHAuthAgent}
}

func TestSSHPing(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		banner string
		code   ErrorCode
	}{
		"ssh banner":    {banner: "SSH-2.0-OpenSSH_9.9\r\n"},
		"other service": {banner: "HTTP/1.1 400 Bad Request\r\n", code: UnknownErrorCode},
		"silent peer":   {code: TimeoutErrorCode},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client, err := NewSSHClient(cli.Launcher{}, sshServer(t, tt.banner), nil)
			require.NoError(t, err)

			ctx, cancel := context.WithTimeout(t.Context(), 200*time.Millisecond)
			defer cancel()

			result, err := client.Ping(ctx)
			if tt.code == "" {
				require.NoError(t, err)
				assert.Positive(t, result.PingTime)
				return
			}

			var opErr *OpError
			require.ErrorAs(t, err, &opErr)
			assert.Equal(t, tt.code, opErr.Code)
		})
	}
}

func TestSSHFirstHopAddress(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		jump string
		want string
	}{
		"no jump":                 {want: "target:2222"},
		"jump without port":       {jump: "bastion", want: "bastion:22"},
		"jump with user and port": {jump: "deploy@bastion:2200", want: "bastion:2200"},
		"first of many hops":      {jump: " a@hop1:23, hop2", want: "hop1:23"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cfg := config.SSH{Hostname: "target", PortNumber: 2222, Jump: tt.jump}
			assert.Equal(t, tt.want, firstHopAddress(cfg))
		})
	}
}

func TestAwaitSSHBanner(t *testing.T) {
	t.Parallel()

	const banner = "SSH-2.0-OpenSSH_9.9\r\n"

	tests := map[string]struct {
		sent    string
		wantErr error
	}{
		"banner first":           {sent: banner},
		"banner without crlf":    {sent: "SSH-2.0-OpenSSH_9.9"},
		"preamble before banner": {sent: "welcome\r\nauthorized use only\r\n" + banner},
		"banner at line limit":   {sent: strings.Repeat("x\r\n", sshMaxPreBannerLines) + banner},
		"banner past line limit": {sent: strings.Repeat("x\r\n", sshMaxPreBannerLines+1) + banner, wantErr: errNoSSHBanner},
		"line at length limit":   {sent: strings.Repeat("x", sshMaxLineLength-2) + "\r\n" + banner},
		"line past length limit": {sent: strings.Repeat("x", sshMaxLineLength+1) + "\r\n" + banner, wantErr: errNoSSHBanner},
		"no banner before close": {sent: "HTTP/1.1 400 Bad Request\r\n", wantErr: errNoSSHBanner},
		"nothing before close":   {wantErr: errNoSSHBanner},
		"banner prefix mid-line": {sent: "hello SSH-2.0-OpenSSH_9.9\r\n", wantErr: errNoSSHBanner},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client, server := net.Pipe()
			t.Cleanup(func() { client.Close() })
			go func() {
				_, _ = server.Write([]byte(tt.sent))
				server.Close()
			}()

			err := awaitSSHBanner(client)
			if tt.wantErr == nil {
				assert.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}

	t.Run("read error passes through", func(t *testing.T) {
		t.Parallel()

		client, server := net.Pipe()
		t.Cleanup(func() { client.Close(); server.Close() })
		require.NoError(t, client.SetReadDeadline(time.Now()))

		err := awaitSSHBanner(client)
		require.ErrorIs(t, err, os.ErrDeadlineExceeded)
		assert.NotErrorIs(t, err, errNoSSHBanner)
	})
}
