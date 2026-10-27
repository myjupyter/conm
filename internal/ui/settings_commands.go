package ui

import (
	"context"
	"fmt"
	"slices"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
)

func (m settingsModel) probeVersionsCmd() tea.Cmd {
	var cmds []tea.Cmd
	probed := make(map[string]bool)
	for _, infos := range m.clients {
		for _, info := range infos {
			if !info.Installed || probed[info.Name] {
				continue
			}
			probed[info.Name] = true
			cmds = append(cmds, func() tea.Msg {
				return clientVersionMsg{name: info.Name, version: cli.Version(context.Background(), info.Name)}
			})
		}
	}

	return tea.Batch(cmds...)
}

func (m settingsModel) saveClientCmd(db config.ConnectionSettings) tea.Cmd {
	next, save := m.conm, m.save
	next.Connections = slices.Clone(m.conm.Connections)

	return func() tea.Msg {
		if err := cli.Validate(db.Type, db.CLI); err != nil {
			return settingsSavedMsg{err: err}
		}

		next.SetConnection(db)

		return settingsSavedMsg{conm: next, status: db.Type.String() + " opens with " + db.CLI, err: save(next)}
	}
}

func (m settingsModel) saveKeysCmd(path string, keys []string, status string) tea.Cmd {
	next, save := m.conm, m.save
	next.Keybindings.Set(path, keys)

	return func() tea.Msg {
		return settingsSavedMsg{conm: next, status: status, err: save(next)}
	}
}

func runSettings() (bool, error) {
	conm, err := config.ReadConm(config.ConmPath())
	if err != nil {
		return false, err
	}

	m, err := tea.NewProgram(newSettingsModel(conm)).Run()
	if err != nil {
		return false, err
	}

	sm, ok := m.(settingsModel)
	if !ok {
		return false, fmt.Errorf("settings screen returned an unexpected model %T", m)
	}

	return sm.changed, nil
}
