package importer

import (
	"fmt"

	"github.com/myjupyter/conm/internal/config"
)

type Entry struct {
	Name     string
	Conn     config.Connection
	Password string
	Warnings []string
	Err      error
}

type source struct {
	path func() string
	read func(path string) ([]Entry, error)
}

var sources = map[config.ConnType]source{
	config.PostgresConnType: {path: config.PGServicePath, read: readPostgres},
	config.MySQLConnType:    {path: config.MySQLOptionPath, read: readMySQL},
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

func SecretLocation(t config.ConnType, name string) string {
	return t.String() + "/" + name
}
