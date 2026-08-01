package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/myjupyter/conm/internal/network"
)

func Run(conns []network.Connection) error {
	_, err := tea.NewProgram(New(conns)).Run()
	return err
}
