package cli

import (
	"fmt"
	"strconv"

	"github.com/myjupyter/conm/internal/config"
)

const clickhouseClientCommand = "client"

func clickhouseCommand(cfg config.Connection, password string) (command, error) {
	ch, ok := cfg.(config.ClickHouse)
	if !ok {
		return command{}, fmt.Errorf("%w: %s cannot run a %s connection", ErrConnType, ClickHouse, cfg.ConnType())
	}

	args := []string{
		clickhouseClientCommand,
		"--host", ch.Hostname,
		"--user", ch.User,
		"--database", ch.DBName,
	}
	if ch.PortNumber != 0 {
		args = append(args, "--port", strconv.Itoa(ch.PortNumber))
	}
	if ch.Secure {
		args = append(args, "--secure")
	}

	return command{args: args, env: passwordEnv("CLICKHOUSE_PASSWORD", password)}, nil
}
