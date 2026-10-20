package ui

import (
	"context"
	"errors"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/i18n"
	"github.com/myjupyter/conm/internal/network"
	"github.com/myjupyter/conm/internal/secret"
	"github.com/myjupyter/conm/internal/ui/spec"
	"github.com/myjupyter/conm/internal/ui/view"
)

func (m transportModel) formContext() transportFormContext {
	conns := m.conns
	return transportFormContext{
		sources: spec.Sources{config.SSHConnType.String(): conns.NamesOf(config.SSHConnType)},
		pingVia: func(ctx context.Context, name string) (network.PingResult, error) {
			cfg, ok := conns.Named(config.SSHConnType, name)
			if !ok {
				return network.PingResult{}, errors.New(i18n.Sprintf("no ssh connection named %q", name))
			}

			conn, err := network.NewConnection(cli.Launcher{}, cfg, secret.Default())
			if err != nil {
				return network.PingResult{}, err
			}
			defer conn.Close()

			return conn.Ping(ctx)
		},
	}
}

func (m transportModel) addTransportCmd() tea.Cmd {
	return transportChangeExec(func() error {
		_, err := runAddTransport(m.formContext(), m.transports.Add)
		return err
	})
}

func (m transportModel) editTransportCmd() tea.Cmd {
	existing, ok := m.transports.Transport()
	if !ok {
		return nil
	}
	return transportChangeExec(func() error {
		_, err := runEditTransport(existing, m.formContext(), m.transports.Edit)
		return err
	})
}

func (m transportModel) removeTransportCmd() tea.Cmd {
	return func() tea.Msg {
		return transportChangedMsg{err: m.transports.Remove()}
	}
}

func transportChangeExec(fn func() error) tea.Cmd {
	return tea.Exec(
		runExec{run: fn},
		func(err error) tea.Msg { return transportChangedMsg{err: err} },
	)
}

func runTransportTable(transports *view.Transports, conns *view.Connections) error {
	_, err := tea.NewProgram(newTransportModel(transports, conns)).Run()
	return err
}
