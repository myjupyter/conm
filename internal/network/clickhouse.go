package network

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

const clickhouseDialTimeout = 5 * time.Second

// The clickhouse binary is a multi-call tool: the interactive client is a
// subcommand of it rather than a program of its own.
const clickhouseClientCommand = "client"

const (
	clickhouseErrAuthFailed     = 516
	clickhouseErrUnknownUser    = 192
	clickhouseErrAccessDenied   = 497
	clickhouseErrUnknownDB      = 81
	clickhouseErrDBDoesNotExist = 82
)

type ClickHouseClient struct {
	conmCfg config.Conm
	cfg     config.Connection
	ref     secret.Reference
	sec     secret.Provider
}

func NewClickHouseClient(
	conmConfig config.Conm,
	cfg config.Connection,
	sec secret.Provider,
) (*ClickHouseClient, error) {
	ref, ok := cfg.(secret.Reference)
	if !ok {
		return nil, fmt.Errorf("connection %q does not support secrets", cfg.Name())
	}

	return &ClickHouseClient{
		conmCfg: conmConfig,
		cfg:     cfg,
		ref:     ref,
		sec:     sec,
	}, nil
}

func (c *ClickHouseClient) Ping(ctx context.Context) (PingResult, error) {
	raw, err := c.dsn(ctx)
	if err != nil {
		return PingResult{}, c.fail(PingOperation, SecretErrorCode, err)
	}

	target, err := parseClickHouseDSN(raw)
	if err != nil {
		return PingResult{}, c.fail(PingOperation, InvalidErrorCode, err)
	}

	opts, err := clickhouse.ParseDSN(target.driverDSN())
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
	raw, err := c.dsn(ctx)
	if err != nil {
		return c.fail(ConnectOperation, SecretErrorCode, err)
	}

	target, err := parseClickHouseDSN(raw)
	if err != nil {
		return c.fail(ConnectOperation, InvalidErrorCode, err)
	}

	cli := c.conmCfg.CLI(config.ClickHouseConnType)
	if cli == "" {
		return c.fail(ConnectOperation, InvalidErrorCode, errors.New("no clickhouse client configured, run conm init"))
	}

	if err := runCLI(ctx, cli, target.args(cli, raw), target.env(cli)); err != nil {
		return c.fail(ConnectOperation, clickhouseErrorCode(err), err)
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

// clickhouseDSN is the clickhouse:// URL the driver parses, kept apart so
// clickhouse-client can be handed the same connection as flags with the
// password out of band.
type clickhouseDSN struct {
	raw      *url.URL
	addr     string
	user     string
	password string
	database string
	secure   bool
}

func parseClickHouseDSN(raw string) (clickhouseDSN, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return clickhouseDSN{}, err
	}

	password, _ := u.User.Password()

	return clickhouseDSN{
		raw:      u,
		addr:     u.Host,
		user:     u.User.Username(),
		password: password,
		database: strings.TrimPrefix(u.Path, "/"),
		secure:   u.Query().Get("secure") == "true",
	}, nil
}

func (d clickhouseDSN) driverDSN() string {
	u := *d.raw
	q := u.Query()
	q.Set("dial_timeout", clickhouseDialTimeout.String())
	u.RawQuery = q.Encode()

	return u.String()
}

func (d clickhouseDSN) args(cli, raw string) []string {
	if cli != config.ClickHouseCLI {
		return []string{raw}
	}

	host, port := d.hostPort()
	args := []string{clickhouseClientCommand, "--host", host, "--user", d.user, "--database", d.database}
	if port != "" {
		args = append(args, "--port", port)
	}
	if d.secure {
		args = append(args, "--secure")
	}

	return args
}

// env keeps the password out of the process table: the clickhouse client reads
// it from CLICKHOUSE_PASSWORD, while a client taking the URL already carries it.
func (d clickhouseDSN) env(cli string) []string {
	if cli != config.ClickHouseCLI || d.password == "" {
		return nil
	}

	return append(os.Environ(), "CLICKHOUSE_PASSWORD="+d.password)
}

func (d clickhouseDSN) hostPort() (string, string) {
	host, port, err := net.SplitHostPort(d.addr)
	if err != nil {
		return d.addr, ""
	}
	if _, err := strconv.Atoi(port); err != nil {
		return host, ""
	}

	return host, port
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
