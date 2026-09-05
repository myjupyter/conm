package network

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/myjupyter/conm/internal/cli"
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
	launcher cli.Launcher
	cfg      config.Connection
	ref      secret.Reference
	sec      secret.Provider
}

func NewRedisClient(
	launcher cli.Launcher,
	cfg config.Connection,
	sec secret.Provider,
) (*RedisClient, error) {
	ref, ok := cfg.(secret.Reference)
	if !ok {
		return nil, fmt.Errorf("connection %q does not support secrets", cfg.Name())
	}

	return &RedisClient{
		launcher: launcher,
		cfg:      cfg,
		ref:      ref,
		sec:      sec,
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
	password, err := r.sec.Resolve(ctx, r.ref)
	if err != nil {
		return r.fail(ConnectOperation, SecretErrorCode, err)
	}

	if err := r.launcher.Run(ctx, r.cfg, password); err != nil {
		return r.fail(ConnectOperation, cliErrorCode(err, redisErrorCode), err)
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
