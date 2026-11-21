package importer

import (
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/adrg/xdg"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/parser/params"
)

const (
	pgPassfileKeyword = "passfile"
	pgPasswordKeyword = "password"
	pgDefaultHost     = "localhost"
)

var conninfoEscaper = strings.NewReplacer(`\`, `\\`, `'`, `\'`)

func readPostgres(path string) ([]Entry, error) {
	services, err := ReadINI(path)
	if err != nil {
		return nil, err
	}

	shared, _ := ReadPGPass(PGPassPath())

	entries := make([]Entry, 0, len(services))
	for _, service := range services {
		passwords := shared
		if own := values(service.Params)[pgPassfileKeyword]; own != "" {
			passwords, _ = ReadPGPass(own)
		}
		entries = append(entries, postgresEntry(service, passwords))
	}

	return entries, nil
}

func postgresEntry(service Section, passwords PGPass) Entry {
	entry := Entry{Name: service.Name}

	pg, warnings, err := postgresService(service.Params)
	if err == nil && pg.Password == "" {
		host := cmp.Or(pg.Hostname, pgDefaultHost)
		if found := passwords.Password(host, strconv.Itoa(pg.PortNumber), pg.DBName, pg.User); found != "" {
			fromFile := Param{Key: pgPasswordKeyword, Value: found}
			pg, warnings, err = postgresService(slices.Insert(slices.Clone(service.Params), 0, fromFile))
		}
	}
	if err != nil {
		entry.Err = err
		return entry
	}
	pg.Metadata.Name = service.Name

	entry.Conn, entry.Warnings = pg, warnings

	return entry
}

func postgresService(pairs []Param) (config.Postgres, []string, error) {
	res, err := params.ParseOptions(config.PostgresConnType, []string{conninfo(pairs)})
	if err != nil {
		return config.Postgres{}, nil, err
	}

	pg, ok := res.Conn.(config.Postgres)
	if !ok {
		return config.Postgres{}, nil, fmt.Errorf("service was read as a %T", res.Conn)
	}

	return pg, res.Warnings, nil
}

func conninfo(pairs []Param) string {
	keywords := make([]string, 0, len(pairs))
	seen := make(map[string]bool, len(pairs))
	for _, pair := range pairs {
		if seen[pair.Key] {
			continue
		}
		seen[pair.Key] = true
		keywords = append(keywords, pair.Key+"='"+conninfoEscaper.Replace(pair.Value)+"'")
	}

	return strings.Join(keywords, " ")
}

const (
	pgServiceFileName = ".pg_service.conf"
	pgPassFileName    = ".pgpass"
	pgPassFields      = 5
	pgPassWildcard    = "*"
)

func PGServicePath() string {
	if path := os.Getenv("PGSERVICEFILE"); path != "" {
		return path
	}

	return filepath.Join(xdg.Home, pgServiceFileName)
}

func PGPassPath() string {
	if path := os.Getenv("PGPASSFILE"); path != "" {
		return path
	}

	return filepath.Join(xdg.Home, pgPassFileName)
}

type PGPass [][]string

func ReadPGPass(path string) (PGPass, error) {
	raw, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var rows PGPass
	for line := range strings.Lines(string(raw)) {
		line = strings.TrimRight(line, "\r\n")
		if strings.HasPrefix(line, "#") {
			continue
		}
		if fields := splitPGPassRow(line); len(fields) == pgPassFields {
			rows = append(rows, fields)
		}
	}

	return rows, nil
}

func (p PGPass) Password(host, port, database, user string) string {
	want := []string{host, port, database, user}

rows:
	for _, row := range p {
		for i, value := range want {
			if row[i] != pgPassWildcard && row[i] != value {
				continue rows
			}
		}

		return row[pgPassFields-1]
	}

	return ""
}

func splitPGPassRow(line string) []string {
	var (
		fields  []string
		field   strings.Builder
		escaped bool
	)

	for _, r := range line {
		switch {
		case escaped:
			field.WriteRune(r)
			escaped = false
		case r == '\\':
			escaped = true
		case r == ':' && len(fields) < pgPassFields-1:
			fields = append(fields, field.String())
			field.Reset()
		default:
			field.WriteRune(r)
		}
	}

	return append(fields, field.String())
}
