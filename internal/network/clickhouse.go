package network

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

const clickhouseDialTimeout = 5 * time.Second

// The clickhouse binary is a multi-call tool: the interactive client is a
// subcommand of it rather than a program of its own.
const (
	clickhouseErrAuthFailed     = 516
	clickhouseErrUnknownUser    = 192
	clickhouseErrAccessDenied   = 497
	clickhouseErrUnknownDB      = 81
	clickhouseErrDBDoesNotExist = 82
)

type ClickHouseClient struct {
	launcher cli.Launcher
	cfg      config.Connection
	ref      secret.Reference
	sec      secret.Provider
}

func NewClickHouseClient(
	launcher cli.Launcher,
	cfg config.Connection,
	sec secret.Provider,
) (*ClickHouseClient, error) {
	ref, ok := cfg.(secret.Reference)
	if !ok {
		return nil, fmt.Errorf("connection %q does not support secrets", cfg.Name())
	}

	return &ClickHouseClient{
		launcher: launcher,
		cfg:      cfg,
		ref:      ref,
		sec:      sec,
	}, nil
}

func (c *ClickHouseClient) Ping(ctx context.Context) (PingResult, error) {
	raw, err := c.dsn(ctx)
	if err != nil {
		return PingResult{}, c.fail(PingOperation, SecretErrorCode, err)
	}

	dsn, err := clickhouseDriverDSN(raw)
	if err != nil {
		return PingResult{}, c.fail(PingOperation, InvalidErrorCode, err)
	}

	opts, err := clickhouse.ParseDSN(dsn)
	if err != nil {
		return PingResult{}, c.fail(PingOperation, InvalidErrorCode, err)
	}

	conn, err := clickhouse.Open(opts)
	if err != nil {
		return PingResult{}, c.fail(PingOperation, InvalidErrorCode, err)
	}
	defer conn.Close()

	now := time.Now()
	if err := conn.Ping(ctx); err != nil {
		return PingResult{}, c.fail(PingOperation, clickhouseErrorCode(err), err)
	}

	return PingResult{PingTime: time.Since(now)}, nil
}

func (c *ClickHouseClient) Run(ctx context.Context) error {
	password, err := c.sec.Resolve(ctx, c.ref)
	if err != nil {
		return c.fail(ConnectOperation, SecretErrorCode, err)
	}

	if err := c.launcher.Run(ctx, c.cfg, password); err != nil {
		return c.fail(ConnectOperation, cliErrorCode(err, clickhouseErrorCode), err)
	}

	return nil
}

func (c *ClickHouseClient) Close() error {
	return nil
}

func (c *ClickHouseClient) dsn(ctx context.Context) (string, error) {
	password, err := c.sec.Resolve(ctx, c.ref)
	if err != nil {
		return "", err
	}
	return c.cfg.ConnectionString(password), nil
}

func (c *ClickHouseClient) fail(op Operation, code ErrorCode, err error) error {
	if err == nil {
		return nil
	}

	return &OpError{
		Op:     op,
		Code:   code,
		Target: clickhouseTarget(c.cfg),
		During: during(op, c.cfg),
		Err:    err,
	}
}

func clickhouseDriverDSN(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("dial_timeout", clickhouseDialTimeout.String())
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func clickhouseErrorCode(err error) ErrorCode {
	if serverErr, ok := errors.AsType[*clickhouse.Exception](err); ok {
		switch serverErr.Code {
		case clickhouseErrAuthFailed, clickhouseErrUnknownUser, clickhouseErrAccessDenied:
			return AuthErrorCode
		case clickhouseErrUnknownDB, clickhouseErrDBDoesNotExist:
			return InvalidErrorCode
		}
	}

	return transportErrorCode(err)
}

func clickhouseTarget(cfg config.Connection) string {
	return fmt.Sprintf("clickhouse://%s@%s:%d/%s", cfg.Username(), cfg.Host(), cfg.Port(), cfg.Database())
}
