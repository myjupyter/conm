package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
)

// RunInit opens the setup screen. It runs before a workspace exists, so it
// talks to config directly: it writes conm.toml itself and creates the
// connection file of every database it enables. Leaving it with enter opens
// the connections of the database the cursor was on.
func RunInit() error {
	if err := config.CreateConmConfigPath(); err != nil {
		return err
	}

	conm, err := readConm()
	if err != nil {
		return err
	}

	m, err := tea.NewProgram(newInitModel(conm)).Run()
	if err != nil {
		return err
	}

	im, ok := m.(initModel)
	if !ok {
		return fmt.Errorf("setup screen returned an unexpected model %T", m)
	}
	if !im.open {
		return nil
	}

	return RunOn(im.conm, im.chosen)
}

func readConm() (config.Conm, error) {
	c, err := config.OpenConfig[*config.ConmConfigWrapper](config.ConmPath())
	if err != nil {
		return config.Conm{}, err
	}
	defer c.Close()

	return c.Get(0), nil
}

func runDatabaseTable() (config.ConnType, bool, error) {
	conm, err := readConm()
	if err != nil {
		return 0, false, err
	}

	m, err := tea.NewProgram(newAddDatabaseModel(conm)).Run()
	if err != nil {
		return 0, false, err
	}

	im, ok := m.(initModel)
	if !ok {
		return 0, false, fmt.Errorf("setup screen returned an unexpected model %T", m)
	}

	return im.chosen, im.open, nil
}

func (m initModel) saveDatabaseCmd(db config.Database) tea.Cmd {
	next := m.conm
	return func() tea.Msg {
		err := saveDatabase(&next, db)
		return initChangedMsg{conm: next, saved: err == nil, err: err}
	}
}

// editDatabaseCmd hands the terminal to the availability form and folds what
// it saved back in.
func (m initModel) editDatabaseCmd() tea.Cmd {
	db, ok := m.database()
	if !ok {
		return nil
	}
	if db.CLI == "" {
		db.CLI = m.installedClient(db.Type)
	}

	next, saved := m.conm, false
	return tea.Exec(
		runExec{run: func() error {
			edited, ok, err := runDatabaseForm(db)
			if err != nil || !ok {
				return err
			}
			saved = true
			return saveDatabase(&next, edited)
		}},
		func(err error) tea.Msg { return initChangedMsg{conm: next, saved: saved, err: err} },
	)
}

func saveDatabase(conm *config.Conm, db config.Database) error {
	if err := cli.Validate(db.Type, db.CLI); err != nil {
		return err
	}

	conm.SetDatabase(db)
	if db.Enabled {
		if err := config.CreateDatabaseConfig(db.Type); err != nil {
			return err
		}
	}

	return config.CreateConm(*conm)
}
