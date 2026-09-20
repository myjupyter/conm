package config

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/pelletier/go-toml/v2"
)

type ConnType int

const (
	PostgresConnType ConnType = iota + 1
	MySQLConnType
	MSSQLConnType
	ClickHouseConnType
	RedisConnType
	MongoDBConnType
	SSHConnType
)

type ConnKind int

const (
	DatabaseConnKind ConnKind = iota + 1
	SSHConnKind
)

type Conm struct {
	Connections []ConnectionSettings `toml:"connection"`
}

type ConnectionSettings struct {
	Kind    ConnKind `toml:"kind"`
	Type    ConnType `toml:"type"`
	CLI     string   `toml:"cli"`
	Enabled bool     `toml:"enabled"`
}

func (c Conm) Connection(t ConnType) (ConnectionSettings, bool) {
	for _, conn := range c.Connections {
		if conn.Type == t {
			return conn, true
		}
	}
	return ConnectionSettings{}, false
}

func (c Conm) CLI(t ConnType) string {
	conn, _ := c.Connection(t)
	return conn.CLI
}

func (c *Conm) SetConnection(s ConnectionSettings) {
	for i, conn := range c.Connections {
		if conn.Type == s.Type {
			c.Connections[i] = s
			return
		}
	}

	c.Connections = append(c.Connections, s)
}

func withKnownTypes(stored []ConnectionSettings) []ConnectionSettings {
	conns := make([]ConnectionSettings, 0, len(ConnTypes))
	for _, t := range ConnTypes {
		conn := ConnectionSettings{Type: t}
		for _, s := range stored {
			if s.Type == t {
				conn = s
				break
			}
		}
		conn.Kind = t.Kind()
		conns = append(conns, conn)
	}
	return conns
}

func (t ConnType) Kind() ConnKind {
	if t == SSHConnType {
		return SSHConnKind
	}
	return DatabaseConnKind
}

func (k ConnKind) String() string {
	switch k {
	case DatabaseConnKind:
		return "database"
	case SSHConnKind:
		return "ssh"
	default:
		return "unknown"
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
	case MongoDBConnType:
		return "mongodb"
	case SSHConnType:
		return "ssh"
	default:
		return "unknown"
	}
}

func ParseConnType(name string) (ConnType, error) {
	for _, t := range ConnTypes {
		if t.String() == name {
			return t, nil
		}
	}
	return 0, fmt.Errorf("unknown connection type %q", name)
}

func (t ConnType) MarshalText() ([]byte, error) {
	if !slices.Contains(ConnTypes, t) {
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

func (k ConnKind) MarshalText() ([]byte, error) {
	if k != DatabaseConnKind && k != SSHConnKind {
		return nil, fmt.Errorf("unsupported connection kind %d", int(k))
	}
	return []byte(k.String()), nil
}

func (k *ConnKind) UnmarshalText(raw []byte) error {
	parsed := string(raw)
	switch parsed {
	case "database":
		*k = DatabaseConnKind
	case "ssh":
		*k = SSHConnKind
	default:
		return fmt.Errorf("unknown connection kind %q", parsed)
	}

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
	w.Conm.Connections = withKnownTypes(w.Conm.Connections)
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

	t.Conm.Connections = withKnownTypes(t.Conm.Connections)

	return t.Conm, nil
}
