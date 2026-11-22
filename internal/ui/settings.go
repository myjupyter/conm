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

const (
	settingsClientTab = iota
	settingsKeybindTab
	settingsAccentTab
	settingsTextTab
	settingsGlyphsTab
	settingsBordersTab
	settingsTableTab
	settingsPages
)

var (
	settingsTabTitles   = []string{"client", "keybind", "appearance"}
	appearanceTabTitles = []string{"accent", "text", "glyphs", "borders", "table"}
)

type settingsModel struct {
	size screenSize

	conm     config.Conm
	clients  map[config.ConnType][]cli.Info
	versions map[string]string
	save     func(config.Conm) error

	tab        int
	nested     bool
	appearPage int
	cursor     int
	window     view.Scroll
	changed    bool

	binds  bindsTab
	appear appearanceTab
	table  tableTab

	status     string
	statusKind statusKind
	help       bool
}

type clientVersionMsg struct {
	name    string
	version string
}

type settingsSavedMsg struct {
	conm   config.Conm
	status string
	err    error
}

func newSettingsModel(conm config.Conm) settingsModel {
	m := settingsModel{
		conm:     conm,
		clients:  make(map[config.ConnType][]cli.Info, len(conm.Connections)),
		versions: make(map[string]string),
		save:     config.CreateConm,
		status:   statusReady,
	}
	m.window.Resize(settingsSectionWindow)
	m.binds.window.Resize(bindsWindow)
	m.appear.window.Resize(appearanceWindow)
	m.table.window.Resize(tableSettingsWindow)

	for _, db := range conm.Connections {
		m.clients[db.Type] = cli.Detect(db.Type)
	}

	return m
}

func (m settingsModel) Init() tea.Cmd { return m.probeVersionsCmd() }

func (m settingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.size.observe(msg)
	return followed[settingsModel](m.update(msg))
}

func (m settingsModel) followCursor() settingsModel {
	chrome := m.chrome()
	switch m.tab {
	case settingsKeybindTab:
		m.size.fit(&m.binds.window, chrome, m.binds.cursor, len(m.bindings()), func() int { return len(m.bindsLines()) })
	case settingsTableTab:
		m.size.fit(&m.table.window, chrome, m.table.cursor, len(m.tableRows()), func() int { return len(m.tableLines()) })
	case settingsClientTab:
		m.size.fit(&m.window, chrome, m.cursor, len(m.conm.Connections), func() int { return len(m.clientLines()) })
	default:
		m.size.fit(&m.appear.window, chrome, m.appear.cursor, len(m.pageRows()), func() int { return len(m.appearanceLines()) })
	}
	return m
}

func (m settingsModel) update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.appear.editing {
			return m.handleAppearanceInput(msg)
		}
		if m.table.editing {
			return m.handleTableInput(msg)
		}
		return m.handleSettingsKey(msg.String())

	case tea.PasteMsg:
		if m.appear.editing {
			m = m.typeAppearance(msg.Content)
		}
		if m.table.editing {
			m = m.typeTableColor(msg.Content)
		}

	case bindCaptureExpiredMsg:
		return m.captureExpired(msg)

	case clientVersionMsg:
		m.versions[msg.name] = msg.version

	case settingsSavedMsg:
		if msg.err != nil {
			m.setSettingsStatus(msg.err.Error(), kindErr)
			break
		}
		m.conm, m.changed = msg.conm, true
		applyConm(m.conm)
		m.setSettingsStatus(msg.status, kindOK)
	}

	return m, nil
}

func (m settingsModel) handleSettingsKey(key string) (tea.Model, tea.Cmd) {
	var (
		model   tea.Model
		cmd     tea.Cmd
		handled bool
	)
	if next, ok := m.toggleNested(key); ok {
		return next, nil
	}
	switch {
	case m.tab == settingsKeybindTab:
		model, cmd, handled = m.handleBindsKey(key)
	case m.topTab() == settingsAccentTab && !m.nested:
	case m.onAppearancePage():
		model, cmd, handled = m.handleAppearanceKey(key)
	case m.tab == settingsTableTab:
		model, cmd, handled = m.handleTableKey(key)
	}
	if handled {
		return model, cmd
	}

	switch {
	case keyMap.Help.matches(key):
		m.help = !m.help
		m.setSettingsStatus(keyhintStatus(m.help), kindIdle)
	case keyMap.Quit.matches(key):
		return m, tea.Quit
	case keyMap.NextSection.matches(key):
		m = m.switchTab(1)
	case keyMap.PrevSection.matches(key):
		m = m.switchTab(-1)
	case m.tab != settingsClientTab:
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

func (m settingsModel) toggleNested(key string) (settingsModel, bool) {
	if m.topTab() != settingsAccentTab {
		return m, false
	}
	entering := !m.nested && keyMap.Confirm.matches(key)
	leaving := m.nested && keyMap.Cancel.matches(key)
	if !entering && !leaving {
		return m, false
	}
	m.nested = entering
	m.setSettingsStatus(statusReady, kindIdle)
	return m, true
}

func (m settingsModel) switchTab(step int) settingsModel {
	if m.nested {
		n := settingsPages - settingsAccentTab
		m.tab = settingsAccentTab + (m.appearanceTab()+step+n)%n
		m.appear.cursor = 0
		m.appear.window.Top()
	} else {
		n := len(settingsTabTitles)
		top := (m.topTab() + step + n) % n
		if m.topTab() == settingsAccentTab {
			m.appearPage = m.tab
		}
		m.tab = top
		if top == settingsAccentTab {
			m.tab = max(m.appearPage, settingsAccentTab)
		}
	}
	m.setSettingsStatus(statusReady, kindIdle)
	return m
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

func (m settingsModel) topTab() int { return min(m.tab, settingsAccentTab) }

func (m settingsModel) appearanceTab() int { return m.tab - settingsAccentTab }

func (m settingsModel) onAppearancePage() bool {
	return m.tab >= settingsAccentTab && m.tab < settingsTableTab
}
