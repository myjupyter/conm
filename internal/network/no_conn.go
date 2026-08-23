package network

import (
	"context"
	"errors"

	"github.com/myjupyter/conm/internal/config"
)

var errInvalidConfig = errors.New("invalid connection config")

type NoClient struct {
	cfg config.Connection
}

func NewNoClient(cfg config.Connection) (*NoClient, error) {
	return &NoClient{cfg: cfg}, nil
}

func (n *NoClient) ConnType() config.ConnType {
	return n.cfg.ConnType()
}

func (n *NoClient) Name() string {
	return n.cfg.Name()
}

func (n *NoClient) Description() string {
	return n.cfg.Description()
}

func (n *NoClient) Tags() []string {
	return n.cfg.Tags()
}

func (n *NoClient) Host() string {
	return n.cfg.Host()
}

func (n *NoClient) Username() string {
	return n.cfg.Username()
}

func (n *NoClient) Port() int {
	return n.cfg.Port()
}

func (n *NoClient) Database() string {
	return n.cfg.Database()
}

func (n *NoClient) Schema() string {
	return n.cfg.Schema()
}

func (n *NoClient) ConnectionString(secret string) string {
	return n.cfg.ConnectionString(secret)
}

func (n *NoClient) IsValid() bool {
	return n.cfg.IsValid()
}

func (n *NoClient) Validate() []error {
	return n.cfg.Validate()
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
		Err:    errInvalidConfig,
	}
}

func (*NoClient) Close() error {
	return nil
}
