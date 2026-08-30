package network

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

const redisSingleAttempt = -1

type discardRedisLogger struct{}

func (discardRedisLogger) Printf(context.Context, string, ...any) {}

func init() {
	redis.SetLogger(discardRedisLogger{})
}

type RedisClient struct {
	conmCfg config.Conm
	cfg     config.Connection
	ref     secret.Reference
	sec     secret.Provider
}

func NewRedisClient(
	conmConfig config.Conm,
	cfg config.Connection,
	sec secret.Provider,
) (*RedisClient, error) {
	ref, ok := cfg.(secret.Reference)
	if !ok {
		return nil, fmt.Errorf("connection %q does not support secrets", cfg.Name())
	}

	return &RedisClient{
		conmCfg: conmConfig,
		cfg:     cfg,
		ref:     ref,
		sec:     sec,
	}, nil
}

func (r *RedisClient) Ping(ctx context.Context) (PingResult, error) {
	dsn, err := r.dsn(ctx)
	if err != nil {
		return PingResult{}, r.fail(PingOperation, SecretErrorCode, err)
	}

	opts, err := redis.ParseURL(dsn)
	if err != nil {
		return PingResult{}, r.fail(PingOperation, InvalidErrorCode, err)
	}
	opts.MaxRetries = redisSingleAttempt

	client := redis.NewClient(opts)
	defer client.Close()

	now := time.Now()
	if err := client.Ping(ctx).Err(); err != nil {
		return PingResult{}, r.fail(PingOperation, redisErrorCode(err), err)
	}

	return PingResult{PingTime: time.Since(now)}, nil
}

func (r *RedisClient) Run(ctx context.Context) error {
	dsn, err := r.dsn(ctx)
	if err != nil {
		return r.fail(ConnectOperation, SecretErrorCode, err)
	}

	cli := r.conmCfg.CLI(config.RedisConnType)
	if cli == "" {
		return r.fail(ConnectOperation, InvalidErrorCode, errors.New("no redis client configured, run conm init"))
	}

	if err := runCLI(ctx, cli, redisArgs(cli, dsn), nil); err != nil {
		return r.fail(ConnectOperation, redisErrorCode(err), err)
	}

	return nil
}

func (r *RedisClient) Close() error {
	return nil
}

func (r *RedisClient) dsn(ctx context.Context) (string, error) {
	password, err := r.sec.Resolve(ctx, r.ref)
	if err != nil {
		return "", err
	}
	return r.cfg.ConnectionString(password), nil
}

func (r *RedisClient) fail(op Operation, code ErrorCode, err error) error {
	if err == nil {
		return nil
	}

	return &OpError{
		Op:     op,
		Code:   code,
		Target: redisTarget(r.cfg),
		During: during(op, r.cfg),
		Err:    err,
	}
}

func redisArgs(cli, dsn string) []string {
	if cli == config.IRedisCLI {
		return []string{"--url", dsn}
	}
	return []string{"-u", dsn}
}

func redisErrorCode(err error) ErrorCode {
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "wrongpass"),
		strings.Contains(msg, "noauth"),
		strings.Contains(msg, "invalid username-password pair"),
		strings.Contains(msg, "without any password configured"):
		return AuthErrorCode
	case strings.Contains(msg, "noperm"):
		return AuthErrorCode
	}

	return transportErrorCode(err)
}

func redisTarget(cfg config.Connection) string {
	var user string
	if cfg.Username() != "" {
		user = cfg.Username() + "@"
	}

	return fmt.Sprintf("redis://%s%s:%d/%s", user, cfg.Host(), cfg.Port(), cfg.Database())
}
