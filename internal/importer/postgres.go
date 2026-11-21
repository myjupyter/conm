package importer

import (
	"cmp"
	"fmt"
	"strconv"
	"strings"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/parser/params"
	"github.com/myjupyter/conm/internal/secret"
)

const (
	pgPassfileKeyword = "passfile"
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

	res, err := params.ParseArgs(config.PostgresConnType, []string{conninfo(service.Params)})
	if err != nil {
		entry.Err = err
		return entry
	}

	pg, ok := res.Conn.(config.Postgres)
	if !ok {
		entry.Err = fmt.Errorf("service was read as a %T", res.Conn)
		return entry
	}

	if pg.Password == "" {
		host := cmp.Or(pg.Hostname, pgDefaultHost)
		found := passwords.Password(host, strconv.Itoa(pg.PortNumber), pg.DBName, pg.User)
		pg.Password = secret.Ref(secret.Literal, found)
	}
	pg.Metadata.Name = service.Name

	entry.Conn, entry.Password = withoutPassword(pg, config.PostgresConnType, service.Name)
	entry.Warnings = res.Warnings

	return entry
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
