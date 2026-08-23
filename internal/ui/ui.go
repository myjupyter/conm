package ui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/repository"
	"github.com/myjupyter/conm/internal/ui/view"
)

func Run(cfg config.Conm) error {
	ws, err := repository.NewWorkspace(cfg)
	if err != nil {
		return err
	}
	defer ws.Close()

	_, err = tea.NewProgram(New(view.NewConnections(ws.Postgres), view.NewSecrets(ws.Keyring))).Run()
	return err
}
