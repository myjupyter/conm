package network

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

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
	conmCfg config.Conm
	cfg     config.Connection
	ref     secret.Reference
	sec     secret.Provider
}

func NewMySQLClient(
	conmConfig config.Conm,
	cfg config.Connection,
	sec secret.Provider,
) (*MySQLClient, error) {
	ref, ok := cfg.(secret.Reference)
	if !ok {
		return nil, fmt.Errorf("connection %q does not support secrets", cfg.Name())
	}

	return &MySQLClient{
		conmCfg: conmConfig,
		cfg:     cfg,
		ref:     ref,
		sec:     sec,
	}, nil
}

func (m *MySQLClient) dsn(ctx context.Context) (string, error) {
	password, err := m.sec.Resolve(ctx, m.ref)
	if err != nil {
		return "", err
	}
	return m.cfg.ConnectionString(password), nil
}

func (m *MySQLClient) Ping(ctx context.Context) (PingResult, error) {
	raw, err := m.dsn(ctx)
	if err != nil {
		return PingResult{}, m.fail(PingOperation, SecretErrorCode, err)
	}

	target, err := parseMySQLDSN(raw)
	if err != nil {
		return PingResult{}, m.fail(PingOperation, InvalidErrorCode, err)
	}

	db, err := sql.Open("mysql", target.driverDSN())
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
	raw, err := m.dsn(ctx)
	if err != nil {
		return m.fail(ConnectOperation, SecretErrorCode, err)
	}

	target, err := parseMySQLDSN(raw)
	if err != nil {
		return m.fail(ConnectOperation, InvalidErrorCode, err)
	}

	cli := m.conmCfg.CLI(config.MySQLConnType)
	if cli == "" {
		return m.fail(ConnectOperation, InvalidErrorCode, errors.New("no mysql client configured, run conm init"))
	}

	executor := exec.CommandContext(ctx, cli, target.args(cli, raw)...)
	executor.Env = target.env(cli)

	executor.Stdin = os.Stdin
	executor.Stdout = os.Stdout
	executor.Stderr = os.Stderr

	if err := executor.Run(); err != nil {
		return m.fail(ConnectOperation, mysqlErrorCode(err), err)
	}

	return nil
}

func (m *MySQLClient) Close() error {
	return nil
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

// mysqlDSN is the one connection string in the three dialects MySQL speaks:
// config builds a mysql:// URL, the driver wants user:pass@tcp(addr)/db, and
// the stock client takes flags with the password out of band.
type mysqlDSN struct {
	addr     string
	user     string
	password string
	database string
	tls      string
}

func parseMySQLDSN(raw string) (mysqlDSN, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return mysqlDSN{}, err
	}

	password, _ := u.User.Password()

	return mysqlDSN{
		addr:     u.Host,
		user:     u.User.Username(),
		password: password,
		database: strings.TrimPrefix(u.Path, "/"),
		tls:      u.Query().Get("tls"),
	}, nil
}

func (d mysqlDSN) driverDSN() string {
	cfg := mysql.NewConfig()
	cfg.Net = "tcp"
	cfg.Addr = d.addr
	cfg.User = d.user
	cfg.Passwd = d.password
	cfg.DBName = d.database
	cfg.TLSConfig = d.tls
	cfg.Timeout = mysqlDialTimeout

	return cfg.FormatDSN()
}

func (d mysqlDSN) args(cli, raw string) []string {
	if cli != config.MySQLCLI {
		return []string{raw}
	}

	host, port := d.hostPort()
	args := []string{"--protocol=TCP", "--host=" + host, "--user=" + d.user}
	if port != "" {
		args = append(args, "--port="+port)
	}
	if mode := d.sslMode(); mode != "" {
		args = append(args, "--ssl-mode="+mode)
	}

	return append(args, d.database)
}

// env keeps the password out of the process table: the stock client reads it
// from MYSQL_PWD, while the clients that take a URL already carry it inside.
func (d mysqlDSN) env(cli string) []string {
	if cli != config.MySQLCLI || d.password == "" {
		return nil
	}

	return append(os.Environ(), "MYSQL_PWD="+d.password)
}

func (d mysqlDSN) hostPort() (string, string) {
	host, port, err := net.SplitHostPort(d.addr)
	if err != nil {
		return d.addr, ""
	}
	if _, err := strconv.Atoi(port); err != nil {
		return host, ""
	}

	return host, port
}

func (d mysqlDSN) sslMode() string {
	switch d.tls {
	case config.MySQLTLSModeDisable:
		return "DISABLED"
	case config.MySQLTLSModePreferred:
		return "PREFERRED"
	case config.MySQLTLSModeSkipVerify:
		return "REQUIRED"
	case config.MySQLTLSModeVerify:
		return "VERIFY_CA"
	}

	return ""
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
