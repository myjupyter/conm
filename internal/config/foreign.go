package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
)

const (
	pgServiceFileName   = ".pg_service.conf"
	pgPassFileName      = ".pgpass"
	mysqlOptionFileName = ".my.cnf"
	pgPassFields        = 5
	pgPassWildcard      = "*"
)

func PGServicePath() string {
	if path := os.Getenv("PGSERVICEFILE"); path != "" {
		return path
	}

	return filepath.Join(xdg.Home, pgServiceFileName)
}

func MySQLOptionPath() string {
	return filepath.Join(xdg.Home, mysqlOptionFileName)
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
		case r == ':':
			fields = append(fields, field.String())
			field.Reset()
		default:
			field.WriteRune(r)
		}
	}

	return append(fields, field.String())
}
