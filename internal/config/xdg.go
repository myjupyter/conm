package config

import (
	"path/filepath"

	"github.com/adrg/xdg"
)

const (
	conmDirName            = "conm"
	postgresConfigFileName = "postgres.toml"
	conmConfigFileName     = "conm.toml"
	secretConfigFileName   = "secret.toml"
)

// Connection configs.
var (
	postgresPath   = filepath.Join(xdg.ConfigHome, conmDirName, postgresConfigFileName)
	PostgresPath   = pathGetter(postgresPath)
	conmConfigPath = filepath.Join(xdg.ConfigHome, conmDirName, conmConfigFileName)
	ConmPath       = pathGetter(conmConfigPath)
)

// Stored data.
var (
	secretConfigPath = filepath.Join(xdg.DataHome, conmDirName, secretConfigFileName)
	SecretConfigPath = pathGetter(secretConfigPath)
)

func pathGetter(path string) func() string {
	return func() string {
		return path
	}
}
