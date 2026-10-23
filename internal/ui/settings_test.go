package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
)

func newSettingsFixture(current string) settingsModel {
	m := settingsModel{
		conm: config.Conm{Connections: []config.ConnectionSettings{
			{Type: config.PostgresConnType, CLI: current},
			{Type: config.MySQLConnType},
			{Type: config.MSSQLConnType},
			{Type: config.RedisConnType},
		}},
		clients: map[config.ConnType][]cli.Info{
			config.PostgresConnType: {
				{Name: cli.Psql, Installed: true},
				{Name: cli.Pgcli},
				{Name: cli.USQL, Installed: true},
			},
		},
		versions: map[string]string{},
	}
	m.window.Resize(settingsSectionWindow)

	return m
}

func TestSettingsModel_nextInstalled(t *testing.T) {
	tests := []struct {
		name    string
		current string
		step    int
		want    string
		wantOK  bool
	}{
		{name: "forward skips a missing client", current: cli.Psql, step: 1, want: cli.USQL, wantOK: true},
		{name: "forward wraps around", current: cli.USQL, step: 1, want: cli.Psql, wantOK: true},
		{name: "backward skips a missing client", current: cli.USQL, step: -1, want: cli.Psql, wantOK: true},
		{name: "unset picks the first installed", current: "", step: 1, want: cli.Psql, wantOK: true},
		{name: "unset backward picks the last installed", current: "", step: -1, want: cli.USQL, wantOK: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newSettingsFixture(tt.current)
			db, ok := m.database()
			require.True(t, ok)

			got, gotOK := m.nextInstalled(db, tt.step)
			assert.Equal(t, tt.wantOK, gotOK)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSettingsModel_nextInstalledWithoutAlternative(t *testing.T) {
	m := newSettingsFixture(cli.Psql)
	m.clients[config.PostgresConnType][2].Installed = false
	db, ok := m.database()
	require.True(t, ok)

	_, gotOK := m.nextInstalled(db, 1)
	assert.False(t, gotOK)
}

func TestSettingsModel_move(t *testing.T) {
	tests := []struct {
		name      string
		steps     []int
		wantAt    int
		wantAbove int
		wantBelow int
	}{
		{name: "opens at the top", wantAt: 0, wantAbove: 0, wantBelow: 1},
		{name: "scrolls past the window", steps: []int{1, 1, 1}, wantAt: 3, wantAbove: 1, wantBelow: 0},
		{name: "wraps to the bottom", steps: []int{-1}, wantAt: 3, wantAbove: 1, wantBelow: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newSettingsFixture(cli.Psql)
			for _, step := range tt.steps {
				m = m.move(step)
			}

			n := len(m.conm.Connections)
			assert.Equal(t, tt.wantAt, m.cursor)
			assert.Equal(t, tt.wantAbove, m.window.Above())
			assert.Equal(t, tt.wantBelow, m.window.Below(n))
		})
	}
}
