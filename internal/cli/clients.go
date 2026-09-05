package cli

import (
	"fmt"
	"slices"

	"github.com/myjupyter/conm/internal/config"
)

const (
	Psql       = "psql"
	Pgcli      = "pgcli"
	USQL       = "usql"
	MySQL      = "mysql"
	MyCLI      = "mycli"
	SQLCmd     = "sqlcmd"
	ClickHouse = "clickhouse"
	RedisCLI   = "redis-cli"
	ValkeyCLI  = "valkey-cli"
	IRedis     = "iredis"
	Mongosh    = "mongosh"
	Mongo      = "mongo"
)

var clients = map[config.ConnType][]string{
	config.PostgresConnType:   {Psql, Pgcli, USQL},
	config.MySQLConnType:      {MySQL, MyCLI, USQL},
	config.MSSQLConnType:      {SQLCmd, USQL},
	config.ClickHouseConnType: {ClickHouse, USQL},
	config.RedisConnType:      {RedisCLI, ValkeyCLI, IRedis},
	config.MongoDBConnType:    {Mongosh, Mongo},
}

func Clients(t config.ConnType) []string {
	return clients[t]
}

func Validate(t config.ConnType, name string) error {
	names := Clients(t)
	if len(names) == 0 {
		return fmt.Errorf("unknown connection type %q", t)
	}

	if slices.Contains(names, name) {
		return nil
	}

	return fmt.Errorf("invalid CLI %q for connection type %q", name, t)
}
