package ui

import (
	"context"
	"io"

	tea "charm.land/bubbletea/v2"

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
		cfg, ok, err := runAddForm(m.conns.Active(), m.secrets)
		if err != nil || !ok {
			return err
		}
		return m.conns.Add(cfg)
	})
}

func (m Model) editCmd() tea.Cmd {
	cfg, ok := m.conns.Config()
	if !ok {
		return nil
	}
	return changeExec(func() error {
		edited, ok, err := RunEditForm(cfg.ConnType(), cfg, m.secrets)
		if err != nil || !ok {
			return err
		}
		return m.conns.Edit(edited)
	})
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

// databasesCmd hands the terminal to the setup screen, the only place a
// database is enabled. A database that changes there is picked up the next
// time conm starts, since the workspace was built from the config this screen
// was opened with.
func (m Model) databasesCmd() tea.Cmd {
	return changeExec(runDatabaseTable)
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
