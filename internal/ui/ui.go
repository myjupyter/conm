package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/registry"
)

func Run(cfg config.Conm) error {
	reg, err := registry.NewPostgresRegistry(cfg)
	if err != nil {
		return err
	}
	defer reg.Close()

	_, err = tea.NewProgram(New(reg)).Run()
	return err
}
