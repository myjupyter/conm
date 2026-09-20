package config

import (
	"fmt"
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
}

func ConnTypeNames() []string {
	names := make([]string, 0, len(ConnTypes))
	for _, t := range ConnTypes {
		names = append(names, t.String())
	}

	return names
}

func CreateDatabaseConfig(t ConnType) error {
	switch t {
	case PostgresConnType:
		c, err := OpenConfig[*PostgresConfigWrapper](PostgresPath())
		if err != nil {
			return err
		}
		defer c.Close()

		return c.Save()
	case MySQLConnType:
		c, err := OpenConfig[*MySQLConfigWrapper](MySQLPath())
		if err != nil {
			return err
		}
		defer c.Close()

		return c.Save()
	case MSSQLConnType:
		c, err := OpenConfig[*MSSQLConfigWrapper](MSSQLPath())
		if err != nil {
			return err
		}
		defer c.Close()

		return c.Save()
	case ClickHouseConnType:
		c, err := OpenConfig[*ClickHouseConfigWrapper](ClickHousePath())
		if err != nil {
			return err
		}
		defer c.Close()

		return c.Save()
	case RedisConnType:
		c, err := OpenConfig[*RedisConfigWrapper](RedisPath())
		if err != nil {
			return err
		}
		defer c.Close()

		return c.Save()
	case MongoDBConnType:
		c, err := OpenConfig[*MongoDBConfigWrapper](MongoDBPath())
		if err != nil {
			return err
		}
		defer c.Close()

		return c.Save()
	case SSHConnType:
		c, err := OpenConfig[*SSHConfigWrapper](SSHPath())
		if err != nil {
			return err
		}
		defer c.Close()

		return c.Save()
	default:
		return fmt.Errorf("unsupported connection type %q", t)
	}
}
