package transport

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

const (
	sshConnectTimeout = 10 * time.Second
	sshPollInterval   = 50 * time.Millisecond
)

type launcher interface {
	Run(ctx context.Context, cfg config.Connection, tr config.Transport, password string, target config.Connection) error
}

var _ Transport = (*SSH)(nil)

type SSH struct {
	launcher launcher
	bastion  config.SSH
	tr       config.Transport
	target   config.Connection
	sec      secret.Provider

	mx      sync.Mutex
	users   int
	current *tunnel
}

func NewSSH(
	launcher launcher,
	bastion config.SSH,
	tr config.Transport,
	target config.Connection,
	sec secret.Provider,
) *SSH {
	return &SSH{
		launcher: launcher,
		bastion:  bastion,
		tr:       tr,
		target:   target,
		sec:      sec,
	}
}

func (s *SSH) Open(ctx context.Context) (string, int, error) {
	s.mx.Lock()
	defer s.mx.Unlock()

	if s.current == nil || !s.current.isAlive() {
		current, err := s.open(ctx)
		if err != nil {
			return "", 0, s.fail(err)
		}
		s.current = current
	}
	s.users++

	return s.current.host, s.current.port, nil
}

func (s *SSH) Close(ctx context.Context) error {
	s.mx.Lock()
	defer s.mx.Unlock()

	if s.users == 0 {
		return nil
	}
	s.users--
	if s.users > 0 {
		return nil
	}

	current := s.current
	s.current = nil
	return current.stop(ctx)
}

func (s *SSH) open(ctx context.Context) (*tunnel, error) {
	password, err := s.sec.Resolve(ctx, s.bastion)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrSecret, err)
	}

	tr := s.tr
	tr.LocalPort, err = reserveLocalPort(tr.LocalHost(), tr.LocalPort)
	if err != nil {
		return nil, err
	}

	return openTunnel(ctx, tr.LocalHost(), tr.LocalPort, s.connectTimeout(), func(ctx context.Context) error {
		return s.launcher.Run(ctx, s.bastion, tr, password, s.target)
	})
}

func (s *SSH) connectTimeout() time.Duration {
	if s.tr.ConnectTimeout > 0 {
		return time.Duration(s.tr.ConnectTimeout) * time.Second
	}
	return sshConnectTimeout
}

func (s *SSH) fail(err error) error {
	return fmt.Errorf("transport %q to %s via %s: %w", s.tr.Name, address(s.target), address(s.bastion), err)
}

type tunnel struct {
	host   string
	port   int
	cancel context.CancelFunc
	exited chan struct{}
	err    error
}

func openTunnel(
	ctx context.Context,
	host string,
	port int,
	readyTimeout time.Duration,
	run func(context.Context) error,
) (*tunnel, error) {
	runCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	t := &tunnel{
		host:   host,
		port:   port,
		cancel: cancel,
		exited: make(chan struct{}),
	}
	go func() {
		t.err = run(runCtx)
		close(t.exited)
	}()

	if err := t.awaitReady(ctx, readyTimeout); err != nil {
		return nil, errors.Join(err, t.stop(context.WithoutCancel(ctx)))
	}

	return t, nil
}

func (t *tunnel) isAlive() bool {
	select {
	case <-t.exited:
		return false
	default:
		return true
	}
}

func (t *tunnel) awaitReady(ctx context.Context, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(sshPollInterval)
	defer ticker.Stop()

	dialer := net.Dialer{Timeout: sshPollInterval}
	local := net.JoinHostPort(t.host, strconv.Itoa(t.port))
	for {
		select {
		case <-t.exited:
			if t.err != nil {
				return t.err
			}
			return ErrNotReady
		case <-ctx.Done():
			return ErrNotReady
		case <-ticker.C:
			conn, err := dialer.DialContext(ctx, "tcp", local)
			if err == nil {
				return conn.Close()
			}
		}
	}
}

func (t *tunnel) stop(ctx context.Context) error {
	t.cancel()
	select {
	case <-t.exited:
		return t.err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func reserveLocalPort(host string, port int) (int, error) {
	ln, err := net.Listen("tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return 0, fmt.Errorf("%w: %d: %w", ErrLocalPortBusy, port, err)
	}
	defer ln.Close()

	addr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		return 0, fmt.Errorf("unexpected local address %s", ln.Addr())
	}
	return addr.Port, nil
}

func address(c config.Connection) string {
	return net.JoinHostPort(c.Host(), strconv.Itoa(c.Port()))
}
