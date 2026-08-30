package network

import (
	"context"
	"fmt"
	"time"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

// connection quality intervals.
const (
	Slowest = 500 * time.Millisecond
	Slow    = 250 * time.Millisecond
	Medium  = 120 * time.Millisecond
	Fast    = 60 * time.Millisecond
)

type Connection interface {
	Ping(ctx context.Context) (PingResult, error)
	Run(ctx context.Context) error
	Close() error
}

type PingResult struct {
	PingTime time.Duration
}

func NewConnection(
	conmConfig config.Conm,
	cfg config.Connection,
	sec secret.Provider,
) (Connection, error) {
	switch cfg.ConnType() {
	case config.PostgresConnType:
		return NewPGClient(conmConfig, cfg, sec)
	case config.RedisConnType:
		return NewRedisClient(conmConfig, cfg, sec)
	default:
		return nil, fmt.Errorf("unsupported connection type %q", cfg.ConnType())
	}
}

func (r PingResult) Bars() int {
	switch {
	case r.PingTime < Fast:
		return 4
	case r.PingTime < Medium:
		return 3
	case r.PingTime < Slow:
		return 2
	case r.PingTime < Slowest:
		return 1
	default:
		return 0
	}
}
