package config

import (
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type PostgresSSLMode = string

const (
	PostgresSSLModeDisable    PostgresSSLMode = "disable"
	PostgresSSLModeAllow      PostgresSSLMode = "allow"
	PostgresSSLModePrefer     PostgresSSLMode = "prefer"
	PostgresSSLModeRequire    PostgresSSLMode = "require"
	PostgresSSLModeVerifyCA   PostgresSSLMode = "verify-ca"
	PostgresSSLModeVerifyFull PostgresSSLMode = "verify-full"
)

type PostgresFormField = string

const (
	PostgresFormFieldName        PostgresFormField = "Name"
	PostgresFormFieldDescription PostgresFormField = "Description"
	PostgresFormFieldTags        PostgresFormField = "Tags"
	PostgresFormFieldHost        PostgresFormField = "Host"
	PostgresFormFieldPort        PostgresFormField = "Port"
	PostgresFormFieldUsername    PostgresFormField = "Username"
	PostgresFormFieldPassword    PostgresFormField = "Password"
	PostgresFormFieldDatabase    PostgresFormField = "Database"
	PostgresFormFieldSSLMode     PostgresFormField = "SSLMode"
)

type Postgres struct {
	Meta       ConnMeta `toml:"meta"`
	Hostname   string   `toml:"host" json:"host"`
	PortNumber int      `toml:"port" json:"port"`
	User       string   `toml:"username" json:"username"`
	Password   string   `toml:"password" json:"password"`
	DBName     string   `toml:"dbname" json:"dbname"`
	SSLMode    string   `toml:"sslmode,omitempty" json:"sslmode,omitempty"`
}

type PostgresConfigWrapper struct {
	Conns []Postgres `toml:"postgres"`
}

func ImportFromPGPass() ([]Postgres, error) {
	pgPassPath := os.Getenv("PGPASSFILE")
	if pgPassPath == "" {
		homeDir := os.Getenv("HOME")
		const pgpassFile = ".pgpass"
		pgPassPath = filepath.Join(homeDir, pgpassFile)
	}

	raw, err := os.ReadFile(pgPassPath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(raw), "\n")

	var confs []Postgres
	for _, line := range lines {
		values := strings.Split(line, ":")
		if skipPGPassRow(values) {
			continue
		}

		port, err := strconv.Atoi(values[1])
		if err != nil {
			continue
		}

		// .pgpass format: hostname:port:database:username:password
		conf := Postgres{
			Hostname:   values[0],
			PortNumber: port,
			DBName:     values[2],
			User:       values[3],
			Password:   values[4],
			Meta: ConnMeta{
				Name: values[0],
			},
		}

		confs = append(confs, conf)
	}

	return confs, nil
}

func (w *PostgresConfigWrapper) Add(conn Postgres) {
	w.Conns = append(w.Conns, conn)
}

func (w *PostgresConfigWrapper) Unmarshal(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	return toml.Unmarshal(data, w)
}

func (w *PostgresConfigWrapper) Marshal() ([]byte, error) {
	return toml.Marshal(w)
}

func (w *PostgresConfigWrapper) Len() int {
	return len(w.Conns)
}

func (w *PostgresConfigWrapper) Get(i int) Postgres {
	return w.Conns[i]
}

func (w *PostgresConfigWrapper) Put(i int, conn Postgres) {
	w.Conns[i] = conn
}

func (w *PostgresConfigWrapper) Remove(i int) {
	if i < 0 || i >= len(w.Conns) {
		return
	}
	w.Conns = append(w.Conns[:i], w.Conns[i+1:]...)
}

func (p Postgres) ConnType() ConnType {
	return PostgresConnType
}

func (p Postgres) ConnMeta() ConnMeta {
	return p.Meta
}

func (p Postgres) Name() string {
	return p.Meta.Name
}

func (p Postgres) Description() string {
	return p.Meta.Description
}

func (p Postgres) Tags() []string {
	return p.Meta.Tags
}

func (p Postgres) Host() string {
	return p.Hostname
}

func (p Postgres) Port() int {
	return p.PortNumber
}

func (p Postgres) Database() string {
	return p.DBName
}

func (p Postgres) Username() string {
	return p.User
}

func (p Postgres) URL() string {
	u := url.URL{
		Scheme: "postgresql",
		User:   url.UserPassword(p.User, p.Password),
		Host:   net.JoinHostPort(p.Hostname, strconv.Itoa(p.PortNumber)),
		Path:   "/" + p.DBName,
	}

	if p.SSLMode != "" {
		q := url.Values{}
		q.Set("sslmode", p.SSLMode)
		u.RawQuery = q.Encode()
	}

	return u.String()
}

func skipPGPassRow(values []string) bool {
	if len(values) != 5 {
		return true
	}

	return values[0] == "*" || values[1] == "*" || values[2] == "*" || values[3] == "*" || values[4] == "*"
}
