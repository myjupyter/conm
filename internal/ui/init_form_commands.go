package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui/spec"
)

// runDatabaseForm opens the availability form for one database and reports
// what the user saved, if anything.
func runDatabaseForm(db config.Database) (config.Database, bool, error) {
	formSpec, ok := spec.DatabaseFormSpecs[db.Type]
	if !ok {
		return config.Database{}, false, fmt.Errorf("edit form is not implemented for database %q", db.Type)
	}

	initial := formSpec.SeedFunc(db)
	if initial == nil {
		return config.Database{}, false, fmt.Errorf("edit form cannot seed a %q database", db.Type)
	}

	m, err := tea.NewProgram(newInitFormModel(formSpec, db.Type, initial)).Run()
	if err != nil {
		return config.Database{}, false, err
	}

	fm, ok := m.(initFormModel)
	if !ok {
		return config.Database{}, false, fmt.Errorf("database form returned an unexpected model %T", m)
	}
	if !fm.submitted {
		return config.Database{}, false, nil
	}

	edited, err := fm.result()
	if err != nil {
		return config.Database{}, false, err
	}

	return edited, true, nil
}
