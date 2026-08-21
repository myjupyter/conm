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

	if slices.Contains(clies, cli) {
		return nil
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
