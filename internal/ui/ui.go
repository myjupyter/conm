package ui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/repository"
	"github.com/myjupyter/conm/internal/ui/view"
)

func Run(cfg config.Conm) error {
	if enabledDatabases(cfg) == 0 {
		return RunInit()
	}

	ws, err := repository.NewWorkspace(cfg)
	if err != nil {
		return err
	}
	defer ws.Close()

	_, err = tea.NewProgram(New(view.NewConnections(ws.Connections()...), view.NewSecrets(ws.Keyring))).Run()
	return err
}

func enabledDatabases(cfg config.Conm) int {
	n := 0
	for _, db := range cfg.Databases {
		if db.Enabled {
			n++
		}
	}
	return n
}
