package importer

import (
	"fmt"

	"github.com/myjupyter/conm/internal/config"
)

type Entry struct {
	Name     string
	Conn     config.Connection
	Warnings []string
	Err      error
}

type source struct {
	path func() string
	read func(path string) ([]Entry, error)
}

var sources = map[config.ConnType]source{
	config.PostgresConnType:   {path: config.PGServicePath, read: readPostgres},
	config.MySQLConnType:      {path: config.MySQLOptionPath, read: readMySQL},
	config.ClickHouseConnType: {path: config.ClickHouseClientPath, read: readClickHouse},
	config.MSSQLConnType:      {path: config.SQLConfigPath, read: readMSSQL},
	config.SSHConnType:        {path: config.SSHConfigPath, read: readSSH},
}

func Supported(t config.ConnType) bool {
	_, ok := sources[t]
	return ok
}

func DefaultPath(t config.ConnType) string {
	return sources[t].path()
}

func Read(t config.ConnType, path string) ([]Entry, error) {
	src, ok := sources[t]
	if !ok {
		return nil, fmt.Errorf("%s has no config file to import from", t)
	}

	return src.read(path)
}

func values(params []config.ForeignParam) map[string]string {
	found := make(map[string]string, len(params))
	for _, param := range params {
		found[param.Key] = param.Value
	}

	return found
}
