package cli

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/myjupyter/conm/internal/config"
)

const transportStopGrace = 3 * time.Second

var transports = map[config.TransportKind]transportClient{
	config.SSHTransportKind: {name: SSH, build: sshTransportCommand},
}

type TransportLauncher struct{}

type transportClient struct {
	name  string
	build transportBuilder
}

type transportBuilder func(
	cfg config.Connection,
	tr config.Transport,
	password string,
	target config.Connection,
) (command, error)

func (TransportLauncher) Run(
	ctx context.Context,
	cfg config.Connection,
	tr config.Transport,
	password string,
	target config.Connection,
) error {
	client, ok := transports[tr.Kind]
	if !ok {
		return fmt.Errorf("%w: %q", ErrUnknownTransport, tr.Kind)
	}

	cmd, err := client.build(cfg, tr, password, target)
	if err != nil {
		return err
	}

	return cmd.runInBackground(ctx, client.name)
}

func (c command) runInBackground(ctx context.Context, name string) error {
	if c.cleanup != nil {
		defer c.cleanup()
	}

	var stderr bytes.Buffer
	executor := exec.CommandContext(ctx, name, c.args...)
	executor.Env = c.env
	executor.Stderr = &stderr
	executor.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	executor.Cancel = func() error { return executor.Process.Signal(syscall.SIGTERM) }
	executor.WaitDelay = transportStopGrace

	err := executor.Run()
	if ctx.Err() != nil {
		return nil
	}

	reason := strings.TrimSpace(stderr.String())
	switch {
	case reason != "":
		return fmt.Errorf("%w: %s", ErrTransport, reason)
	case err != nil:
		return fmt.Errorf("%w: %w", ErrTransport, err)
	default:
		return fmt.Errorf("%w: %s exited", ErrTransport, name)
	}
}
