package ui

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/importer"
	"github.com/myjupyter/conm/internal/repository"
	"github.com/myjupyter/conm/internal/secret"
	"github.com/myjupyter/conm/internal/ui/spec"
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

	plan := planImport(repo, entries)
	results := plan.results
	if mode == ImportShow || (mode == ImportStrict && len(plan.refused) > 0) {
		return results, nil
	}

	saved := make([]int, 0, len(plan.ready))
	accepted := make([]config.Connection, 0, len(plan.ready))
	for _, i := range plan.ready {
		if err := storeImportedPassword(ctx, ws.Keyring, entries[i].Conn, entries[i].Password); err != nil {
			results[i].Status, results[i].Note = ImportFailed, passwordNote(err)
			continue
		}
		saved, accepted = append(saved, i), append(accepted, entries[i].Conn)
	}
	if mode == ImportStrict && len(saved) < len(plan.ready) {
		return results, nil
	}

	if err := repo.AddAll(accepted...); err != nil {
		return results, err
	}
	for _, i := range saved {
		results[i].Status = ImportDone
	}
	for i, twin := range plan.twins {
		if results[twin].Status == ImportDone {
			results[i].Status, results[i].Note = ImportPresent, results[twin].Name
		}
	}

	fixable := make([]int, 0, len(plan.refused))
	for _, i := range plan.refused {
		if entries[i].Err == nil {
			fixable = append(fixable, i)
		}
	}
	if mode != ImportInteractive || len(fixable) == 0 {
		return results, nil
	}

	err = WithScreens(func() error {
		for _, i := range fixable {
			fixed, err := fixEntry(ctx, ws.Keyring, repo, formCtx, entries[i])
			if err != nil {
				return err
			}
			results[i] = fixed
		}

		return nil
	})

	return results, err
}

type importPlan struct {
	results []ImportResult
	ready   []int
	refused []int
	twins   map[int]int
}

func planImport(repo repository.Connections, entries []importer.Entry) importPlan {
	plan := importPlan{results: make([]ImportResult, len(entries)), twins: map[int]int{}}

	var accepted []config.Connection
	for i, entry := range entries {
		res := &plan.results[i]
		res.Name, res.Warnings = entry.Name, entry.Warnings

		err, inFile := entry.Err, error(nil)
		if err == nil {
			res.DSN = entry.Conn.ConnectionString("")
			if err = repo.Check(entry.Conn); err == nil {
				inFile = repo.Check(entry.Conn, accepted...)
			}
		}

		dup := (*repository.DuplicateConnectionError)(nil)
		switch {
		case errors.As(err, &dup):
			res.Status, res.Note = ImportPresent, dup.Name
		case errors.As(inFile, &dup):
			twin := slices.IndexFunc(accepted, func(c config.Connection) bool { return c.Meta().Name == dup.Name })
			plan.twins[i], res.Note = plan.ready[twin], fmt.Sprintf("same connection as %q", dup.Name)
		case err != nil || inFile != nil:
			plan.refused = append(plan.refused, i)
			res.Status, res.Note = ImportFailed, oneLine(cmp.Or(err, inFile))
		default:
			plan.ready, accepted = append(plan.ready, i), append(accepted, entry.Conn)
		}
	}

	return plan
}

func fixEntry(
	ctx context.Context,
	secrets repository.Secrets,
	repo repository.Connections,
	formCtx formContext,
	entry importer.Entry,
) (ImportResult, error) {
	res := ImportResult{
		Name:     entry.Name,
		DSN:      entry.Conn.ConnectionString(""),
		Status:   ImportFailed,
		Warnings: entry.Warnings,
	}

	refusal := repo.Check(entry.Conn)
	if dup := (*repository.DuplicateConnectionError)(nil); errors.As(refusal, &dup) {
		res.Status, res.Note = ImportPresent, dup.Name
		return res, nil
	}

	var saved config.Connection
	add := func(c config.Connection) error {
		if entry.Password != "" {
			bound, password, err := keyringBound(repo.Kind(), c)
			if err != nil {
				return err
			}
			if err := repo.Check(bound); err != nil {
				return err
			}
			if err := storeImportedPassword(ctx, secrets, bound, password); err != nil {
				return err
			}
			c = bound
		}

		if err := repo.Add(c); err != nil {
			return err
		}
		saved = c

		return nil
	}

	initial, err := importSeed(repo.Kind(), entry)
	if err != nil {
		return res, err
	}

	loop, err := connectionForm(repo.Kind(), false, formCtx, add)
	if err != nil {
		return res, err
	}

	fixed, err := loop.run(initial, refusalStatus(refusal))
	if err != nil || !fixed {
		res.Note = oneLine(refusal)
		return res, err
	}

	res.Name, res.DSN, res.Status = saved.Meta().Name, saved.ConnectionString(""), ImportDone

	return res, nil
}

func importSeed(t config.ConnType, entry importer.Entry) (formValues, error) {
	initial, err := seedConnection(t, entry.Conn)
	if err != nil || entry.Password == "" {
		return initial, err
	}
	initial[spec.SecretProviderKey], initial[spec.SecretValueKey] = secret.Literal, entry.Password

	return initial, nil
}

func refusalStatus(refusal error) formStatus {
	if refusal == nil {
		return formStatus{}
	}

	return formStatus{text: oneLine(refusal), kind: kindErr}
}

func keyringBound(t config.ConnType, conn config.Connection) (config.Connection, string, error) {
	ref, ok := conn.(secret.Reference)
	if !ok {
		return conn, "", nil
	}

	scheme, password := spec.SplitSecret(ref.SecretRef())
	if scheme != secret.Literal || password == "" {
		return conn, "", nil
	}

	values, err := seedConnection(t, conn)
	if err != nil {
		return nil, "", err
	}
	values[spec.SecretProviderKey] = secret.Keyring
	values[spec.SecretValueKey] = importer.SecretLocation(t, conn.Meta().Name)

	bound, err := spec.FormSpecs[t].BuildFunc(values)

	return bound, password, err
}

func storeImportedPassword(
	ctx context.Context,
	secrets repository.Secrets,
	conn config.Connection,
	password string,
) error {
	ref, ok := conn.(secret.Reference)
	if !ok || password == "" {
		return nil
	}

	scheme, location := spec.SplitSecret(ref.SecretRef())
	if scheme != secrets.Kind() {
		return nil
	}

	i, found := secretAt(secrets, location)
	if !found {
		err := secrets.Add(ctx, config.Keyring{
			SecretID:    conn.Meta().Name,
			ProviderID:  scheme,
			LocationRef: location,
		}, password)
		if err == nil {
			return nil
		}
		if i, found = secretAt(secrets, location); !found {
			return err
		}
	}

	existing, _ := secrets.Get(i)
	if used := secrets.UsagesAt(i); len(used) > 0 {
		return &repository.SecretInUseError{ID: existing.ID(), By: used}
	}

	return secrets.Edit(ctx, i, existing, password)
}

func secretAt(secrets repository.Secrets, location string) (int, bool) {
	for i := range secrets.Len() {
		if s, ok := secrets.Get(i); ok && s.Location() == location {
			return i, true
		}
	}

	return 0, false
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
