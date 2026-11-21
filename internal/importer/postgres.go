package importer

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
	"strings"

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
	services, err := config.ReadINI(path)
	if err != nil {
		return nil, err
	}

	shared, _ := config.ReadPGPass(config.PGPassPath())

	entries := make([]Entry, 0, len(services))
	for _, service := range services {
		passwords := shared
		if own := values(service.Params)[pgPassfileKeyword]; own != "" {
			passwords, _ = config.ReadPGPass(own)
		}
		entries = append(entries, postgresEntry(service, passwords))
	}

	return entries, nil
}

func postgresEntry(service config.ForeignSection, passwords config.PGPass) Entry {
	entry := Entry{Name: service.Name}

	pg, warnings, err := postgresService(service.Params)
	if err == nil && pg.Password == "" {
		host := cmp.Or(pg.Hostname, pgDefaultHost)
		if found := passwords.Password(host, strconv.Itoa(pg.PortNumber), pg.DBName, pg.User); found != "" {
			fromFile := config.ForeignParam{Key: pgPasswordKeyword, Value: found}
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

func postgresService(pairs []config.ForeignParam) (config.Postgres, []string, error) {
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

func conninfo(pairs []config.ForeignParam) string {
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
