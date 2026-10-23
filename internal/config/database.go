package config

import (
	"slices"
)

var Databases = []ConnType{
	PostgresConnType,
	MySQLConnType,
	MSSQLConnType,
	ClickHouseConnType,
	RedisConnType,
	MongoDBConnType,
}

var ConnTypes = append(slices.Clone(Databases), SSHConnType)

type DBConnection interface {
	Connection
	Database() string
	Schema() string
	SetHost(host string) DBConnection
	SetPort(port int) DBConnection
}

func ConnTypeNames() []string {
	names := make([]string, 0, len(ConnTypes))
	for _, t := range ConnTypes {
		names = append(names, t.String())
	}

	return names
}
