package ui

import (
	"context"
	"io"

	tea "charm.land/bubbletea/v2"
)

func (m Model) pingCmd(i int) tea.Cmd {
	conn, ok := m.reg.ConnectionAt(i)
	if !ok {
		return nil
	}
	return func() tea.Msg {
		result, err := conn.Ping(context.Background())
		return pingResultMsg{index: i, result: result, err: err}
	}
}

func (m Model) runCmd(i int) tea.Cmd {
	conn, ok := m.reg.ConnectionAt(i)
	if !ok {
		return nil
	}
	return tea.Exec(
		runExec{run: func() error { return conn.Run(context.Background()) }},
		func(err error) tea.Msg { return runResultMsg{index: i, err: err} },
	)
}

func (m Model) addCmd() tea.Cmd {
	return changeExec(func() error {
		conn, ok, err := runAddForm(m.reg.Kind(), m.secrets)
		if err != nil || !ok {
			return err
		}
		return m.reg.Add(conn)
	})
}

func (m Model) editCmd(i int) tea.Cmd {
	// ConfigAt, not ConnectionAt: the form seeds from the stored config value,
	// and a live network.Connection satisfies config.Connection without being
	// the concrete type the spec's SeedFunc asserts on.
	cfg, ok := m.reg.ConfigAt(i)
	if !ok {
		return nil
	}
	return changeExec(func() error {
		conn, ok, err := RunEditForm(m.reg.Kind(), cfg, m.secrets)
		if err != nil || !ok {
			return err
		}
		return m.reg.Edit(i, conn)
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

func (m Model) removeCmd(i int) tea.Cmd {
	return func() tea.Msg {
		return connChangedMsg{err: m.reg.Remove(i)}
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
