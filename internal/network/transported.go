package network

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
	"github.com/myjupyter/conm/internal/transport"
)

var _ Connection = (*Transported)(nil)

type Transported struct {
	cfg        config.DBConnection
	ref        string
	transports transport.Provider
	build      func(config.DBConnection) (Connection, error)
}

func NewTransported(
	launcher cli.Launcher,
	cfg config.DBConnection,
	transports transport.Provider,
	sec secret.Provider,
) (*Transported, error) {
	ref, ok := cfg.(transport.Reference)
	if !ok {
		return nil, fmt.Errorf("connection %q does not support transports", cfg.Meta().Name)
	}

	return &Transported{
		cfg:        cfg,
		ref:        ref.TransportRef(),
		transports: transports,
		build: func(moved config.DBConnection) (Connection, error) {
			return NewConnection(launcher, moved, sec)
		},
	}, nil
}

func (t *Transported) Ping(ctx context.Context) (PingResult, error) {
	var result PingResult
	err := t.through(ctx, PingOperation, func(inner Connection) error {
		var err error
		result, err = inner.Ping(ctx)
		return err
	})

	return result, err
}

func (t *Transported) Run(ctx context.Context) error {
	return t.through(ctx, ConnectOperation, func(inner Connection) error {
		return inner.Run(ctx)
	})
}

func (t *Transported) Close() error {
	return nil
}

func (t *Transported) through(ctx context.Context, op Operation, use func(Connection) error) error {
	tr, err := t.transports.Resolve(t.cfg)
	if err != nil {
		return t.fail(op, err)
	}

	host, port, err := tr.Open(ctx)
	if err != nil {
		return t.fail(op, err)
	}
	defer func() { _ = tr.Close(context.WithoutCancel(ctx)) }()

	inner, err := t.build(t.cfg.SetHost(host).SetPort(port))
	if err != nil {
		return t.fail(op, err)
	}

	return t.configuredAddress(use(inner), net.JoinHostPort(host, strconv.Itoa(port)))
}

func (t *Transported) configuredAddress(err error, dialled string) error {
	opErr, ok := errors.AsType[*OpError](err)
	if !ok {
		return err
	}

	configured := address(t.cfg)
	opErr.Target = strings.ReplaceAll(opErr.Target, dialled, configured)
	opErr.During = strings.ReplaceAll(opErr.During, dialled, configured)
	return opErr
}

func (t *Transported) fail(op Operation, err error) error {
	code := TransportErrorCode
	if errors.Is(err, transport.ErrSecret) {
		code = SecretErrorCode
	}

	return &OpError{
		Op:     op,
		Code:   code,
		Target: t.ref,
		During: during(op, t.cfg),
		Err:    err,
	}
}
