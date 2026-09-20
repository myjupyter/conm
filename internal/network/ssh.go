package network

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

const sshBannerPrefix = "SSH-"

var errNoSSHBanner = errors.New("no ssh banner received")

type SSHClient struct {
	launcher cli.Launcher
	cfg      config.Connection
	ref      secret.Reference
	sec      secret.Provider
}

func NewSSHClient(
	launcher cli.Launcher,
	cfg config.Connection,
	sec secret.Provider,
) (*SSHClient, error) {
	ref, ok := cfg.(secret.Reference)
	if !ok {
		return nil, fmt.Errorf("connection %q does not support secrets", cfg.Meta().Name)
	}

	return &SSHClient{
		launcher: launcher,
		cfg:      cfg,
		ref:      ref,
		sec:      sec,
	}, nil
}

func (s *SSHClient) Ping(ctx context.Context) (PingResult, error) {
	now := time.Now()

	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "tcp", address(s.cfg))
	if err != nil {
		return PingResult{}, s.fail(PingOperation, transportErrorCode(err), err)
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetReadDeadline(deadline)
	}

	banner, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return PingResult{}, s.fail(PingOperation, transportErrorCode(err), err)
	}
	if !strings.HasPrefix(banner, sshBannerPrefix) {
		return PingResult{}, s.fail(PingOperation, UnknownErrorCode, errNoSSHBanner)
	}

	return PingResult{PingTime: time.Since(now)}, nil
}

func (s *SSHClient) Run(ctx context.Context) error {
	password, err := s.sec.Resolve(ctx, s.ref)
	if err != nil {
		return s.fail(ConnectOperation, SecretErrorCode, err)
	}

	if err := s.launcher.Run(ctx, s.cfg, password); err != nil {
		return s.fail(ConnectOperation, cliErrorCode(err, transportErrorCode), err)
	}

	return nil
}

func (s *SSHClient) Close() error {
	return nil
}

func (s *SSHClient) fail(op Operation, code ErrorCode, err error) error {
	if err == nil {
		return nil
	}

	return &OpError{
		Op:     op,
		Code:   code,
		Target: s.cfg.ConnectionString(""),
		During: during(op, s.cfg),
		Err:    err,
	}
}
