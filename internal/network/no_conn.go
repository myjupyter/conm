package network

import (
	"context"
	"errors"

	"github.com/myjupyter/conm/internal/config"
)

var errInvalidConfig = errors.New("invalid connection config")

type NoClient struct {
	cfg    config.Connection
	reason error
}

func NewNoClient(cfg config.Connection) (*NoClient, error) {
	return NewNoClientWithReson(cfg, errInvalidConfig)
}

func NewNoClientWithReson(cfg config.Connection, reason error) (*NoClient, error) {
	return &NoClient{cfg: cfg, reason: reason}, nil
}

func (n *NoClient) ConnType() config.ConnType {
	return n.cfg.ConnType()
}

func (n *NoClient) Ping(ctx context.Context) (PingResult, error) {
	return PingResult{}, n.fail(PingOperation)
}

func (n *NoClient) Run(ctx context.Context) error {
	return n.fail(ConnectOperation)
}

func (n *NoClient) fail(op Operation) error {
	return &OpError{
		Op:     op,
		Code:   InvalidErrorCode,
		Target: address(n.cfg),
		During: during(op, n.cfg),
		Err:    n.reason,
	}
}

func (*NoClient) Close() error {
	return nil
}
