package cli

import (
	"context"
	"os"
	"os/exec"

	"github.com/myjupyter/conm/internal/config"
)

type builder func(cfg config.Connection, password string) (command, error)

type command struct {
	args []string
	env  []string
}

var commands = map[string]builder{
	Psql:       urlCommand,
	Pgcli:      urlCommand,
	USQL:       urlCommand,
	Mongosh:    urlCommand,
	Mongo:      urlCommand,
	MyCLI:      urlCommand,
	RedisCLI:   urlFlagCommand("-u"),
	ValkeyCLI:  urlFlagCommand("-u"),
	IRedis:     urlFlagCommand("--url"),
	MySQL:      mysqlCommand,
	SQLCmd:     mssqlCommand,
	ClickHouse: clickhouseCommand,
}

func urlCommand(cfg config.Connection, password string) (command, error) {
	return command{args: []string{cfg.ConnectionString(password)}}, nil
}

func urlFlagCommand(flag string) builder {
	return func(cfg config.Connection, password string) (command, error) {
		return command{args: []string{flag, cfg.ConnectionString(password)}}, nil
	}
}

func passwordEnv(name, password string) []string {
	if password == "" {
		return nil
	}

	return append(os.Environ(), name+"="+password)
}

func (c command) run(ctx context.Context, name string) error {
	executor := exec.CommandContext(ctx, name, c.args...)
	executor.Env = c.env

	executor.Stdin = os.Stdin
	executor.Stdout = os.Stdout
	executor.Stderr = os.Stderr

	return executor.Run()
}
