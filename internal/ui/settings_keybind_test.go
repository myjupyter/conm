package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
)

const (
	downPath  = "navigation.down"
	retryPath = "connection.retry"
)

func newBindsFixture(t *testing.T, path string, prepare func(*config.KeybindingSettings)) settingsModel {
	t.Helper()

	m := newSettingsFixture(cli.Psql)
	m.conm.Keybindings = config.DefaultKeybindings()
	if prepare != nil {
		prepare(&m.conm.Keybindings)
	}
	m.save = func(config.Conm) error { return nil }
	m.binds.window.Resize(bindsWindow)
	m.tab = settingsKeybindTab
	for m.binding().Path != path {
		m = m.moveBind(1)
		require.NotZero(t, m.binds.cursor, "no binding at %s", path)
	}

	return m
}

func pressKeys(t *testing.T, m settingsModel, keys ...string) (settingsModel, *settingsSavedMsg) {
	t.Helper()

	var saved *settingsSavedMsg
	apply := func(next tea.Model, cmd tea.Cmd) {
		got, ok := next.(settingsModel)
		require.True(t, ok)
		m = got
		if cmd == nil || m.binds.first != "" {
			return
		}
		msg, ok := cmd().(settingsSavedMsg)
		require.True(t, ok)
		saved = &msg
	}

	for _, key := range keys {
		apply(m.handleSettingsKey(key))
	}
	if m.binds.first != "" {
		apply(m.Update(bindCaptureExpiredMsg{id: m.binds.captureID}))
	}
	return m, saved
}

func keysOf(c config.Conm, path string) []string {
	for _, b := range c.Keybindings.Bindings() {
		if b.Path == path {
			return b.Keys
		}
	}
	return nil
}

func TestSettingsModel_handleBindsKey(t *testing.T) {
	type args struct {
		path    string
		prepare func(*config.KeybindingSettings)
		keys    []string
	}
	tests := []struct {
		name        string
		args        args
		want        []string
		wantErr     error
		wantCapture bool
		wantConfirm bool
	}{
		{
			name: "a then a key appends it",
			args: args{path: retryPath, keys: []string{"a", "ctrl+n"}},
			want: []string{"r", "R", "ctrl+n"},
		},
		{
			name: "e replaces the focused key",
			args: args{path: retryPath, keys: []string{"l", "e", "n"}},
			want: []string{"r", "n"},
		},
		{
			name:        "a waits for the key to bind",
			args:        args{path: retryPath, keys: []string{"a"}},
			wantCapture: true,
		},
		{
			name: "esc cancels the capture",
			args: args{path: retryPath, keys: []string{"a", "esc"}},
		},
		{
			name: "a key already on the action is refused",
			args: args{path: retryPath, keys: []string{"a", "R"}},
		},
		{
			name: "d removes the focused key",
			args: args{path: retryPath, keys: []string{"l", "d"}},
			want: []string{"r"},
		},
		{
			name: "d refuses the last key",
			args: args{path: retryPath, prepare: func(k *config.KeybindingSettings) { k.Connection.Retry = []string{"r"} }, keys: []string{"d"}},
		},
		{
			name:        "r asks before resetting",
			args:        args{path: retryPath, prepare: func(k *config.KeybindingSettings) { k.Connection.Retry = []string{"x"} }, keys: []string{"r"}},
			wantConfirm: true,
		},
		{
			name: "r then y resets to the default",
			args: args{path: retryPath, prepare: func(k *config.KeybindingSettings) { k.Connection.Retry = []string{"x"} }, keys: []string{"r", "y"}},
			want: []string{"r", "R"},
		},
		{
			name: "r then n keeps the keys",
			args: args{path: retryPath, prepare: func(k *config.KeybindingSettings) { k.Connection.Retry = []string{"x"} }, keys: []string{"r", "n"}},
		},
		{
			name: "r on the default does not ask",
			args: args{path: retryPath, keys: []string{"r"}},
		},
		{
			name: "a second key within the window makes a sequence",
			args: args{path: retryPath, keys: []string{"a", "g", "r"}},
			want: []string{"r", "R", "g r"},
		},
		{
			name: "enter after the first key keeps it alone",
			args: args{path: retryPath, keys: []string{"a", "g", "enter"}},
			want: []string{"r", "R", "g"},
		},
		{
			name:    "a sequence starting with a navigation key is refused",
			args:    args{path: retryPath, keys: []string{"a", "j", "x"}},
			wantErr: config.ErrSequenceStart,
		},
		{
			name: "esc after the first key cancels the sequence",
			args: args{path: retryPath, keys: []string{"a", "g", "esc"}},
		},
		{
			name: "a sequence already on the action is refused",
			args: args{path: "screen.settings", keys: []string{"a", "s", "s"}},
		},
		{
			name: "append-only takes a new key",
			args: args{path: downPath, keys: []string{"a", "ctrl+n"}},
			want: []string{"down", "j", "ctrl+n"},
		},
		{
			name: "append-only refuses to replace a default key",
			args: args{path: downPath, keys: []string{"l", "e"}},
		},
		{
			name: "append-only refuses to remove a default key",
			args: args{path: downPath, keys: []string{"d"}},
		},
		{
			name: "append-only removes an added key",
			args: args{path: downPath, prepare: func(k *config.KeybindingSettings) { k.Navigation.Down = []string{"down", "j", "x"} }, keys: []string{"l", "l", "d"}},
			want: []string{"down", "j"},
		},
		{
			name: "append-only replaces an added key",
			args: args{path: downPath, prepare: func(k *config.KeybindingSettings) { k.Navigation.Down = []string{"down", "j", "x"} }, keys: []string{"l", "l", "e", "n"}},
			want: []string{"down", "j", "n"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, saved := pressKeys(t, newBindsFixture(t, tt.args.path, tt.args.prepare), tt.args.keys...)

			switch {
			case tt.wantErr != nil:
				require.NotNil(t, saved)
				require.ErrorIs(t, saved.err, tt.wantErr)
			case tt.want == nil:
				assert.Nil(t, saved)
			default:
				require.NotNil(t, saved)
				assert.Equal(t, tt.want, keysOf(saved.conm, tt.args.path))
			}
			assert.Equal(t, tt.wantCapture, m.binds.capture)
			assert.Equal(t, tt.wantConfirm, m.binds.confirm)
		})
	}
}

func TestSettingsModel_saveClientCmd(t *testing.T) {
	type args struct {
		prepare func(*config.KeybindingSettings)
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "default keybindings are kept",
		},
		{
			name: "changed keybindings are kept",
			args: args{prepare: func(k *config.KeybindingSettings) { k.Connection.Retry = []string{"x"} }},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newBindsFixture(t, retryPath, tt.args.prepare)
			db, ok := m.database()
			require.True(t, ok)
			db.CLI = cli.USQL

			msg, ok := m.saveClientCmd(db)().(settingsSavedMsg)

			require.True(t, ok)
			require.NoError(t, msg.err)
			assert.Equal(t, m.conm.Keybindings, msg.conm.Keybindings)
			assert.Equal(t, cli.USQL, msg.conm.CLI(config.PostgresConnType))
		})
	}
}

func TestSettingsModel_bindings(t *testing.T) {
	tests := []struct {
		name     string
		wantPath []string
		hidePath []string
	}{
		{
			name:     "locked bindings are not listed",
			wantPath: []string{downPath, retryPath},
			hidePath: []string{"general.interrupt", "general.backspace"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newBindsFixture(t, downPath, nil)

			bindings := m.bindings()
			got := make([]string, 0, len(bindings))
			for _, b := range bindings {
				got = append(got, b.Path)
			}

			for _, path := range tt.wantPath {
				assert.Contains(t, got, path)
			}
			for _, path := range tt.hidePath {
				assert.NotContains(t, got, path)
			}
		})
	}
}
