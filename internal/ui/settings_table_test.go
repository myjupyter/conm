package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
)

func tableFixture(t *testing.T, postgres config.ConnectionSettings, column string) settingsModel {
	t.Helper()
	m := newSettingsFixture(cli.Psql)
	postgres.Type = config.PostgresConnType
	m.conm.Connections[0] = postgres
	m.save = func(config.Conm) error { return nil }
	m.tab = settingsTableTab
	for i, r := range m.tableRows() {
		if r.connType == config.PostgresConnType && r.column == column {
			m.table.cursor = i
			return m
		}
	}
	require.Failf(t, "no row", "postgres %q", column)
	return m
}

func savedPostgres(t *testing.T, cmd tea.Cmd) config.ConnectionSettings {
	t.Helper()
	require.NotNil(t, cmd)
	msg, ok := cmd().(settingsSavedMsg)
	require.True(t, ok)
	require.NoError(t, msg.err)
	s, ok := msg.conm.Connection(config.PostgresConnType)
	require.True(t, ok)
	return s
}

func TestSettingsModel_tableTab(t *testing.T) {
	tests := []struct {
		name        string
		current     config.ConnectionSettings
		column      string
		act         func(settingsModel) (tea.Model, tea.Cmd)
		wantColumns []string
		wantColor   string
		wantCursor  string
	}{
		{
			name:        "hiding a column moves it below the shown ones",
			column:      "username",
			act:         func(m settingsModel) (tea.Model, tea.Cmd) { return m.toggleColumn() },
			wantColumns: []string{"name", "host", "port", "database", "ping"},
			wantColor:   "postgres",
			wantCursor:  "username",
		},
		{
			name:        "showing a column puts it last",
			current:     config.ConnectionSettings{Columns: []string{"name", "host"}},
			column:      "ping",
			act:         func(m settingsModel) (tea.Model, tea.Cmd) { return m.toggleColumn() },
			wantColumns: []string{"name", "host", "ping"},
			wantColor:   "postgres",
			wantCursor:  "ping",
		},
		{
			name:        "a column moves earlier",
			column:      "host",
			act:         func(m settingsModel) (tea.Model, tea.Cmd) { return m.stepTableRow(-1) },
			wantColumns: []string{"name", "host", "username", "port", "database", "ping"},
			wantColor:   "postgres",
			wantCursor:  "host",
		},
		{
			name:        "name moves too",
			column:      "name",
			act:         func(m settingsModel) (tea.Model, tea.Cmd) { return m.stepTableRow(1) },
			wantColumns: []string{"username", "name", "host", "port", "database", "ping"},
			wantColor:   "postgres",
			wantCursor:  "name",
		},
		{
			name:        "reset restores the default order",
			current:     config.ConnectionSettings{Columns: []string{"ping", "name"}},
			column:      "ping",
			act:         func(m settingsModel) (tea.Model, tea.Cmd) { return m.resetTableRow() },
			wantColumns: []string{"name", "username", "host", "port", "database", "ping"},
			wantColor:   "postgres",
			wantCursor:  "ping",
		},
		{
			name:        "the colour steps to the next type",
			column:      "",
			act:         func(m settingsModel) (tea.Model, tea.Cmd) { return m.stepTableRow(1) },
			wantColumns: []string{"name", "username", "host", "port", "database", "ping"},
			wantColor:   "mysql",
			wantCursor:  "",
		},
		{
			name:        "the colour steps back to the own type",
			current:     config.ConnectionSettings{TabColor: "mysql"},
			column:      "",
			act:         func(m settingsModel) (tea.Model, tea.Cmd) { return m.stepTableRow(-1) },
			wantColumns: []string{"name", "username", "host", "port", "database", "ping"},
			wantColor:   "postgres",
			wantCursor:  "",
		},
		{
			name:    "a colour is typed",
			current: config.ConnectionSettings{TabColor: "auto"},
			column:  "",
			act: func(m settingsModel) (tea.Model, tea.Cmd) {
				m.table.editing = true
				m = m.typeTableColor("#FF8800")
				return m.saveTableDraft()
			},
			wantColumns: []string{"name", "username", "host", "port", "database", "ping"},
			wantColor:   "#ff8800",
			wantCursor:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tableFixture(t, tt.current, tt.column)

			model, cmd := tt.act(m)
			saved := savedPostgres(t, cmd)
			moved, ok := model.(settingsModel)
			require.True(t, ok)
			moved.conm.SetConnection(saved)

			is := assert.New(t)
			is.Equal(tt.wantColumns, saved.VisibleColumns())
			is.Equal(tt.wantColor, saved.ResolvedTabColor())
			is.Equal(tt.wantCursor, moved.tableRow().column)
			is.Equal(config.PostgresConnType, moved.tableRow().connType)
		})
	}
}

func TestSettingsModel_tableTabRefusals(t *testing.T) {
	tests := []struct {
		name    string
		current config.ConnectionSettings
		column  string
		act     func(settingsModel) (tea.Model, tea.Cmd)
	}{
		{name: "name cannot be hidden", column: "name", act: func(m settingsModel) (tea.Model, tea.Cmd) { return m.toggleColumn() }},
		{
			name:    "a hidden column does not move",
			current: config.ConnectionSettings{Columns: []string{"name"}},
			column:  "port",
			act:     func(m settingsModel) (tea.Model, tea.Cmd) { return m.stepTableRow(1) },
		},
		{
			name:   "an unknown colour keeps the input open",
			column: "",
			act: func(m settingsModel) (tea.Model, tea.Cmd) {
				m.table.editing = true
				m = m.typeTableColor("neon")
				return m.saveTableDraft()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tableFixture(t, tt.current, tt.column)

			model, cmd := tt.act(m)
			assert.Nil(t, cmd)
			got, ok := model.(settingsModel)
			require.True(t, ok)
			assert.Equal(t, kindWarn, got.statusKind)
		})
	}
}

func TestLayoutWidthKeepsTheRowWidth(t *testing.T) {
	t.Parallel()

	full := 0
	for _, c := range config.DefaultColumns {
		full += layoutWidth(config.DefaultColumns, c, 0) + 1
	}
	tests := []struct {
		name    string
		columns []string
	}{
		{name: "defaults", columns: config.DefaultColumns},
		{name: "tags take room from host", columns: config.TableColumns},
		{name: "tags without host take room from name", columns: []string{"name", "tags", "ping"}},
		{name: "reordered", columns: []string{"ping", "host", "name", "port", "username", "database"}},
		{name: "host hidden gives name the room", columns: []string{"name", "port", "ping"}},
		{name: "ssh without database", columns: []string{"name", "username", "host", "port", "ping"}},
		{name: "name alone", columns: []string{"name"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := 0
			for _, c := range tt.columns {
				got += layoutWidth(tt.columns, c, 0) + 1
			}
			assert.Equal(t, full, got)
		})
	}
}

func TestTabColorFollowsTheSetting(t *testing.T) {
	tests := []struct {
		name  string
		color string
		tab   string
		want  func() any
	}{
		{name: "type keeps the built-in colour", want: func() any { return cPostgres }},
		{name: "auto follows the accent", color: "auto", want: func() any { return cAccent }},
		{name: "auto follows the tab background", color: "auto", tab: "pink", want: func() any { return accentColors["pink"] }},
		{name: "a custom colour", color: "#123456", want: func() any { return namedColor("#123456") }},
	}
	t.Cleanup(func() { applyConm(config.Conm{}); applyAppearance(plainAppearance) })
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conm := config.Conm{
				Connections: []config.ConnectionSettings{{Type: config.PostgresConnType, TabColor: tt.color}},
				Appearance:  plainAppearance.Clone(),
			}
			if tt.tab != "" {
				conm.Appearance.TextBackgrounds = map[string]string{"tab": tt.tab}
			}
			applyConm(conm)

			assert.Equal(t, tt.want(), tabColor(config.PostgresConnType))
		})
	}
}
