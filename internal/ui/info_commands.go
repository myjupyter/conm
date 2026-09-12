package ui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/ui/view"
)

func (m infoModel) clientInfoCmd() tea.Cmd {
	return func() tea.Msg {
		info, _ := m.conns.ClientInfo(context.Background())

		return clientInfoMsg(info)
	}
}

func (m infoModel) connectCmd() tea.Cmd {
	conn, ok := m.conns.Connection()
	if !ok {
		return nil
	}

	return tea.Exec(
		runExec{run: func() error { return conn.Run(context.Background()) }},
		func(err error) tea.Msg { return infoRanMsg{err: err} },
	)
}

func runInfo(conns *view.Connections) (bool, error) {
	m, err := tea.NewProgram(newInfoModel(conns)).Run()
	if err != nil {
		return false, err
	}

	final, ok := m.(infoModel)
	if !ok {
		return false, fmt.Errorf("info screen returned an unexpected model %T", m)
	}

	return final.edit, nil
}
