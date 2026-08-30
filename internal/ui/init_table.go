package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui/spec"
)

// initModel is the setup screen: one row per database conm knows, the clients
// found on this host, and whether the database is enabled. It runs before a
// workspace exists, so it holds a config.Conm instead of a view.
type initModel struct {
	conm    config.Conm
	dbs     []config.Database
	clients map[config.ConnType][]config.CLIInfo

	cursor int
	help   bool
	open   bool
	chosen config.ConnType
	adding bool

	status     string
	statusKind statusKind
}

type initChangedMsg struct {
	conm  config.Conm
	saved bool
	err   error
}

func newInitModel(conm config.Conm) initModel {
	m := initModel{
		conm:       conm,
		dbs:        conm.Databases,
		status:     statusReady,
		statusKind: kindIdle,
	}

	m.clients = make(map[config.ConnType][]config.CLIInfo, len(m.dbs))
	for _, db := range m.dbs {
		m.clients[db.Type] = config.DetectCLI(db.Type)
	}

	return m
}

// newAddDatabaseModel opens the same screen from the connections table, where
// it is the way to add another database: enter goes back there instead of
// opening the connections again.
func newAddDatabaseModel(conm config.Conm) initModel {
	m := newInitModel(conm)
	m.adding = true
	return m
}

func (m initModel) Init() tea.Cmd { return nil }

func (m initModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m.handleInitKey(msg.String())

	case initChangedMsg:
		return m.applyInitChange(msg), nil
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
		if m.cursor < len(m.dbs)-1 {
			m.cursor++
		}
	case keyMap.Toggle.matches(key):
		return m.toggle()
	case keyMap.Confirm.matches(key):
		return m.openConnections()
	case keyMap.Edit.matches(key):
		return m, m.editDatabaseCmd()
	case keyMap.Help.matches(key):
		m.help = !m.help
		m.setInitStatus(keyhintStatus(m.help), kindIdle)
	}

	return m, nil
}

func (m initModel) toggle() (tea.Model, tea.Cmd) {
	db, ok := m.database()
	if !ok {
		return m, nil
	}

	if !db.Enabled && db.CLI == "" {
		db.CLI = m.installedClient(db.Type)
	}
	if db.CLI == "" {
		m.setInitStatus("no client for "+db.Type.String()+" · install "+
			strings.Join(config.Clients(db.Type), " or "), kindErr)
		return m, nil
	}

	db.Enabled = !db.Enabled
	return m, m.saveDatabaseCmd(db)
}

// openConnections leaves the setup screen for the connection table of the
// database under the cursor; RunInit opens it once this program is done.
func (m initModel) openConnections() (tea.Model, tea.Cmd) {
	db, ok := m.database()
	if !ok {
		return m, nil
	}

	if db.Enabled {
		m.chosen, m.open = db.Type, true
		return m, tea.Quit
	}

	if m.adding {
		return m, tea.Quit
	}

	m.setInitStatus(db.Type.String()+" is disabled · "+keyMap.Toggle.hint+" to enable it", kindWarn)
	return m, nil
}

func (m initModel) applyInitChange(msg initChangedMsg) initModel {
	if msg.err != nil {
		m.setInitStatus(msg.err.Error(), kindErr)
		return m
	}
	if !msg.saved {
		return m
	}

	m.conm = msg.conm
	m.dbs = m.conm.Databases
	if db, ok := m.database(); ok {
		m.setInitStatus(databaseSummary(db), kindOK)
	}
	return m
}

func (m initModel) database() (config.Database, bool) {
	if m.cursor < 0 || m.cursor >= len(m.dbs) {
		return config.Database{}, false
	}
	return m.dbs[m.cursor], true
}

func (m initModel) installedClient(t config.ConnType) string {
	for _, cli := range m.clients[t] {
		if cli.Exists {
			return cli.Name
		}
	}
	return ""
}

func (m initModel) enabledCount() int {
	return enabledDatabases(m.conm)
}

func (m *initModel) setInitStatus(s string, k statusKind) {
	m.status, m.statusKind = s, k
}

func databaseSummary(db config.Database) string {
	return db.Type.String() + " " + spec.DatabaseState(db.Enabled) + " · " + db.CLI
}
