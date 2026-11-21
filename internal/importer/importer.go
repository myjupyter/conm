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
	config.PostgresConnType:   {path: PGServicePath, read: readPostgres},
	config.MySQLConnType:      {path: MySQLOptionPath, read: readMySQL},
	config.ClickHouseConnType: {path: ClickHouseClientPath, read: readClickHouse},
	config.MSSQLConnType:      {path: SQLConfigPath, read: readMSSQL},
	config.SSHConnType:        {path: SSHConfigPath, read: readSSH},
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

func values(params []Param) map[string]string {
	found := make(map[string]string, len(params))
	for _, param := range params {
		found[param.Key] = param.Value
	}

	return found
}

type Section struct {
	Name   string
	Params []Param
}

type Param struct {
	Key   string
	Value string
}
