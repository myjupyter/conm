package ui

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui/view"
)

func TestModel_renderFitsWindow(t *testing.T) {
	type args struct {
		rows  int
		size  tea.WindowSizeMsg
		downs int
	}
	tests := []struct {
		name       string
		args       args
		wantWidth  int
		wantHeight int
		wantRow    string
	}{
		{
			name:       "no size yet keeps the minimum width and shows every row",
			args:       args{rows: 30},
			wantWidth:  tableInner + 2,
			wantHeight: 38,
			wantRow:    "conn-29",
		},
		{
			name:       "a wide window stretches the frame",
			args:       args{rows: 3, size: tea.WindowSizeMsg{Width: 160, Height: 40}},
			wantWidth:  160,
			wantHeight: 11,
			wantRow:    "conn-2",
		},
		{
			name:       "a narrow window keeps the minimum width",
			args:       args{rows: 3, size: tea.WindowSizeMsg{Width: 60, Height: 40}},
			wantWidth:  tableInner + 2,
			wantHeight: 11,
			wantRow:    "conn-2",
		},
		{
			name:       "a short window scrolls the rows to the cursor",
			args:       args{rows: 30, size: tea.WindowSizeMsg{Width: 120, Height: 20}, downs: 25},
			wantWidth:  120,
			wantHeight: 20,
			wantRow:    "conn-25",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confs := make([]config.Connection, tt.args.rows)
			for i := range confs {
				confs[i] = config.Postgres{Metadata: config.ConnMeta{Name: fmt.Sprintf("conn-%d", i)}}
			}
			var model tea.Model = New(view.NewConnections(stubConnections{confs: confs}), nil, nil)
			if tt.args.size.Height > 0 {
				model, _ = model.Update(tt.args.size)
			}
			for range tt.args.downs {
				model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyDown})
			}
			m, ok := model.(Model)
			require.True(t, ok)

			out := m.render()
			lines := strings.Split(out, "\n")

			assert.Len(t, lines, tt.wantHeight)
			assert.Equal(t, tt.wantWidth, lipgloss.Width(out))
			assert.Contains(t, out, tt.wantRow)
		})
	}
}

type stubMySQLConnections struct {
	stubConnections
}

func (stubMySQLConnections) Kind() config.ConnType { return config.MySQLConnType }

func TestModel_tabColorPaintsOnlyTabs(t *testing.T) {
	const (
		customType    = "2;18;52;86"
		greenAccent   = "2;99;209;138"
		postgresType  = "2;90;160;214"
		tabBackground = "2;171;205;239"
	)
	active := view.NewConnections(stubConnections{confs: []config.Connection{config.Postgres{}}})
	inactive := view.NewConnections(
		stubMySQLConnections{stubConnections{confs: []config.Connection{config.MySQL{}}}},
		stubConnections{confs: []config.Connection{config.Postgres{}}},
	)
	selectedRow := func(m Model) string { return m.rowLine(0) }
	formBadge := func(Model) string { return formModel{kind: config.PostgresConnType}.badgeLine() }
	infoTitle := func(m Model) string { return newInfoModel(m.conns).titleLine() }
	pickerRow := func(Model) string {
		return initModel{
			types:   []config.ConnType{config.MySQLConnType, config.PostgresConnType},
			clients: map[config.ConnType][]cli.Info{config.PostgresConnType: {{Installed: true}}},
		}.initRowLine(1)
	}
	tests := []struct {
		name    string
		color   string
		tab     string
		conns   *view.Connections
		line    func(Model) string
		want    string
		notWant string
	}{
		{name: "active tab", color: "#123456", conns: active, line: Model.tabsLine, want: "48;" + customType, notWant: greenAccent},
		{name: "active tab under a tab background", color: "#123456", tab: "#abcdef", conns: active, line: Model.tabsLine, want: "48;" + customType, notWant: tabBackground},
		{name: "active tab under an accent tab background", color: "#123456", tab: "accent", conns: active, line: Model.tabsLine, want: "48;" + customType, notWant: greenAccent},
		{name: "active tab on auto follows the tab background", color: "auto", tab: "#abcdef", conns: active, line: Model.tabsLine, want: "48;" + tabBackground, notWant: greenAccent},
		{name: "inactive tab", color: "#123456", conns: inactive, line: Model.tabsLine, want: "38;" + customType, notWant: "48;" + customType},
		{name: "selected row", color: "#123456", conns: active, line: selectedRow, want: "48;" + greenAccent, notWant: customType},
		{name: "form badge", color: "#123456", conns: active, line: formBadge, want: postgresType, notWant: customType},
		{name: "info title", color: "#123456", conns: active, line: infoTitle, want: postgresType, notWant: customType},
		{name: "type picker mark", color: "#123456", conns: active, line: pickerRow, want: postgresType, notWant: customType},
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

			got := tt.line(New(tt.conns, nil, nil))

			is := assert.New(t)
			is.Contains(got, tt.want)
			is.NotContains(got, tt.notWant)
		})
	}
}
