package network

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"

	_ "github.com/jackc/pgx/stdlib"
)

type PGClient struct {
	conmCfg config.Conm
	cfg     config.Connection
	ref     secret.Reference
	sec     secret.Provider
}

func NewPGClient(
	conmConfig config.Conm,
	cfg config.Connection,
	sec secret.Provider,
) (*PGClient, error) {
	ref, ok := cfg.(secret.Reference)
	if !ok {
		return nil, fmt.Errorf("connection %q does not support secrets", cfg.Name())
	}

	return &PGClient{
		conmCfg: conmConfig,
		cfg:     cfg,
		ref:     ref,
		sec:     sec,
	}, nil
}

func (p *PGClient) dsn(ctx context.Context) (string, error) {
	password, err := p.sec.Resolve(ctx, p.ref)
	if err != nil {
		return "", err
	}
	return p.cfg.ConnectionString(password), nil
}

func (p *PGClient) ConnType() config.ConnType {
	return p.cfg.ConnType()
}

func (p *PGClient) Name() string {
	return p.cfg.Name()
}

func (p *PGClient) Description() string {
	return p.cfg.Description()
}

func (p *PGClient) Tags() []string {
	return p.cfg.Tags()
}

func (p *PGClient) Host() string {
	return p.cfg.Host()
}

func (p *PGClient) Username() string {
	return p.cfg.Username()
}

func (p *PGClient) Port() int {
	return p.cfg.Port()
}

func (p *PGClient) Database() string {
	return p.cfg.Database()
}

func (p *PGClient) Schema() string {
	return p.cfg.Schema()
}

func (p *PGClient) ConnectionString(password string) string {
	return p.cfg.ConnectionString(password)
}

func (p *PGClient) IsValid() bool {
	return p.cfg.IsValid()
}

func (p *PGClient) Validate() []error {
	return p.cfg.Validate()
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
	dsn, err := p.dsn(ctx)
	if err != nil {
		return p.fail(ConnectOperation, SecretErrorCode, err)
	}

	executor := exec.CommandContext(ctx, p.conmCfg.Postgres.CLI, dsn)

	executor.Stdin = os.Stdin
	executor.Stdout = os.Stdout
	executor.Stderr = os.Stderr

	if err := executor.Run(); err != nil {
		return p.fail(ConnectOperation, pgErrorCode(err), err)
	}

	return nil
}

func (p *PGClient) Close() error {
	return nil
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
