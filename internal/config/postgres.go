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

const postgresConfigTemplate = `
[[postgres]]
name = "<*>" #required
tags = [<*>] #optional
host = "staging-db.example.com" #required
port = 5432 #required
database = "" #required
username = "" #required
password = "" #required
sslmode = "require" #optional
`

type Postgres struct {
	Meta       ConnMeta `toml:"meta"`
	Hostname   string   `toml:"host" json:"host"`
	PortNumber int      `toml:"port" json:"port"`
	User       string   `toml:"username" json:"username"`
	Password   string   `toml:"password" json:"password"`
	DBName     string   `toml:"dbname" json:"dbname"`
	SSLMode    string   `toml:"sslmode,omitempty" json:"sslmode,omitempty"`
}

func (p *Postgres) ConnMeta() ConnMeta {
	return p.Meta
}

func (p *Postgres) Name() string {
	return p.Meta.Name
}

func (p *Postgres) Description() string {
	return p.Meta.Description
}

func (p *Postgres) Tags() []string {
	return p.Meta.Tags
}

func (p *Postgres) Host() string {
	return p.Hostname
}

func (p *Postgres) Port() int {
	return p.PortNumber
}

func (p *Postgres) Database() string {
	return p.DBName
}

func (p *Postgres) Username() string {
	return p.User
}

func (p *Postgres) URL() string {
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

func ImportFromPGPass() ([]Postgres, bool) {
	pgPassPath := os.Getenv("PGPASSFILE")
	if pgPassPath == "" {
		homeDir := os.Getenv("HOME")
		const pgpassFile = ".pgpass"
		pgPassPath = filepath.Join(homeDir, pgpassFile)
	}

	raw, err := os.ReadFile(pgPassPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false
		}
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

	return confs, true
}

func CreatePG(filename string, confs []Postgres) error {
	raw, err := toml.Marshal(struct {
		Postgres []Postgres `toml:"postgres"`
	}{Postgres: confs})
	if err != nil {
		return err
	}

	return os.WriteFile(filename, raw, 0600)
}

func ReadPG(filename string) ([]Postgres, error) {
	t := struct {
		Postgres []Postgres `toml:"postgres"`
	}{}

	raw, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = toml.Unmarshal(raw, &t)
	if err != nil {
		return nil, err
	}

	return t.Postgres, err
}

func skipPGPassRow(values []string) bool {
	if len(values) != 5 {
		return true
	}

	return values[0] == "*" || values[1] == "*" || values[2] == "*" || values[3] == "*" || values[4] == "*"
}
