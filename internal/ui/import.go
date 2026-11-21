package ui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/importer"
	"github.com/myjupyter/conm/internal/repository"
	"github.com/myjupyter/conm/internal/secret"
)

type ImportMode int

const (
	ImportStrict ImportMode = iota
	ImportSkipErrors
	ImportInteractive
	ImportShow
)

type ImportStatus int

const (
	ImportNew ImportStatus = iota
	ImportDone
	ImportPresent
	ImportFailed
)

type ImportResult struct {
	Name     string
	DSN      string
	Status   ImportStatus
	Warnings []string
	Note     string
}

func RenderImport(w io.Writer, results []ImportResult) {
	width := 0
	for _, res := range results {
		width = max(width, lipgloss.Width(res.Name))
	}

	for _, res := range results {
		lipgloss.Fprintf(w, "[%s] %-*s  %s\n", importGlyph(res.Status), width, res.Name, res.DSN)
		if note := importNote(res); note != "" {
			fmt.Fprintf(w, "  - %s\n", note)
		}
		for _, warning := range res.Warnings {
			fmt.Fprintf(w, "  - %s\n", warning)
		}
	}
}

func importGlyph(status ImportStatus) string {
	switch status {
	case ImportDone:
		return styledGlyph(kindOK)
	case ImportPresent:
		return styledGlyph(kindWarn)
	case ImportFailed:
		return styledGlyph(kindErr)
	default:
		return lipgloss.NewStyle().Foreground(cPostgres).Render(gImportNew)
	}
}

func importNote(res ImportResult) string {
	if res.Status == ImportPresent {
		if res.Note == res.Name {
			return ""
		}

		return fmt.Sprintf("already there as %q", res.Note)
	}

	return res.Note
}

func ImportConnections(
	ctx context.Context,
	conm config.Conm,
	t config.ConnType,
	entries []importer.Entry,
	mode ImportMode,
) ([]ImportResult, error) {
	keyMap = newKeymap(conm.Keybindings)

	ws, repo, formCtx, err := openConnectionForm(conm, t)
	if err != nil {
		return nil, err
	}
	defer ws.Close()

	results, ready, refused := planImport(repo, entries)
	if mode == ImportShow {
		for i, entry := range entries {
			results[i].Warnings = entry.Warnings
		}
	}
	if mode == ImportShow || (mode == ImportStrict && len(refused) > 0) {
		return results, nil
	}

	accepted := make([]config.Connection, 0, len(ready))
	for _, i := range ready {
		accepted = append(accepted, entries[i].Conn)
	}
	if err := repo.AddAll(accepted...); err != nil {
		return results, err
	}
	for _, i := range ready {
		results[i].Status = ImportDone
		results[i].Note = passwordNote(storeImportedPassword(ctx, ws.Keyring, entries[i].Conn, entries[i].Password))
	}

	if mode != ImportInteractive {
		return results, nil
	}

	for _, i := range refused {
		if entries[i].Err != nil {
			continue
		}

		results[i], err = fixEntry(ctx, ws.Keyring, repo, formCtx, entries[i])
		if err != nil {
			return results, err
		}
	}

	return results, nil
}

func planImport(
	repo repository.Connections,
	entries []importer.Entry,
) (results []ImportResult, ready, refused []int) {
	results = make([]ImportResult, len(entries))

	var accepted []config.Connection
	for i, entry := range entries {
		results[i].Name = entry.Name

		err := entry.Err
		if err == nil {
			results[i].DSN = entry.Conn.ConnectionString("")
			err = repo.Check(entry.Conn, accepted...)
		}

		dup := (*repository.DuplicateConnectionError)(nil)
		switch {
		case err == nil:
			ready, accepted = append(ready, i), append(accepted, entry.Conn)
		case errors.As(err, &dup):
			results[i].Status, results[i].Note = ImportPresent, dup.Name
		default:
			refused = append(refused, i)
			results[i].Status, results[i].Note = ImportFailed, oneLine(err)
		}
	}

	return results, ready, refused
}

func fixEntry(
	ctx context.Context,
	secrets repository.Secrets,
	repo repository.Connections,
	formCtx formContext,
	entry importer.Entry,
) (ImportResult, error) {
	res := ImportResult{Name: entry.Name, DSN: entry.Conn.ConnectionString(""), Status: ImportFailed}

	refusal := repo.Check(entry.Conn)
	if dup := (*repository.DuplicateConnectionError)(nil); errors.As(refusal, &dup) {
		res.Status, res.Note = ImportPresent, dup.Name
		return res, nil
	}

	var saved config.Connection
	add := func(c config.Connection) error {
		if err := repo.Add(c); err != nil {
			return err
		}
		saved = c

		return nil
	}

	initial, err := seedConnection(repo.Kind(), entry.Conn)
	if err != nil {
		return res, err
	}

	loop, err := connectionForm(repo.Kind(), false, formCtx, add)
	if err != nil {
		return res, err
	}

	var fixed bool
	err = WithScreens(func() error {
		fixed, err = loop.run(initial, refusalStatus(refusal))
		return err
	})
	if err != nil || !fixed {
		res.Note = oneLine(refusal)
		return res, err
	}

	res.Name, res.DSN, res.Status = saved.Meta().Name, saved.ConnectionString(""), ImportDone
	res.Note = passwordNote(storeImportedPassword(ctx, secrets, saved, entry.Password))

	return res, nil
}

func refusalStatus(refusal error) formStatus {
	if refusal == nil {
		return formStatus{}
	}

	return formStatus{text: oneLine(refusal), kind: kindErr}
}

func storeImportedPassword(
	ctx context.Context,
	secrets repository.Secrets,
	saved config.Connection,
	password string,
) error {
	ref, ok := saved.(secret.Reference)
	if !ok || password == "" {
		return nil
	}

	scheme, location, _ := secret.ParseRef(ref.SecretRef())
	if scheme != secret.Keyring {
		return nil
	}

	return secrets.Add(ctx, config.Keyring{
		SecretID:    saved.Meta().Name,
		ProviderID:  secret.Keyring,
		LocationRef: location,
	}, password)
}

func passwordNote(err error) string {
	if err == nil {
		return ""
	}

	return "password was not stored · " + oneLine(err)
}

func oneLine(err error) string {
	if err == nil {
		return ""
	}

	return strings.ReplaceAll(err.Error(), "\n", " · ")
}
