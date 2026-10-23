package ui

import (
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui/view"
)

const settingsSectionWindow = 3

var settingsTabTitles = []string{"client"}

type settingsModel struct {
	conm     config.Conm
	clients  map[config.ConnType][]cli.Info
	versions map[string]string

	cursor  int
	window  view.Scroll
	changed bool

	status     string
	statusKind statusKind
	help       bool
}

type clientVersionMsg struct {
	name    string
	version string
}

type settingsSavedMsg struct {
	conm config.Conm
	err  error
}

func newSettingsModel(conm config.Conm) settingsModel {
	m := settingsModel{
		conm:     conm,
		clients:  make(map[config.ConnType][]cli.Info, len(conm.Connections)),
		versions: make(map[string]string),
		status:   statusReady,
	}
	m.window.Resize(settingsSectionWindow)

	for _, db := range conm.Connections {
		m.clients[db.Type] = cli.Detect(db.Type)
	}

	return m
}

func (m settingsModel) Init() tea.Cmd { return m.probeVersionsCmd() }

func (m settingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m.handleSettingsKey(msg.String())

	case clientVersionMsg:
		m.versions[msg.name] = msg.version

	case settingsSavedMsg:
		if msg.err != nil {
			m.setSettingsStatus(msg.err.Error(), kindErr)
			break
		}
		m.conm, m.changed = msg.conm, true
		if db, ok := m.database(); ok {
			m.setSettingsStatus(db.Type.String()+" opens with "+db.CLI, kindOK)
		}
	}

	return m, nil
}

func (m settingsModel) handleSettingsKey(key string) (tea.Model, tea.Cmd) {
	switch {
	case keyMap.Help.matches(key):
		m.help = !m.help
		m.setSettingsStatus(keyhintStatus(m.help), kindIdle)
	case keyMap.Quit.matches(key):
		return m, tea.Quit
	case keyMap.Down.matches(key):
		m = m.move(1)
	case keyMap.Up.matches(key):
		m = m.move(-1)
	case keyMap.Right.matches(key):
		return m.cycleClient(1)
	case keyMap.Left.matches(key):
		return m.cycleClient(-1)
	}

	return m, nil
}

func (m settingsModel) move(step int) settingsModel {
	n := len(m.conm.Connections)
	if n == 0 {
		return m
	}

	m.cursor = (m.cursor + step + n) % n
	m.window.Follow(m.cursor, n)
	m.setSettingsStatus(statusReady, kindIdle)

	return m
}

func (m settingsModel) cycleClient(step int) (tea.Model, tea.Cmd) {
	db, ok := m.database()
	if !ok {
		return m, nil
	}

	next, ok := m.nextInstalled(db, step)
	if !ok {
		m.setSettingsStatus("no other client for "+db.Type.String()+" · install "+
			strings.Join(cli.Clients(db.Type), " or "), kindWarn)
		return m, nil
	}

	db.CLI = next

	return m, m.saveClientCmd(db)
}

func (m settingsModel) nextInstalled(db config.ConnectionSettings, step int) (string, bool) {
	clients := m.clients[db.Type]
	n := len(clients)
	current := m.current(db)
	at := max(slices.IndexFunc(clients, func(c cli.Info) bool { return c.Name == current }), 0)

	for i := 1; i <= n; i++ {
		c := clients[((at+step*i)%n+n)%n]
		if c.Installed && c.Name != current {
			return c.Name, true
		}
	}

	return "", false
}

func (m settingsModel) database() (config.ConnectionSettings, bool) {
	if m.cursor < 0 || m.cursor >= len(m.conm.Connections) {
		return config.ConnectionSettings{}, false
	}

	return m.conm.Connections[m.cursor], true
}

func (m settingsModel) current(db config.ConnectionSettings) string {
	if db.CLI != "" {
		return db.CLI
	}

	i := slices.IndexFunc(m.clients[db.Type], func(c cli.Info) bool { return c.Installed })
	if i < 0 {
		return ""
	}

	return m.clients[db.Type][i].Name
}

func (m settingsModel) installed(t config.ConnType) bool {
	return slices.ContainsFunc(m.clients[t], func(c cli.Info) bool { return c.Installed })
}

func (m settingsModel) clientInfo(db config.ConnectionSettings) (cli.Info, bool) {
	current := m.current(db)
	i := slices.IndexFunc(m.clients[db.Type], func(c cli.Info) bool { return c.Name == current })
	if i < 0 {
		return cli.Info{}, false
	}

	return m.clients[db.Type][i], true
}

func (m *settingsModel) setSettingsStatus(s string, k statusKind) {
	m.status, m.statusKind = s, k
}
