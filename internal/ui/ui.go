package ui

import (
	"context"

	tea "charm.land/bubbletea/v2"
)

type Connection interface {
	Name() string
	Description() string
	Tags() []string

	Username() string
	Host() string
	Database() string
	Port() int

	Ping(ctx context.Context) error
	Run(ctx context.Context) error
	Close() error
}

// Run launches the interactive connection list UI.
func Run(conns []Connection) error {
	_, err := tea.NewProgram(New(conns)).Run()
	return err
}
