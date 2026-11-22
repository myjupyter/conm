package ui

import (
	"cmp"
	"context"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/exporter"
	"github.com/myjupyter/conm/internal/repository"
)

type ExportOptions struct {
	Path          string
	Show          bool
	SkipErrors    bool
	WithPasswords bool
}

func ExportConnections(
	ctx context.Context,
	conm config.Conm,
	t config.ConnType,
	names []string,
	opts ExportOptions,
) ([]ImportResult, error) {
	ws, repo, err := openConnections(conm, t)
	if err != nil {
		return nil, err
	}
	defer ws.Close()

	conns, err := pickConnections(repo, names)
	if err != nil {
		return nil, err
	}

	passwords, unresolved := make([]string, len(conns)), make([]error, len(conns))
	if opts.WithPasswords && exporter.WritesPasswords(t) {
		for i, conn := range conns {
			passwords[i], unresolved[i] = ws.Password(ctx, conn)
		}
	}

	entries, err := exporter.Plan(t, opts.Path, conns, passwords)
	if err != nil {
		return nil, err
	}

	results, refused := make([]ImportResult, len(entries)), 0
	for i := range entries {
		entries[i].Err = cmp.Or(unresolved[i], entries[i].Err)

		results[i] = ImportResult{Name: entries[i].Name, DSN: conns[i].ConnectionString(""), Warnings: entries[i].Warnings}
		if entries[i].Err != nil {
			results[i].Status, results[i].Note = ImportFailed, oneLine(entries[i].Err)
			refused++
		}
	}

	if opts.Show || (refused > 0 && !opts.SkipErrors) {
		return results, nil
	}

	if err := exporter.Write(t, opts.Path, entries); err != nil {
		return results, err
	}
	for i := range results {
		if results[i].Status != ImportFailed {
			results[i].Status = ImportDone
		}
	}

	return results, nil
}

func pickConnections(repo repository.Connections, names []string) ([]config.Connection, error) {
	if len(names) == 0 {
		conns := make([]config.Connection, 0, repo.Len())
		for i := range repo.Len() {
			if cfg, ok := repo.ConfigAt(i); ok {
				conns = append(conns, cfg)
			}
		}

		return conns, nil
	}

	conns := make([]config.Connection, 0, len(names))
	for _, name := range names {
		_, cfg, err := namedConnection(repo, name)
		if err != nil {
			return nil, err
		}
		conns = append(conns, cfg)
	}

	return conns, nil
}
