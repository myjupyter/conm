package network

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

const mysqlDialTimeout = 5 * time.Second

const (
	mysqlErrAccessDenied     = 1045
	mysqlErrDBAccessDenied   = 1044
	mysqlErrUnknownDatabase  = 1049
	mysqlErrAuthPluginFailed = 1698
)

type MySQLClient struct {
	launcher cli.Launcher
	cfg      config.Connection
	ref      secret.Reference
	sec      secret.Provider
}

func NewMySQLClient(
	launcher cli.Launcher,
	cfg config.Connection,
	sec secret.Provider,
) (*MySQLClient, error) {
	ref, ok := cfg.(secret.Reference)
	if !ok {
		return nil, fmt.Errorf("connection %q does not support secrets", cfg.Name())
	}

	return &MySQLClient{
		launcher: launcher,
		cfg:      cfg,
		ref:      ref,
		sec:      sec,
	}, nil
}

func (m *MySQLClient) Ping(ctx context.Context) (PingResult, error) {
	raw, err := m.dsn(ctx)
	if err != nil {
		return PingResult{}, m.fail(PingOperation, SecretErrorCode, err)
	}

	dsn, err := mysqlDriverDSN(raw)
	if err != nil {
		return PingResult{}, m.fail(PingOperation, InvalidErrorCode, err)
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return PingResult{}, m.fail(PingOperation, InvalidErrorCode, err)
	}
	defer db.Close()

	now := time.Now()
	if err := db.PingContext(ctx); err != nil {
		return PingResult{}, m.fail(PingOperation, mysqlErrorCode(err), err)
	}

	return PingResult{PingTime: time.Since(now)}, nil
}

func (m *MySQLClient) Run(ctx context.Context) error {
	password, err := m.sec.Resolve(ctx, m.ref)
	if err != nil {
		return m.fail(ConnectOperation, SecretErrorCode, err)
	}

	if err := m.launcher.Run(ctx, m.cfg, password); err != nil {
		return m.fail(ConnectOperation, cliErrorCode(err, mysqlErrorCode), err)
	}

	return nil
}

func (m *MySQLClient) Close() error {
	return nil
}

func (m *MySQLClient) dsn(ctx context.Context) (string, error) {
	password, err := m.sec.Resolve(ctx, m.ref)
	if err != nil {
		return "", err
	}
	return m.cfg.ConnectionString(password), nil
}

func (m *MySQLClient) fail(op Operation, code ErrorCode, err error) error {
	if err == nil {
		return nil
	}

	return &OpError{
		Op:     op,
		Code:   code,
		Target: mysqlTarget(m.cfg),
		During: during(op, m.cfg),
		Err:    err,
	}
}

func mysqlDriverDSN(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}

	password, _ := u.User.Password()

	cfg := mysql.NewConfig()
	cfg.Net = "tcp"
	cfg.Addr = u.Host
	cfg.User = u.User.Username()
	cfg.Passwd = password
	cfg.DBName = strings.TrimPrefix(u.Path, "/")
	cfg.TLSConfig = u.Query().Get("tls")
	cfg.Timeout = mysqlDialTimeout

	return cfg.FormatDSN(), nil
}

func mysqlErrorCode(err error) ErrorCode {
	if serverErr, ok := errors.AsType[*mysql.MySQLError](err); ok {
		switch serverErr.Number {
		case mysqlErrAccessDenied, mysqlErrDBAccessDenied, mysqlErrAuthPluginFailed:
			return AuthErrorCode
		case mysqlErrUnknownDatabase:
			return InvalidErrorCode
		}
	}

	if strings.Contains(strings.ToLower(err.Error()), "tls requested but server does not support") {
		return TLSErrorCode
	}

	return transportErrorCode(err)
}

func mysqlTarget(cfg config.Connection) string {
	return fmt.Sprintf("mysql://%s@%s:%d/%s", cfg.Username(), cfg.Host(), cfg.Port(), cfg.Database())
}
