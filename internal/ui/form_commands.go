package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/repository"
	"github.com/myjupyter/conm/internal/ui/spec"
	"github.com/myjupyter/conm/internal/ui/view"
)

func RunAddForm(
	conm config.Conm,
	t config.ConnType,
	initial map[spec.FormFieldKey]spec.FormFieldValue,
	warnings []string,
) (bool, error) {
	// The workspace, not just the connection repository: the form offers the
	// secret stores as password modes, so it needs them open too.
	ws, err := repository.NewWorkspace(conm)
	if err != nil {
		return false, err
	}
	defer ws.Close()

	cfg, ok, err := runAddForm(t, initial, warnings, view.NewSecrets(ws.Keyring))
	if err != nil || !ok {
		return false, err
	}

	repo, ok := ws.ConnectionsOf(cfg.ConnType())
	if !ok {
		return false, fmt.Errorf("%q is not enabled\nrun 'conm init' to enable it", cfg.ConnType())
	}

	if err := repo.Add(cfg); err != nil {
		return false, err
	}

	return true, nil
}

func runAddForm(
	t config.ConnType,
	initial map[spec.FormFieldKey]spec.FormFieldValue,
	warnings []string,
	secrets *view.Secrets,
) (config.Connection, bool, error) {
	formSpec, ok := spec.FormSpecs[t]
	if !ok {
		return nil, false, fmt.Errorf("add form is not implemented for connection type %q", t)
	}

	model := newFormModel(t, formSpec, formSpec.AddTitle, initial, false, secrets)
	if len(warnings) > 0 {
		model.status = strings.Join(warnings, " · ")
		model.statusKind = kindWarn
	}

	return runForm(model)
}

func RunEditForm(t config.ConnType, existing config.Connection, secrets *view.Secrets) (config.Connection, bool, error) {
	formSpec, ok := spec.FormSpecs[t]
	if !ok {
		return nil, false, fmt.Errorf("edit form is not implemented for connection type %q", t)
	}
	if formSpec.SeedFunc == nil {
		return nil, false, fmt.Errorf("edit form is not seedable for connection type %q", t)
	}

	// A nil seed would make newFormModel read the form as an add — silently
	// showing "new connection" instead of the entry the user picked.
	initial := formSpec.SeedFunc(existing)
	if initial == nil {
		return nil, false, fmt.Errorf("edit form cannot seed a %T as connection type %q", existing, t)
	}

	return runForm(newFormModel(t, formSpec, formSpec.EditTitle, initial, true, secrets))
}

// pickSecretCmd hands the terminal to the keyring screen and folds the chosen
// location back into the field.
func (m formModel) pickSecretCmd() tea.Cmd {
	if m.secrets == nil {
		return nil
	}

	var picked string
	var ok bool
	current := m.vals[spec.SecretValueKey]

	return tea.Exec(
		runExec{run: func() error {
			var err error
			picked, ok, err = runSecretPicker(m.secrets, current)
			return err
		}},
		func(err error) tea.Msg { return secretPickedMsg{ref: picked, ok: ok, err: err} },
	)
}

func runForm(model formModel) (config.Connection, bool, error) {
	m, err := tea.NewProgram(model).Run()
	if err != nil {
		return nil, false, err
	}

	fm, ok := m.(formModel)
	if !ok {
		return nil, false, fmt.Errorf("connection form returned an unexpected model %T", m)
	}
	if !fm.submitted {
		return nil, false, nil
	}

	cfg, err := fm.result()
	if err != nil {
		return nil, false, err
	}

	return cfg, true, nil
}
