package ui

import (
	"context"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui/view"
)

func (m Model) pingCmd(ref view.ConnRef) tea.Cmd {
	conn, ok := m.conns.ConnectionFor(ref)
	if !ok {
		return nil
	}
	return func() tea.Msg {
		result, err := conn.Ping(context.Background())
		return pingResultMsg{ref: ref, result: result, err: err}
	}
}

func (m Model) runCmd() tea.Cmd {
	ref, ok := m.conns.Ref()
	if !ok {
		return nil
	}
	conn, ok := m.conns.ConnectionFor(ref)
	if !ok {
		return nil
	}
	return execCmd(
		func() error { return conn.Run(context.Background()) },
		func(err error) tea.Msg { return runResultMsg{ref: ref, err: err} },
	)
}

func (m Model) addCmd() tea.Cmd {
	return changeExec(func() error {
		_, err := runAddConnection(m.conns.Active(), m.formContext(), m.conns.Add)

		return err
	})
}

func (m Model) editCmd() tea.Cmd {
	cfg, ok := m.conns.Config()
	if !ok {
		return nil
	}
	return changeExec(func() error {
		_, err := runEditConnection(cfg.ConnType(), cfg, m.formContext(), m.conns.Edit)

		return err
	})
}

func (m Model) infoCmd() tea.Cmd {
	var edit bool
	return screenCmd(
		func() error {
			var err error
			edit, err = runInfo(m.conns)
			return err
		},
		func(err error) tea.Msg { return infoClosedMsg{edit: edit, err: err} },
	)
}

func (m Model) secretsCmd() tea.Cmd {
	if m.secrets == nil {
		return nil
	}
	return changeExec(func() error { return runSecretTable(m.secrets) })
}

func (m Model) transportsCmd() tea.Cmd {
	if m.transports == nil {
		return nil
	}
	return changeExec(func() error { return runTransportTable(m.transports, m.conns) })
}

func (m Model) workspacesCmd() tea.Cmd {
	var chosen string
	return screenCmd(
		func() error {
			var err error
			chosen, _, err = runWorkspaceTable()
			return err
		},
		func(err error) tea.Msg {
			if err != nil || chosen == config.Workspace() {
				return connChangedMsg{err: err}
			}
			return workspaceSwitchedMsg{name: chosen}
		},
	)
}

func (m Model) settingsCmd() tea.Cmd {
	var changed bool
	return screenCmd(
		func() error {
			var err error
			changed, err = runSettings()
			return err
		},
		func(err error) tea.Msg {
			if err != nil || !changed {
				return connChangedMsg{err: err}
			}
			return settingsChangedMsg{}
		},
	)
}

func (m Model) addOfTypeCmd() tea.Cmd {
	return changeExec(func() error {
		t, picked, err := runTypePicker(newAddConnectionModel())
		if err != nil || !picked {
			return err
		}

		m.conns.SetActive(t)
		_, err = runAddConnection(t, m.formContext(), m.conns.Add)

		return err
	})
}

func (m Model) removeCmd() tea.Cmd {
	return func() tea.Msg {
		return connChangedMsg{err: m.conns.Remove(), renumbered: true}
	}
}

func changeExec(fn func() error) tea.Cmd {
	return screenCmd(
		fn,
		func(err error) tea.Msg { return connChangedMsg{err: err} },
	)
}
