package network

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	mssql "github.com/microsoft/go-mssqldb"

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
	conmCfg config.Conm
	cfg     config.Connection
	ref     secret.Reference
	sec     secret.Provider
}

func NewMSSQLClient(
	conmConfig config.Conm,
	cfg config.Connection,
	sec secret.Provider,
) (*MSSQLClient, error) {
	ref, ok := cfg.(secret.Reference)
	if !ok {
		return nil, fmt.Errorf("connection %q does not support secrets", cfg.Name())
	}

	return &MSSQLClient{
		conmCfg: conmConfig,
		cfg:     cfg,
		ref:     ref,
		sec:     sec,
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

	target, err := parseMSSQLDSN(raw)
	if err != nil {
		return PingResult{}, m.fail(PingOperation, InvalidErrorCode, err)
	}

	db, err := sql.Open("sqlserver", target.driverDSN())
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
	raw, err := m.dsn(ctx)
	if err != nil {
		return m.fail(ConnectOperation, SecretErrorCode, err)
	}

	target, err := parseMSSQLDSN(raw)
	if err != nil {
		return m.fail(ConnectOperation, InvalidErrorCode, err)
	}

	cli := m.conmCfg.CLI(config.MSSQLConnType)
	if cli == "" {
		return m.fail(ConnectOperation, InvalidErrorCode, errors.New("no mssql client configured, run conm init"))
	}

	if err := runCLI(ctx, cli, target.args(cli, raw), target.env(cli)); err != nil {
		return m.fail(ConnectOperation, mssqlErrorCode(err), err)
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

// mssqlDSN is the sqlserver:// URL the driver reads natively, kept apart so
// sqlcmd can be handed the same connection as flags with the password out of
// band.
type mssqlDSN struct {
	raw       *url.URL
	addr      string
	user      string
	password  string
	database  string
	encrypt   string
	trustCert bool
}

func parseMSSQLDSN(raw string) (mssqlDSN, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return mssqlDSN{}, err
	}

	password, _ := u.User.Password()
	q := u.Query()

	return mssqlDSN{
		raw:       u,
		addr:      u.Host,
		user:      u.User.Username(),
		password:  password,
		database:  q.Get("database"),
		encrypt:   q.Get("encrypt"),
		trustCert: q.Get("trustservercertificate") == "true",
	}, nil
}

func (d mssqlDSN) driverDSN() string {
	u := *d.raw
	q := u.Query()
	q.Set("dial timeout", strconv.Itoa(int(mssqlDialTimeout.Seconds())))
	u.RawQuery = q.Encode()

	return u.String()
}

func (d mssqlDSN) args(cli, raw string) []string {
	if cli != config.SQLCmdCLI {
		return []string{raw}
	}

	args := []string{"-S", d.server(), "-U", d.user, "-d", d.database}
	if d.encrypt == config.MSSQLEncryptRequire || d.encrypt == config.MSSQLEncryptStrict {
		args = append(args, "-N")
	}
	if d.trustCert {
		args = append(args, "-C")
	}

	return args
}

// env keeps the password out of the process table: sqlcmd reads it from
// SQLCMDPASSWORD, while a client taking the URL already carries it inside.
func (d mssqlDSN) env(cli string) []string {
	if cli != config.SQLCmdCLI || d.password == "" {
		return nil
	}

	return append(os.Environ(), "SQLCMDPASSWORD="+d.password)
}

func (d mssqlDSN) server() string {
	host, port, err := net.SplitHostPort(d.addr)
	if err != nil {
		return "tcp:" + d.addr
	}

	return "tcp:" + host + "," + port
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
