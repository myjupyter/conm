package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
)

func RunAdd(conm config.Conm) error {
	t, picked, err := runTypePicker(newAddConnectionModel())
	if err != nil || !picked {
		return err
	}

	saved, err := RunAddForm(conm, t, nil, nil)
	if err != nil || !saved {
		return err
	}

	return RunOn(conm, t)
}

func readConm() (config.Conm, error) {
	c, err := config.OpenConfig[*config.ConmConfigWrapper](config.ConmPath())
	if err != nil {
		return config.Conm{}, err
	}

	return c.Get(0), nil
}

func runTypePicker(start initModel) (config.ConnType, bool, error) {
	m, err := tea.NewProgram(start).Run()
	if err != nil {
		return 0, false, err
	}

	im, ok := m.(initModel)
	if !ok {
		return 0, false, fmt.Errorf("connection type picker returned an unexpected model %T", m)
	}

	return im.chosen, im.picked, nil
}
