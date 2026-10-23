package ui

import (
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
)

type initModel struct {
	types   []config.ConnType
	clients map[config.ConnType][]cli.Info

	cursor int
	help   bool
	adding bool

	chosen config.ConnType
	picked bool

	status     string
	statusKind statusKind
}

func newInitModel() initModel {
	m := initModel{
		types:      config.ConnTypes,
		clients:    make(map[config.ConnType][]cli.Info, len(config.ConnTypes)),
		status:     statusReady,
		statusKind: kindIdle,
	}

	for _, t := range m.types {
		m.clients[t] = cli.Detect(t)
	}

	if i := slices.IndexFunc(m.types, m.available); i >= 0 {
		m.cursor = i
	}

	return m
}

func newAddConnectionModel() initModel {
	m := newInitModel()
	m.adding = true

	return m
}

func (m initModel) Init() tea.Cmd { return nil }

func (m initModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		return m.handleInitKey(msg.String())
	}

	return m, nil
}

func (m initModel) handleInitKey(key string) (tea.Model, tea.Cmd) {
	switch {
	case keyMap.Quit.matches(key):
		return m, tea.Quit
	case keyMap.Up.matches(key):
		if m.cursor > 0 {
			m.cursor--
		}
	case keyMap.Down.matches(key):
		if m.cursor < len(m.types)-1 {
			m.cursor++
		}
	case keyMap.Confirm.matches(key):
		return m.pick()
	case keyMap.Help.matches(key):
		m.help = !m.help
		m.setInitStatus(keyhintStatus(m.help), kindIdle)
	}

	return m, nil
}

func (m initModel) pick() (tea.Model, tea.Cmd) {
	t, ok := m.connType()
	if !ok {
		return m, nil
	}

	if !m.available(t) {
		m.setInitStatus("no client for "+t.String()+" · install "+
			strings.Join(cli.Clients(t), " or "), kindErr)
		return m, nil
	}

	m.chosen, m.picked = t, true

	return m, tea.Quit
}

func (m initModel) connType() (config.ConnType, bool) {
	if m.cursor < 0 || m.cursor >= len(m.types) {
		return 0, false
	}

	return m.types[m.cursor], true
}

func (m initModel) available(t config.ConnType) bool {
	return slices.ContainsFunc(m.clients[t], func(c cli.Info) bool { return c.Installed })
}

func (m initModel) availableCount() int {
	n := 0
	for _, t := range m.types {
		if m.available(t) {
			n++
		}
	}

	return n
}

func (m *initModel) setInitStatus(s string, k statusKind) {
	m.status, m.statusKind = s, k
}
