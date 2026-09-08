package network

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"

	_ "github.com/jackc/pgx/stdlib"
)

type PGClient struct {
	launcher cli.Launcher
	cfg      config.Connection
	ref      secret.Reference
	sec      secret.Provider
}

func NewPGClient(
	launcher cli.Launcher,
	cfg config.Connection,
	sec secret.Provider,
) (*PGClient, error) {
	ref, ok := cfg.(secret.Reference)
	if !ok {
		return nil, fmt.Errorf("connection %q does not support secrets", cfg.Meta().Name)
	}

	return &PGClient{
		launcher: launcher,
		cfg:      cfg,
		ref:      ref,
		sec:      sec,
	}, nil
}

func (p *PGClient) Ping(ctx context.Context) (PingResult, error) {
	dsn, err := p.dsn(ctx)
	if err != nil {
		return PingResult{}, p.fail(PingOperation, SecretErrorCode, err)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return PingResult{}, p.fail(PingOperation, InvalidErrorCode, err)
	}
	defer db.Close()

	now := time.Now()
	if err := db.PingContext(ctx); err != nil {
		return PingResult{}, p.fail(PingOperation, pgErrorCode(err), err)
	}

	return PingResult{PingTime: time.Since(now)}, nil
}

func (p *PGClient) Run(ctx context.Context) error {
	password, err := p.sec.Resolve(ctx, p.ref)
	if err != nil {
		return p.fail(ConnectOperation, SecretErrorCode, err)
	}

	if err := p.launcher.Run(ctx, p.cfg, password); err != nil {
		return p.fail(ConnectOperation, cliErrorCode(err, pgErrorCode), err)
	}

	return nil
}

func (p *PGClient) Close() error {
	return nil
}

func (p *PGClient) dsn(ctx context.Context) (string, error) {
	password, err := p.sec.Resolve(ctx, p.ref)
	if err != nil {
		return "", err
	}
	return p.cfg.ConnectionString(password), nil
}

func (p *PGClient) fail(op Operation, code ErrorCode, err error) error {
	if err == nil {
		return nil
	}

	return &OpError{
		Op:     op,
		Code:   code,
		Target: pgTarget(p.cfg),
		During: during(op, p.cfg),
		Err:    err,
	}
}

func pgErrorCode(err error) ErrorCode {
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "password authentication"), strings.Contains(msg, "authentication failed"):
		return AuthErrorCode
	case strings.Contains(msg, "sslmode"), strings.Contains(msg, "ssl is not enabled"):
		return TLSErrorCode
	}

	return transportErrorCode(err)
}

func pgTarget(cfg config.Connection) string {
	return fmt.Sprintf("postgres://%s@%s:%d/%s", cfg.Username(), cfg.Host(), cfg.Port(), cfg.Database())
}
