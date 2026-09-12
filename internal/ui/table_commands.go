package ui

import (
	"context"
	"io"

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
	return tea.Exec(
		runExec{run: func() error { return conn.Run(context.Background()) }},
		func(err error) tea.Msg { return runResultMsg{ref: ref, err: err} },
	)
}

func (m Model) addCmd() tea.Cmd {
	return changeExec(func() error {
		_, err := runAddConnection(m.conns.Active(), m.secrets, m.conns.Add)

		return err
	})
}

func (m Model) editCmd() tea.Cmd {
	cfg, ok := m.conns.Config()
	if !ok {
		return nil
	}
	return changeExec(func() error {
		_, err := runEditConnection(cfg.ConnType(), cfg, m.secrets, m.conns.Edit)

		return err
	})
}

func (m Model) infoCmd() tea.Cmd {
	var edit bool
	return tea.Exec(
		runExec{run: func() error {
			var err error
			edit, err = runInfo(m.conns)
			return err
		}},
		func(err error) tea.Msg { return infoClosedMsg{edit: edit, err: err} },
	)
}

// secretsCmd hands the terminal to the secrets screen and comes back with the
// connection list resynced: a secret that moved changes nothing on this side,
// but an entry added while over there may now be referenced from a form.
func (m Model) secretsCmd() tea.Cmd {
	if m.secrets == nil {
		return nil
	}
	return changeExec(func() error { return runSecretTable(m.secrets) })
}

func (m Model) databasesCmd() tea.Cmd {
	var (
		kind   config.ConnType
		chosen bool
	)
	return tea.Exec(
		runExec{run: func() error {
			var err error
			kind, chosen, err = runDatabaseTable()
			return err
		}},
		func(err error) tea.Msg { return databasesChangedMsg{kind: kind, chosen: chosen, err: err} },
	)
}

func (m Model) removeCmd() tea.Cmd {
	return func() tea.Msg {
		return connChangedMsg{err: m.conns.Remove(), renumbered: true}
	}
}

func changeExec(fn func() error) tea.Cmd {
	return tea.Exec(
		runExec{run: fn},
		func(err error) tea.Msg { return connChangedMsg{err: err} },
	)
}

type runExec struct {
	run func() error
}

func (r runExec) Run() error        { return r.run() }
func (runExec) SetStdin(io.Reader)  {}
func (runExec) SetStdout(io.Writer) {}
func (runExec) SetStderr(io.Writer) {}
