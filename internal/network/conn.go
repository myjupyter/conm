package network

import (
	"context"
	"fmt"

	"github.com/myjupyter/conm/internal/config"
)

type Connection interface {
	config.Connection

	Ping(ctx context.Context) error
	Run(ctx context.Context) error
	Close() error
}

func NewConnection(
	conmConfig config.Conm,
	cfg config.Connection,
) (Connection, error) {
	switch cfg.ConnType() {
	case config.PostgresConnType:
		return NewPGClient(conmConfig, cfg)
	default:
		return nil, fmt.Errorf("unsupported connection type %q", cfg.ConnType())
	}
}
