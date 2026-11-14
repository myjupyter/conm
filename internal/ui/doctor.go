package ui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/repository"
)

var ErrDoctorFound = errors.New("doctor found problems")

type doctorCheck struct {
	kind     statusKind
	title    string
	problems []string
}

func Doctor(ctx context.Context, w io.Writer, version string, conm config.Conm, conmErr error) error {
	sections := []struct {
		title  string
		checks []doctorCheck
	}{
		{fmt.Sprintf("Doctor summary (conm %s)", version), []doctorCheck{conmCheck(conm, conmErr)}},
		{"\nClients:", clientChecks(ctx, conm)},
		{"\nConnections:", fileChecks(conm)},
	}

	issues, failed := 0, false
	for _, section := range sections {
		fmt.Fprintln(w, section.title)
		for _, c := range section.checks {
			renderDoctorCheck(w, c)
			if c.kind != kindOK {
				issues++
			}
			failed = failed || c.kind == kindErr
		}
	}

	renderDoctorSummary(w, issues)
	if failed {
		return ErrDoctorFound
	}
	return nil
}

func clientChecks(ctx context.Context, conm config.Conm) []doctorCheck {
	clients := repository.CheckClients(ctx, conm)
	checks := make([]doctorCheck, 0, len(clients))
	for _, c := range clients {
		checks = append(checks, clientCheck(c))
	}

	return checks
}

func fileChecks(conm config.Conm) []doctorCheck {
	files := repository.CheckFiles(conm)
	checks := make([]doctorCheck, 0, len(files))
	for _, c := range files {
		checks = append(checks, fileCheck(c))
	}

	return checks
}

func conmCheck(conm config.Conm, conmErr error) doctorCheck {
	check := doctorCheck{kind: kindOK, title: "Conm internal configs"}
	if conmErr == nil {
		conmErr = conm.Keybindings.Validate()
	}
	if conmErr != nil {
		check.kind = kindErr
		check.problems = errorLines(conmErr)
	}

	return check
}

func clientCheck(c repository.ClientCheck) doctorCheck {
	entity := capitalized(c.Type.String())
	client := strings.TrimSpace(c.Client.Name + " " + c.Client.Version)

	switch {
	case c.Invalid != nil:
		return doctorCheck{kindErr, fmt.Sprintf("%s (%s)", entity, client), errorLines(c.Invalid)}
	case c.Client.Name == "":
		return doctorCheck{kindWarn, entity + " (no client installed)", []string{"no client installed"}}
	case !c.Client.Installed:
		return doctorCheck{kindErr, fmt.Sprintf("%s (%s)", entity, client), []string{c.Client.Name + " is not installed"}}
	default:
		return doctorCheck{kindOK, fmt.Sprintf("%s (%s)", entity, client), nil}
	}
}

func fileCheck(c repository.FileCheck) doctorCheck {
	check := doctorCheck{kind: kindOK, title: capitalized(c.Kind)}
	if c.Err != nil {
		check.problems = errorLines(c.Err)
	}
	for _, rec := range c.Records {
		for _, err := range rec.Errs {
			check.problems = append(check.problems, fmt.Sprintf("%q: %v", rec.Name, err))
		}
	}
	if len(check.problems) > 0 {
		check.kind = kindErr
	}

	return check
}

func renderDoctorCheck(w io.Writer, c doctorCheck) {
	lipgloss.Fprintf(w, "[%s] %s\n", styledGlyph(c.kind), c.title)
	for _, p := range c.problems {
		fmt.Fprintf(w, "  - %s\n", p)
	}
}

func renderDoctorSummary(w io.Writer, issues int) {
	switch issues {
	case 0:
		lipgloss.Fprintf(w, "\n%s No issues found.\n", styledGlyph(kindOK))
	case 1:
		lipgloss.Fprintf(w, "\n%s Doctor found issues in 1 category.\n", styledGlyph(kindWarn))
	default:
		lipgloss.Fprintf(w, "\n%s Doctor found issues in %d categories.\n", styledGlyph(kindWarn), issues)
	}
}

func styledGlyph(k statusKind) string {
	glyph, col := statusGlyph(k)
	return lipgloss.NewStyle().Foreground(col).Render(glyph)
}

func errorLines(err error) []string {
	return strings.Split(err.Error(), "\n")
}

func capitalized(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
