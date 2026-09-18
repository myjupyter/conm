package network

import (
	"context"
	"fmt"
	"time"

	"github.com/myjupyter/conm/internal/cli"
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
	launcher cli.Launcher,
	cfg config.Connection,
	sec secret.Provider,
) (Connection, error) {
	db, isDatabase := cfg.(config.DBConnection)
	if !isDatabase {
		return nil, fmt.Errorf("connection type %q is not a database", cfg.ConnType())
	}

	switch cfg.ConnType() {
	case config.PostgresConnType:
		return NewPGClient(launcher, db, sec)
	case config.MySQLConnType:
		return NewMySQLClient(launcher, db, sec)
	case config.MSSQLConnType:
		return NewMSSQLClient(launcher, db, sec)
	case config.ClickHouseConnType:
		return NewClickHouseClient(launcher, db, sec)
	case config.RedisConnType:
		return NewRedisClient(launcher, db, sec)
	case config.MongoDBConnType:
		return NewMongoDBClient(launcher, db, sec)
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
