package config

import (
	"fmt"
	"os/exec"
)

type ConnType int

const (
	PostgresConnType ConnType = iota + 1
)

var (
	PGClients = []string{
		"psql",
		"pgcli",
		"usql",
	}
)

func ValidateCLI(t ConnType, cli string) error {
	var clies []string
	switch t {
	case PostgresConnType:
		clies = PGClients
	default:
		return fmt.Errorf("unknown connection type %q", t)
	}

	for _, c := range clies {
		if c == cli {
			return nil
		}
	}

	return fmt.Errorf("invalid CLI %q for connection type %q", cli, t)
}

func DetectCLI(t ConnType) []CLIInfo {
	var clies []string
	switch t {
	case PostgresConnType:
		clies = PGClients
	default:
		clies = PGClients
	}

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

func (t ConnType) String() string {
	switch t {
	case PostgresConnType:
		return "postgres"
	default:
		return "unknown"
	}
}
