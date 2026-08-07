package ui

import (
	"context"
	"fmt"
	"io"

	tea "charm.land/bubbletea/v2"
	"github.com/myjupyter/conm/internal/config"
)

func (m Model) pingCmd(i int) tea.Cmd {
	conn, ok := m.reg.Get(i)
	if !ok {
		return nil
	}
	return func() tea.Msg {
		return pingResultMsg{index: i, err: conn.Ping(context.Background())}
	}
}

func (m Model) runCmd(i int) tea.Cmd {
	conn, ok := m.reg.Get(i)
	if !ok {
		return nil
	}
	return tea.Exec(
		runExec{run: func() error { return conn.Run(context.Background()) }},
		func(err error) tea.Msg { return runResultMsg{err: err} },
	)
}

func (m Model) addCmd() tea.Cmd {
	return changeExec(func() error {
		conn, ok, err := runAddForm(config.PostgresConnType)
		if err != nil || !ok {
			return err
		}
		pg, err := toPostgres(conn)
		if err != nil {
			return err
		}
		return m.reg.Add(pg)
	})
}

func (m Model) editCmd(i int) tea.Cmd {
	cfg, ok := m.reg.Config(i)
	if !ok {
		return nil
	}
	return changeExec(func() error {
		conn, ok, err := RunEditForm(config.PostgresConnType, cfg)
		if err != nil || !ok {
			return err
		}
		pg, err := toPostgres(conn)
		if err != nil {
			return err
		}
		return m.reg.Edit(i, pg)
	})
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

func toPostgres(conn config.Connection) (config.Postgres, error) {
	pg, ok := conn.(config.Postgres)
	if !ok {
		return config.Postgres{}, fmt.Errorf("unexpected connection type %T for postgres form", conn)
	}
	return pg, nil
}

type runExec struct {
	run func() error
}

func (r runExec) Run() error        { return r.run() }
func (runExec) SetStdin(io.Reader)  {}
func (runExec) SetStdout(io.Writer) {}
func (runExec) SetStderr(io.Writer) {}
