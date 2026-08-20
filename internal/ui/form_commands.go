package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/repository"
	"github.com/myjupyter/conm/internal/ui/spec"
)

func RunAddForm(cfg config.Conm, t config.ConnType) (bool, error) {
	conn, ok, err := runAddForm(t)
	if err != nil || !ok {
		return false, err
	}

	reg, err := repository.NewPostgresRepository(cfg)
	if err != nil {
		return false, err
	}
	defer reg.Close()

	if err := reg.Add(conn); err != nil {
		return false, err
	}

	return true, nil
}

func runAddForm(t config.ConnType) (config.Connection, bool, error) {
	formSpec, ok := spec.FormSpecs[t]
	if !ok {
		return nil, false, fmt.Errorf("add form is not implemented for connection type %q", t)
	}

	return runForm(newFormModel(formSpec, formSpec.AddTitle, nil))
}

func RunEditForm(t config.ConnType, existing config.Connection) (config.Connection, bool, error) {
	formSpec, ok := spec.FormSpecs[t]
	if !ok {
		return nil, false, fmt.Errorf("edit form is not implemented for connection type %q", t)
	}
	if formSpec.SeedFunc == nil {
		return nil, false, fmt.Errorf("edit form is not seedable for connection type %q", t)
	}

	return runForm(newFormModel(formSpec, formSpec.EditTitle, formSpec.SeedFunc(existing)))
}

func runForm(model formModel) (config.Connection, bool, error) {
	m, err := tea.NewProgram(model).Run()
	if err != nil {
		return nil, false, err
	}

	fm := m.(formModel)
	if !fm.submitted {
		return nil, false, nil
	}

	conn, err := fm.result()
	if err != nil {
		return nil, false, err
	}

	return conn, true, nil
}
