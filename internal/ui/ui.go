package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/repository"
	"github.com/myjupyter/conm/internal/ui/view"
)

func Run(cfg config.Conm) error {
	var anyConnection config.ConnType

	return RunOn(cfg, anyConnection)
}

func RunOn(cfg config.Conm, active config.ConnType) error {
	for {
		if enabledConnections(cfg) == 0 {
			return RunInit()
		}

		m, err := runConnections(cfg, active)
		if err != nil || !m.reload {
			return err
		}

		cfg, err = readConm()
		if err != nil {
			return err
		}
		active = m.active
	}
}

func runConnections(cfg config.Conm, active config.ConnType) (Model, error) {
	ws, err := repository.NewWorkspace(cfg)
	if err != nil {
		return Model{}, err
	}
	defer ws.Close()

	conns := view.NewConnections(ws.Connections()...)
	conns.SetActive(active)

	m, err := tea.NewProgram(New(conns, view.NewSecrets(ws.Keyring))).Run()
	if err != nil {
		return Model{}, err
	}

	cm, ok := m.(Model)
	if !ok {
		return Model{}, fmt.Errorf("connections screen returned an unexpected model %T", m)
	}

	return cm, nil
}

func enabledConnections(cfg config.Conm) int {
	n := 0
	for _, conn := range cfg.Connections {
		if conn.Enabled {
			n++
		}
	}
	return n
}
