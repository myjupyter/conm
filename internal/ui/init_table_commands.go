package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
)

func RunInit() error {
	if err := config.CreateConmConfigPath(); err != nil {
		return err
	}

	conm, err := readConm()
	if err != nil {
		return err
	}
	keyMap = newKeymap(conm.Keybindings)

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
	return config.ReadConm(config.ConmPath())
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
		return 0, false, fmt.Errorf("connection type picker returned an unexpected model %T", m)
	}

	return im.chosen, im.picked, nil
}
