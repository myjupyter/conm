package ui

import (
	"context"
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui/spec"
	"github.com/myjupyter/conm/internal/ui/view"
)

type renderer interface {
	tea.Model
	render() string
}

func secretsScreen(*testing.T) tea.Model { return newSecretModel(view.NewSecrets()) }

func transportsScreen(*testing.T) tea.Model {
	return newTransportModel(view.NewTransports(&stubTransports{}), view.NewConnections())
}

type stubWorkspaces struct{ names []string }

func (s stubWorkspaces) Len() int { return len(s.names) }

func (s stubWorkspaces) Get(i int) (string, bool) {
	if i < 0 || i >= len(s.names) {
		return "", false
	}
	return s.names[i], true
}

func (s stubWorkspaces) Active() string                  { return s.names[0] }
func (stubWorkspaces) Use(int) error                     { return nil }
func (stubWorkspaces) Add(string) error                  { return nil }
func (stubWorkspaces) Rename(int, string) error          { return nil }
func (stubWorkspaces) Remove(context.Context, int) error { return nil }

func workspacesScreen(*testing.T) tea.Model {
	names := make([]string, 30)
	for i := range names {
		names[i] = fmt.Sprintf("ws-%d", i)
	}
	return newWorkspaceModel(view.NewWorkspaces(stubWorkspaces{names: names}))
}

func initScreen(*testing.T) tea.Model { return newInitModel() }

func infoScreen(t *testing.T) tea.Model {
	t.Helper()
	return newInfoFixture(t, []config.WebLink{{Name: "grafana", URL: "https://grafana.internal"}})
}

func connectionFormScreen(*testing.T) tea.Model {
	return newFormModel(config.PostgresConnType, spec.PostgresFormSpec, "add", nil, false, formContext{})
}

func settingsScreen(*testing.T) tea.Model {
	return newSettingsModel(config.Conm{
		Connections: []config.ConnectionSettings{
			{Type: config.PostgresConnType},
			{Type: config.MySQLConnType},
			{Type: config.RedisConnType},
		},
		Keybindings: config.DefaultKeybindings(),
	})
}

func TestScreens_renderFitsWindow(t *testing.T) {
	type args struct {
		model func(t *testing.T) tea.Model
		size  tea.WindowSizeMsg
	}
	tests := []struct {
		name      string
		args      args
		wantWidth int
	}{
		{name: "secrets keep their minimum width", args: args{model: secretsScreen, size: tea.WindowSizeMsg{Width: 60, Height: 40}}, wantWidth: secretInner + 2},
		{name: "secrets stretch to a wide window", args: args{model: secretsScreen, size: tea.WindowSizeMsg{Width: 200, Height: 14}}, wantWidth: 200},
		{name: "transports keep their minimum width", args: args{model: transportsScreen, size: tea.WindowSizeMsg{Width: 60, Height: 40}}, wantWidth: transportInner + 2},
		{name: "transports stretch to a wide window", args: args{model: transportsScreen, size: tea.WindowSizeMsg{Width: 200, Height: 14}}, wantWidth: 200},
		{name: "workspaces stretch to a wide window", args: args{model: workspacesScreen, size: tea.WindowSizeMsg{Width: 200, Height: 14}}, wantWidth: 200},
		{name: "init keeps its minimum width", args: args{model: initScreen, size: tea.WindowSizeMsg{Width: 60, Height: 40}}, wantWidth: initInner + 2},
		{name: "init scrolls in a short window", args: args{model: initScreen, size: tea.WindowSizeMsg{Width: 200, Height: 14}}, wantWidth: 200},
		{name: "info keeps its minimum width", args: args{model: infoScreen, size: tea.WindowSizeMsg{Width: 60, Height: 40}}, wantWidth: infoInner + 2},
		{name: "info stops at the modal width", args: args{model: infoScreen, size: tea.WindowSizeMsg{Width: 250, Height: 40}}, wantWidth: modalMaxInner + 2},
		{name: "info scrolls in a short window", args: args{model: infoScreen, size: tea.WindowSizeMsg{Width: 100, Height: 16}}, wantWidth: 100},
		{name: "connection form keeps its minimum width", args: args{model: connectionFormScreen, size: tea.WindowSizeMsg{Width: 60, Height: 40}}, wantWidth: formInner + 2},
		{name: "connection form stops at the modal width", args: args{model: connectionFormScreen, size: tea.WindowSizeMsg{Width: 250, Height: 40}}, wantWidth: modalMaxInner + 2},
		{name: "connection form scrolls in a short window", args: args{model: connectionFormScreen, size: tea.WindowSizeMsg{Width: 100, Height: 16}}, wantWidth: 100},
		{name: "settings stop at the modal width", args: args{model: settingsScreen, size: tea.WindowSizeMsg{Width: 250, Height: 40}}, wantWidth: modalMaxInner + 2},
		{name: "settings scroll in a short window", args: args{model: settingsScreen, size: tea.WindowSizeMsg{Width: 100, Height: 16}}, wantWidth: 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next, _ := tt.args.model(t).Update(tt.args.size)
			m, ok := next.(renderer)
			require.True(t, ok)

			out := m.render()

			assert.Equal(t, tt.wantWidth, lipgloss.Width(out))
			assert.LessOrEqual(t, len(strings.Split(out, "\n")), tt.args.size.Height)
		})
	}
}

func TestScreenSize_follow(t *testing.T) {
	type args struct {
		size   screenSize
		chrome int
		focus  focusSpan
		total  int
	}
	tests := []struct {
		name      string
		args      args
		wantStart int
		wantEnd   int
	}{
		{name: "unknown height shows every row", args: args{chrome: 8, focus: focusRow(20), total: 30}, wantStart: 0, wantEnd: 30},
		{name: "rows fill what the chrome leaves", args: args{size: screenSize{height: 18}, chrome: 8, total: 30}, wantStart: 0, wantEnd: 10},
		{name: "the window follows the cursor", args: args{size: screenSize{height: 18}, chrome: 8, focus: focusRow(20), total: 30}, wantStart: 11, wantEnd: 21},
		{name: "a focused span stays whole", args: args{size: screenSize{height: 18}, chrome: 8, focus: focusSpan{from: 18, row: 19, to: 21}, total: 30}, wantStart: 12, wantEnd: 22},
		{name: "a span taller than the window keeps its focused row", args: args{size: screenSize{height: 10}, chrome: 8, focus: focusSpan{from: 18, row: 20, to: 21}, total: 30}, wantStart: 19, wantEnd: 21},
		{name: "a window shorter than the chrome keeps one row", args: args{size: screenSize{height: 4}, chrome: 8, focus: focusRow(5), total: 30}, wantStart: 5, wantEnd: 6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rows view.Scroll
			tt.args.size.follow(&rows, tt.args.chrome, tt.args.focus, tt.args.total)

			start, end := rows.Range(tt.args.total)
			assert.Equal(t, tt.wantStart, start)
			assert.Equal(t, tt.wantEnd, end)
		})
	}
}

func TestScreenSize_visibleRows(t *testing.T) {
	type args struct {
		size        screenSize
		followed    int
		chrome      int
		cursor      int
		total       int
		startCursor int
	}
	tests := []struct {
		name      string
		args      args
		wantStart int
		wantEnd   int
	}{
		{name: "no panel shows the whole window", args: args{size: screenSize{height: 18}, followed: 8, chrome: 8, cursor: 3, total: 30}, wantStart: 0, wantEnd: 10},
		{name: "a panel trims the rows below the cursor", args: args{size: screenSize{height: 18}, followed: 8, chrome: 12, cursor: 3, total: 30}, wantStart: 0, wantEnd: 6},
		{name: "a panel trims the rows above a cursor at the bottom", args: args{size: screenSize{height: 18}, followed: 8, chrome: 12, cursor: 9, total: 30}, wantStart: 4, wantEnd: 10},
		{name: "unknown height shows every row", args: args{followed: 8, chrome: 12, cursor: 9, total: 30}, wantStart: 0, wantEnd: 30},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rows view.Scroll
			tt.args.size.follow(&rows, tt.args.followed, focusRow(tt.args.startCursor), tt.args.total)
			tt.args.size.follow(&rows, tt.args.followed, focusRow(tt.args.cursor), tt.args.total)

			start, end := tt.args.size.visibleRows(&rows, tt.args.chrome, tt.args.cursor, tt.args.total)
			assert.Equal(t, tt.wantStart, start)
			assert.Equal(t, tt.wantEnd, end)
		})
	}
}

func TestScreenSize_fit(t *testing.T) {
	type args struct {
		size   screenSize
		chrome int
		cursor int
		total  int
		lines  int
	}
	tests := []struct {
		name     string
		args     args
		wantSize int
	}{
		{name: "unknown height keeps the window", args: args{chrome: 4, total: 10, lines: 3}, wantSize: 2},
		{name: "items of one line fill the budget", args: args{size: screenSize{height: 10}, chrome: 4, total: 10, lines: 1}, wantSize: 6},
		{name: "taller items fit fewer", args: args{size: screenSize{height: 10}, chrome: 4, total: 10, lines: 3}, wantSize: 2},
		{name: "a short list shows whole", args: args{size: screenSize{height: 40}, chrome: 4, total: 3, lines: 3}, wantSize: 3},
		{name: "nothing fits keeps one item", args: args{size: screenSize{height: 5}, chrome: 4, total: 10, lines: 3}, wantSize: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rows view.Scroll
			rows.Resize(2)
			tt.args.size.fit(&rows, tt.args.chrome, tt.args.cursor, tt.args.total, func() int {
				start, end := rows.Range(tt.args.total)
				return (end - start) * tt.args.lines
			})

			start, end := rows.Range(tt.args.total)
			assert.Equal(t, tt.wantSize, end-start)
		})
	}
}

func TestScreenSize_fitKeepsTheWindowWhileTheCursorIsInside(t *testing.T) {
	type args struct {
		settled int
		cursor  int
	}
	tests := []struct {
		name      string
		args      args
		wantStart int
		wantEnd   int
	}{
		{name: "the last row", args: args{settled: 5, cursor: 19}, wantStart: 5, wantEnd: 20},
		{name: "one row up from the end", args: args{settled: 5, cursor: 18}, wantStart: 5, wantEnd: 20},
		{name: "the top of the window", args: args{settled: 5, cursor: 5}, wantStart: 5, wantEnd: 20},
		{name: "above the window", args: args{settled: 5, cursor: 4}, wantStart: 4, wantEnd: 19},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			const budget, total, fitting = 20, 20, 15

			var rows view.Scroll
			rows.Resize(fitting)
			rows.Follow(tt.args.settled+fitting-1, total)

			size := screenSize{height: budget}
			size.fit(&rows, 0, tt.args.cursor, total, func() int {
				start, end := rows.Range(total)
				if end-start > fitting {
					return budget + 1
				}
				return end - start
			})

			start, end := rows.Range(total)
			assert.Equal(t, tt.wantStart, start)
			assert.Equal(t, tt.wantEnd, end)
		})
	}
}

func TestScreenSize_join(t *testing.T) {
	type args struct {
		size  screenSize
		lines []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "unknown height keeps every line", args: args{lines: []string{"a", "b", "c", "d"}}, want: "a\nb\nc\nd"},
		{name: "a frame that fits is untouched", args: args{size: screenSize{height: 4}, lines: []string{"a", "b", "c", "d"}}, want: "a\nb\nc\nd"},
		{name: "overflow goes from above the footer", args: args{size: screenSize{height: 3}, lines: []string{"a", "b", "c", "d"}}, want: "a\nc\nd"},
		{name: "a window no taller than the footer is left alone", args: args{size: screenSize{height: 2}, lines: []string{"a", "b", "c"}}, want: "a\nb\nc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.args.size.join(tt.args.lines))
		})
	}
}

func TestScreenSize_keepEnd(t *testing.T) {
	type args struct {
		size   screenSize
		lines  []string
		chrome int
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{name: "unknown height keeps every line", args: args{lines: []string{"a", "b"}, chrome: 4}, want: []string{"a", "b"}},
		{name: "lines that fit are kept", args: args{size: screenSize{height: 6}, lines: []string{"a", "b"}, chrome: 4}, want: []string{"a", "b"}},
		{name: "the end survives a short budget", args: args{size: screenSize{height: 5}, lines: []string{"a", "b"}, chrome: 4}, want: []string{"b"}},
		{name: "no budget still keeps one line", args: args{size: screenSize{height: 2}, lines: []string{"a", "b"}, chrome: 4}, want: []string{"b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.args.size.keepEnd(tt.args.lines, tt.args.chrome))
		})
	}
}

func press(t *testing.T, m tea.Model, msgs ...tea.Msg) renderer {
	t.Helper()
	for _, msg := range msgs {
		m, _ = m.Update(msg)
	}
	r, ok := m.(renderer)
	require.True(t, ok)
	return r
}

func repeatKey(code rune, n int) []tea.Msg {
	msgs := make([]tea.Msg, n)
	for i := range msgs {
		msgs[i] = tea.KeyPressMsg{Code: code}
	}
	return msgs
}

func tableScreen(*testing.T) tea.Model {
	confs := make([]config.Connection, 30)
	for i := range confs {
		confs[i] = config.Postgres{Metadata: config.ConnMeta{Name: fmt.Sprintf("conn-%d", i)}}
	}
	return New(view.NewConnections(stubConnections{confs: confs}), nil, nil)
}

func TestScreens_linesStayInTheFrame(t *testing.T) {
	long := strings.Repeat("n", 200)
	longConn := config.Postgres{Metadata: config.ConnMeta{Name: long}, Hostname: long, PortNumber: 5432}
	tests := []struct {
		name      string
		model     func(t *testing.T) tea.Model
		msgs      []tea.Msg
		wantWidth int
	}{
		{
			name:      "workspaces at their minimum width while scrolled",
			model:     workspacesScreen,
			msgs:      append([]tea.Msg{tea.WindowSizeMsg{Width: 62, Height: 14}}, repeatKey(tea.KeyDown, 15)...),
			wantWidth: 62,
		},
		{
			name: "info with a long name",
			model: func(*testing.T) tea.Model {
				return newInfoModel(view.NewConnections(stubConnections{confs: []config.Connection{longConn}}))
			},
			msgs:      []tea.Msg{tea.WindowSizeMsg{Width: 120, Height: 40}},
			wantWidth: 120,
		},
		{
			name: "edit form with a long name",
			model: func(*testing.T) tea.Model {
				return newFormModel(config.PostgresConnType, spec.PostgresFormSpec, "add", map[string]string{"name": long}, true, formContext{})
			},
			msgs:      []tea.Msg{tea.WindowSizeMsg{Width: 120, Height: 40}},
			wantWidth: 120,
		},
		{
			name: "table with wide characters",
			model: func(*testing.T) tea.Model {
				return New(view.NewConnections(stubConnections{confs: []config.Connection{
					config.Postgres{Metadata: config.ConnMeta{Name: "数据库生产主库数据库生产主库数据库生产主库"}, Hostname: "主机", PortNumber: 1},
				}}), nil, nil)
			},
			msgs:      []tea.Msg{tea.WindowSizeMsg{Width: 130, Height: 40}},
			wantWidth: 130,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := press(t, tt.model(t), tt.msgs...).render()

			for line := range strings.SplitSeq(out, "\n") {
				assert.Equal(t, tt.wantWidth, lipgloss.Width(line), line)
			}
		})
	}
}

func TestWorkspaces_topLineKeepsTheScrollNote(t *testing.T) {
	msgs := append([]tea.Msg{tea.WindowSizeMsg{Width: 62, Height: 14}}, repeatKey(tea.KeyDown, 15)...)

	first, _, _ := strings.Cut(press(t, workspacesScreen(t), msgs...).render(), "\n")

	assert.Contains(t, first, "below")
	assert.Contains(t, first, "above")
}

func TestScreens_helpStaysInTheWindow(t *testing.T) {
	screens := []struct {
		name  string
		model func(t *testing.T) tea.Model
	}{
		{name: "table", model: tableScreen},
		{name: "secrets", model: secretsScreen},
		{name: "transports", model: transportsScreen},
		{name: "workspaces", model: workspacesScreen},
		{name: "init", model: initScreen},
		{name: "info", model: infoScreen},
		{name: "connection form", model: connectionFormScreen},
		{name: "settings", model: settingsScreen},
	}
	for _, tt := range screens {
		for _, height := range []int{10, 14, 18, 22} {
			t.Run(fmt.Sprintf("%s at %d", tt.name, height), func(t *testing.T) {
				out := press(t, tt.model(t),
					tea.WindowSizeMsg{Width: 130, Height: height},
					tea.KeyPressMsg{Code: '?', Text: "?"},
				).render()
				lines := strings.Split(out, "\n")

				assert.Len(t, lines, min(len(lines), height))
				assert.Contains(t, lines[len(lines)-2], "exit from help")
				assert.Contains(t, lines[len(lines)-1], gCornerBL)
			})
		}
	}
}

func metaTabMsgs(height int) []tea.Msg {
	return []tea.Msg{tea.WindowSizeMsg{Width: 100, Height: height}, tea.KeyPressMsg{Code: tea.KeyTab}, tea.KeyPressMsg{Code: tea.KeyTab}}
}

func TestForm_linksHintIsReachable(t *testing.T) {
	for _, height := range []int{16, 18, 20} {
		t.Run(fmt.Sprintf("height %d", height), func(t *testing.T) {
			msgs := append(metaTabMsgs(height), repeatKey(tea.KeyDown, 2)...)

			out := press(t, connectionFormScreen(t), msgs...).render()

			assert.Contains(t, out, "nothing linked yet")
			assert.NotContains(t, out, "below")
		})
	}
}

func TestForm_focusedFieldIsVisibleUnderASectionHead(t *testing.T) {
	for _, height := range []int{10, 11, 12, 13} {
		t.Run(fmt.Sprintf("height %d", height), func(t *testing.T) {
			out := press(t, connectionFormScreen(t), metaTabMsgs(height)...).render()

			assert.Contains(t, out, gCaret)
		})
	}
}

func TestSettings_selectedRowIsVisibleInAShortWindow(t *testing.T) {
	for _, height := range []int{10, 11, 12, 13, 14} {
		t.Run(fmt.Sprintf("height %d", height), func(t *testing.T) {
			out := press(t, settingsScreen(t), tea.WindowSizeMsg{Width: 100, Height: height}).render()

			assert.Contains(t, out, gCaret)
			assert.LessOrEqual(t, len(strings.Split(out, "\n")), height)
		})
	}
}

func TestSettings_resetPromptStaysReadable(t *testing.T) {
	keybindings := config.DefaultKeybindings()
	keybindings.Navigation.PrevSection = []string{"shift+tab", "ctrl+p"}
	m := newSettingsModel(config.Conm{Keybindings: keybindings})
	m.tab, m.binds.cursor = settingsKeybindTab, 5

	out := press(t, m, tea.WindowSizeMsg{Width: 90, Height: 40}, tea.KeyPressMsg{Code: 'r', Text: "r"}).render()

	assert.Contains(t, out, "y / n")
	assert.Contains(t, out, "default shift+tab")
	for line := range strings.SplitSeq(out, "\n") {
		assert.Equal(t, 90, lipgloss.Width(line), line)
	}
}

func TestSettings_keybindListDoesNotScrollInsideTheWindow(t *testing.T) {
	bindings := len(config.DefaultKeybindings().Bindings())
	m := newSettingsModel(config.Conm{Keybindings: config.DefaultKeybindings()})
	m.tab = settingsKeybindTab
	var model tea.Model = m
	model, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyUp})

	settled, ok := model.(settingsModel)
	require.True(t, ok)
	wantStart, _ := settled.binds.window.Range(bindings)

	model, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	moved, ok := model.(settingsModel)
	require.True(t, ok)
	gotStart, gotEnd := moved.binds.window.Range(bindings)

	assert.Equal(t, wantStart, gotStart)
	assert.Greater(t, gotEnd-1, moved.binds.cursor)
}
