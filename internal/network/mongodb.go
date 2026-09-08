package network

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

const mongodbDialTimeout = 5 * time.Second

const (
	mongodbErrUserNotFound         = 11
	mongodbErrUnauthorized         = 13
	mongodbErrAuthenticationFailed = 18
)

type MongoDBClient struct {
	launcher cli.Launcher
	cfg      config.Connection
	ref      secret.Reference
	sec      secret.Provider
}

func NewMongoDBClient(
	launcher cli.Launcher,
	cfg config.Connection,
	sec secret.Provider,
) (*MongoDBClient, error) {
	ref, ok := cfg.(secret.Reference)
	if !ok {
		return nil, fmt.Errorf("connection %q does not support secrets", cfg.Meta().Name)
	}

	return &MongoDBClient{
		launcher: launcher,
		cfg:      cfg,
		ref:      ref,
		sec:      sec,
	}, nil
}

func (m *MongoDBClient) Ping(ctx context.Context) (PingResult, error) {
	dsn, err := m.dsn(ctx)
	if err != nil {
		return PingResult{}, m.fail(PingOperation, SecretErrorCode, err)
	}

	opts := options.Client().
		ApplyURI(dsn).
		SetConnectTimeout(mongodbDialTimeout).
		SetServerSelectionTimeout(mongodbDialTimeout)

	client, err := mongo.Connect(opts)
	if err != nil {
		return PingResult{}, m.fail(PingOperation, InvalidErrorCode, err)
	}
	defer func() { _ = client.Disconnect(ctx) }()

	now := time.Now()
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return PingResult{}, m.fail(PingOperation, mongodbErrorCode(err), err)
	}

	return PingResult{PingTime: time.Since(now)}, nil
}

func (m *MongoDBClient) Run(ctx context.Context) error {
	password, err := m.sec.Resolve(ctx, m.ref)
	if err != nil {
		return m.fail(ConnectOperation, SecretErrorCode, err)
	}

	if err := m.launcher.Run(ctx, m.cfg, password); err != nil {
		return m.fail(ConnectOperation, cliErrorCode(err, mongodbErrorCode), err)
	}

	return nil
}

func (m *MongoDBClient) Close() error {
	return nil
}

func (m *MongoDBClient) dsn(ctx context.Context) (string, error) {
	password, err := m.sec.Resolve(ctx, m.ref)
	if err != nil {
		return "", err
	}
	return m.cfg.ConnectionString(password), nil
}

func (m *MongoDBClient) fail(op Operation, code ErrorCode, err error) error {
	if err == nil {
		return nil
	}

	return &OpError{
		Op:     op,
		Code:   code,
		Target: mongodbTarget(m.cfg),
		During: during(op, m.cfg),
		Err:    err,
	}
}

func mongodbErrorCode(err error) ErrorCode {
	if serverErr, ok := errors.AsType[mongo.ServerError](err); ok {
		if serverErr.HasErrorCode(mongodbErrAuthenticationFailed) ||
			serverErr.HasErrorCode(mongodbErrUnauthorized) ||
			serverErr.HasErrorCode(mongodbErrUserNotFound) {
			return AuthErrorCode
		}
	}

	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "authenticationfailed"),
		strings.Contains(msg, "authentication failed"),
		strings.Contains(msg, "auth error"),
		strings.Contains(msg, "requires authentication"):
		return AuthErrorCode
	case strings.Contains(msg, "no such host"),
		strings.Contains(msg, "no route to host"):
		return UnreachableErrorCode
	case strings.Contains(msg, "x509"),
		strings.Contains(msg, "certificate"):
		return TLSErrorCode
	}

	return transportErrorCode(err)
}

func mongodbTarget(cfg config.Connection) string {
	var user string
	if cfg.Username() != "" {
		user = cfg.Username() + "@"
	}

	return fmt.Sprintf("mongodb://%s%s:%d/%s", user, cfg.Host(), cfg.Port(), cfg.Database())
}
