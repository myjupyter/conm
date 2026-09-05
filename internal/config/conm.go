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
	MongoDBConnType
)

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

func (c *Conm) SetDatabase(d Database) {
	for i, db := range c.Databases {
		if db.Type == d.Type {
			c.Databases[i] = d
			return
		}
	}

	c.Databases = append(c.Databases, d)
}

func withDatabaseTypes(stored []Database) []Database {
	dbs := make([]Database, 0, len(Databases))
	for _, t := range Databases {
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
	case MongoDBConnType:
		return "mongodb"
	default:
		return "unknown"
	}
}

func ParseConnType(name string) (ConnType, error) {
	for _, t := range Databases {
		if t.String() == name {
			return t, nil
		}
	}
	return 0, fmt.Errorf("unknown connection type %q", name)
}

func (t ConnType) MarshalText() ([]byte, error) {
	if !slices.Contains(Databases, t) {
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
