package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/repository"
)

func Run(cfg config.Conm) error {
	ws, err := repository.NewWorkspace(cfg)
	if err != nil {
		return err
	}
	defer ws.Close()

	_, err = tea.NewProgram(New(ws.Postgres)).Run()
	return err
}
