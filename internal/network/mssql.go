package network

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	mssql "github.com/microsoft/go-mssqldb"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

const mssqlDialTimeout = 5 * time.Second

const (
	mssqlErrLoginFailed      = 18456
	mssqlErrCannotOpenDB     = 4060
	mssqlErrCannotOpenDBUser = 4063
	mssqlErrDatabaseMissing  = 911
)

type MSSQLClient struct {
	launcher cli.Launcher
	cfg      config.Connection
	ref      secret.Reference
	sec      secret.Provider
}

func NewMSSQLClient(
	launcher cli.Launcher,
	cfg config.Connection,
	sec secret.Provider,
) (*MSSQLClient, error) {
	ref, ok := cfg.(secret.Reference)
	if !ok {
		return nil, fmt.Errorf("connection %q does not support secrets", cfg.Name())
	}

	return &MSSQLClient{
		launcher: launcher,
		cfg:      cfg,
		ref:      ref,
		sec:      sec,
	}, nil
}

func (m *MSSQLClient) dsn(ctx context.Context) (string, error) {
	password, err := m.sec.Resolve(ctx, m.ref)
	if err != nil {
		return "", err
	}
	return m.cfg.ConnectionString(password), nil
}

func (m *MSSQLClient) Ping(ctx context.Context) (PingResult, error) {
	raw, err := m.dsn(ctx)
	if err != nil {
		return PingResult{}, m.fail(PingOperation, SecretErrorCode, err)
	}

	dsn, err := mssqlDriverDSN(raw)
	if err != nil {
		return PingResult{}, m.fail(PingOperation, InvalidErrorCode, err)
	}

	db, err := sql.Open("sqlserver", dsn)
	if err != nil {
		return PingResult{}, m.fail(PingOperation, InvalidErrorCode, err)
	}
	defer db.Close()

	now := time.Now()
	if err := db.PingContext(ctx); err != nil {
		return PingResult{}, m.fail(PingOperation, mssqlErrorCode(err), err)
	}

	return PingResult{PingTime: time.Since(now)}, nil
}

func (m *MSSQLClient) Run(ctx context.Context) error {
	password, err := m.sec.Resolve(ctx, m.ref)
	if err != nil {
		return m.fail(ConnectOperation, SecretErrorCode, err)
	}

	if err := m.launcher.Run(ctx, m.cfg, password); err != nil {
		return m.fail(ConnectOperation, cliErrorCode(err, mssqlErrorCode), err)
	}

	return nil
}

func (m *MSSQLClient) Close() error {
	return nil
}

func (m *MSSQLClient) fail(op Operation, code ErrorCode, err error) error {
	if err == nil {
		return nil
	}

	return &OpError{
		Op:     op,
		Code:   code,
		Target: mssqlTarget(m.cfg),
		During: during(op, m.cfg),
		Err:    err,
	}
}

func mssqlDriverDSN(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("dial timeout", strconv.Itoa(int(mssqlDialTimeout.Seconds())))
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func mssqlErrorCode(err error) ErrorCode {
	if serverErr, ok := errors.AsType[mssql.Error](err); ok {
		switch serverErr.Number {
		case mssqlErrLoginFailed:
			return AuthErrorCode
		case mssqlErrCannotOpenDB, mssqlErrCannotOpenDBUser, mssqlErrDatabaseMissing:
			return InvalidErrorCode
		}
	}

	if strings.Contains(strings.ToLower(err.Error()), "tls handshake failed") {
		return TLSErrorCode
	}

	return transportErrorCode(err)
}

func mssqlTarget(cfg config.Connection) string {
	return fmt.Sprintf("sqlserver://%s@%s:%d?database=%s", cfg.Username(), cfg.Host(), cfg.Port(), cfg.Database())
}
