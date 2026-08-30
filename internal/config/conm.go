package config

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/pelletier/go-toml/v2"
)

type ConnMeta struct {
	Name        string   `toml:"name,omitempty"`
	Description string   `toml:"description,omitempty"`
	Tags        []string `toml:"tags,omitempty"`
}

type ConnType int

const (
	PostgresConnType ConnType = iota + 1
	MySQLConnType
	MSSQLConnType
	ClickHouseConnType
	RedisConnType
)

// databaseTypes names every database conm can manage, in the order the setup
// table lists them.
var databaseTypes = []ConnType{
	PostgresConnType,
	MySQLConnType,
	MSSQLConnType,
	ClickHouseConnType,
	RedisConnType,
}

type Conm struct {
	Databases []Database `toml:"database"`
}

type Database struct {
	Type    ConnType `toml:"type"`
	CLI     string   `toml:"cli"`
	Enabled bool     `toml:"enabled"`
}

func (c Conm) Database(t ConnType) (Database, bool) {
	for _, db := range c.Databases {
		if db.Type == t {
			return db, true
		}
	}
	return Database{}, false
}

func (c Conm) CLI(t ConnType) string {
	db, _ := c.Database(t)
	return db.CLI
}

func (c *Conm) SetDatabase(d Database) error {
	if err := ValidateCLI(d.Type, d.CLI); err != nil {
		return err
	}

	for i, db := range c.Databases {
		if db.Type == d.Type {
			c.Databases[i] = d
			return nil
		}
	}

	c.Databases = append(c.Databases, d)
	return nil
}

// withDatabaseTypes keeps one row per database conm knows, in that order, so
// every screen sees the same list whatever the file holds.
func withDatabaseTypes(stored []Database) []Database {
	dbs := make([]Database, 0, len(databaseTypes))
	for _, t := range databaseTypes {
		db := Database{Type: t}
		for _, s := range stored {
			if s.Type == t {
				db = s
				break
			}
		}
		dbs = append(dbs, db)
	}
	return dbs
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
	default:
		return fmt.Errorf("unsupported connection type %q", t)
	}
}

func (t ConnType) String() string {
	switch t {
	case PostgresConnType:
		return "postgres"
	case MySQLConnType:
		return "mysql"
	case MSSQLConnType:
		return "mssql"
	case ClickHouseConnType:
		return "clickhouse"
	case RedisConnType:
		return "redis"
	default:
		return "unknown"
	}
}

func ParseConnType(name string) (ConnType, error) {
	for _, t := range databaseTypes {
		if t.String() == name {
			return t, nil
		}
	}
	return 0, fmt.Errorf("unknown connection type %q", name)
}

func (t ConnType) MarshalText() ([]byte, error) {
	if !slices.Contains(databaseTypes, t) {
		return nil, fmt.Errorf("unsupported connection type %d", int(t))
	}
	return []byte(t.String()), nil
}

func (t *ConnType) UnmarshalText(raw []byte) error {
	parsed, err := ParseConnType(string(raw))
	if err != nil {
		return err
	}

	*t = parsed
	return nil
}

type ConmConfigWrapper struct {
	Conm Conm `toml:"conm"`
}

func (w *ConmConfigWrapper) Add(conm Conm) {
	w.Conm = conm
}

func (w *ConmConfigWrapper) Len() int {
	return 1
}

func (w *ConmConfigWrapper) Get(_ int) Conm {
	return w.Conm
}

func (w *ConmConfigWrapper) Put(_ int, conm Conm) {
	w.Conm = conm
}

func (w *ConmConfigWrapper) ConnectionConfigs() []Connection {
	return nil
}

func (w *ConmConfigWrapper) Remove(_ int) {}

func (w *ConmConfigWrapper) Validate() {
	w.Conm.Databases = withDatabaseTypes(w.Conm.Databases)
}

func (w *ConmConfigWrapper) Unmarshal(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	return toml.Unmarshal(data, w)
}

func (w *ConmConfigWrapper) Marshal() ([]byte, error) {
	return toml.Marshal(w)
}

// CreateConmConfigPath creates both the config dir (conm.toml, postgres.toml)
// and the data dir (secret.toml); XDG may place them under different roots.
func CreateConmConfigPath() error {
	dirs := []string{filepath.Dir(conmConfigPath), filepath.Dir(secretConfigPath)}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return err
		}
	}

	return nil
}

func CreateConm(conf Conm) error {
	raw, err := toml.Marshal(struct {
		Conm Conm `toml:"conm"`
	}{Conm: conf})
	if err != nil {
		return err
	}

	return os.WriteFile(ConmPath(), raw, 0o600)
}

func ReadConm(filename string) (Conm, error) {
	t := struct {
		Conm Conm `toml:"conm"`
	}{}

	raw, err := os.ReadFile(filename)
	if err != nil {
		return Conm{}, err
	}

	err = toml.Unmarshal(raw, &t)
	if err != nil {
		return Conm{}, err
	}

	t.Conm.Databases = withDatabaseTypes(t.Conm.Databases)

	return t.Conm, nil
}
