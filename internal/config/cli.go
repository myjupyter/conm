package config

import (
	"fmt"
	"os/exec"
	"slices"
)

type CLIInfo struct {
	Name   string
	Path   string
	Exists bool
}

const (
	MySQLCLI      = "mysql"
	MyCLI         = "mycli"
	SQLCmdCLI     = "sqlcmd"
	USQLCLI       = "usql"
	ClickHouseCLI = "clickhouse"
	RedisCLI      = "redis-cli"
	ValkeyCLI     = "valkey-cli"
	IRedisCLI     = "iredis"
)

var (
	PGClients = []string{
		"psql",
		"pgcli",
		USQLCLI,
	}

	MySQLClients = []string{
		MySQLCLI,
		MyCLI,
		USQLCLI,
	}

	MSSQLClients = []string{
		SQLCmdCLI,
		USQLCLI,
	}

	ClickHouseClients = []string{
		ClickHouseCLI,
		USQLCLI,
	}

	RedisClients = []string{
		RedisCLI,
		ValkeyCLI,
		IRedisCLI,
	}
)

func Clients(t ConnType) []string {
	switch t {
	case PostgresConnType:
		return PGClients
	case MySQLConnType:
		return MySQLClients
	case MSSQLConnType:
		return MSSQLClients
	case ClickHouseConnType:
		return ClickHouseClients
	case RedisConnType:
		return RedisClients
	default:
		return nil
	}
}

func ValidateCLI(t ConnType, cli string) error {
	clies := Clients(t)
	if len(clies) == 0 {
		return fmt.Errorf("unknown connection type %q", t)
	}

	if slices.Contains(clies, cli) {
		return nil
	}

	return fmt.Errorf("invalid CLI %q for connection type %q", cli, t)
}

func DetectCLI(t ConnType) []CLIInfo {
	clies := Clients(t)

	infos := make([]CLIInfo, 0, len(clies))
	for _, name := range clies {
		path, err := exec.LookPath(name)
		infos = append(infos, CLIInfo{
			Name:   name,
			Path:   path,
			Exists: err == nil,
		})
	}

	return infos
}

func FindCLI(clies []CLIInfo, name string) (CLIInfo, bool) {
	for _, cli := range clies {
		if cli.Name == name {
			return cli, true
		}
	}
	return CLIInfo{}, false
}
