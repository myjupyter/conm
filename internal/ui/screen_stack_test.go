package ui

import (
	"reflect"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type probeScreen struct {
	name     string
	received []tea.Msg
}

type probeQuitMsg struct{}

type probeEchoMsg struct{ text string }

func (p probeScreen) Init() tea.Cmd { return nil }

func (p probeScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	p.received = append(p.received, msg)
	if _, ok := msg.(probeQuitMsg); ok {
		return p, tea.Quit
	}

	return p, nil
}

func (p probeScreen) View() tea.View { return tea.NewView(p.name) }

func Test_wrapScreenCmd(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cmd  tea.Cmd
		want tea.Msg
	}{
		{
			name: "a nil command stays nil",
			cmd:  nil,
			want: nil,
		},
		{
			name: "quitting closes the screen instead of the program",
			cmd:  tea.Quit,
			want: screenDoneMsg{id: 7},
		},
		{
			name: "a screen's own message is addressed back to it",
			cmd:  func() tea.Msg { return probeEchoMsg{text: "pong"} },
			want: screenMsg{id: 7, msg: probeEchoMsg{text: "pong"}},
		},
		{
			name: "a message for the program passes through",
			cmd:  tea.ClearScreen,
			want: tea.ClearScreen(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			wrapped := wrapScreenCmd(7, tt.cmd)
			if tt.cmd == nil {
				assert.Nil(t, wrapped)
				return
			}

			assert.Equal(t, tt.want, wrapped())
		})
	}
}

func Test_wrapScreenCmd_batch(t *testing.T) {
	t.Parallel()

	batch := wrapScreenCmd(3, tea.Batch(tea.Quit, func() tea.Msg { return probeEchoMsg{} }))()

	wrapped, ok := batch.(tea.BatchMsg)
	require.True(t, ok)
	require.Len(t, wrapped, 2)
	assert.Equal(t, screenDoneMsg{id: 3}, wrapped[0]())
	assert.Equal(t, screenMsg{id: 3, msg: probeEchoMsg{}}, wrapped[1]())
}

func Test_wrapScreenCmd_sequence(t *testing.T) {
	t.Parallel()

	sequence := wrapScreenCmd(3, tea.Sequence(tea.Quit, func() tea.Msg { return probeEchoMsg{} }))()

	value := reflect.ValueOf(sequence)
	require.Equal(t, teaPackage, value.Type().PkgPath())
	require.True(t, value.Type().ConvertibleTo(cmdsType))
	wrapped, ok := reflect.TypeAssert[[]tea.Cmd](value.Convert(cmdsType))
	require.True(t, ok)
	require.Len(t, wrapped, 2)
	assert.Equal(t, screenDoneMsg{id: 3}, wrapped[0]())
	assert.Equal(t, screenMsg{id: 3, msg: probeEchoMsg{}}, wrapped[1]())
}

func Test_wrapScreenCmd_exec(t *testing.T) {
	t.Parallel()

	request := execCmd(func() error { return nil }, func(error) tea.Msg { return probeEchoMsg{} })
	msg := wrapScreenCmd(3, request)()

	assert.Equal(t, teaPackage, reflect.TypeOf(msg).PkgPath())
}

func Test_runScreen_withoutHost(t *testing.T) {
	t.Parallel()

	_, err := runScreen(probeScreen{})

	assert.ErrorIs(t, err, errNoScreenHost)
}

func Test_screenStack(t *testing.T) {
	t.Parallel()

	var stack tea.Model = screenStack{last: tea.View{AltScreen: true}}
	stack, _ = stack.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	tableDone, formDone := make(chan tea.Model, 1), make(chan tea.Model, 1)
	stack, _ = stack.Update(pushScreenMsg{model: probeScreen{name: "table"}, done: tableDone})
	stack, _ = stack.Update(pushScreenMsg{model: probeScreen{name: "form"}, done: formDone})
	assert.Equal(t, "form", stack.View().Content)

	stack, _ = stack.Update(screenMsg{id: 1, msg: probeEchoMsg{text: "ping result"}})
	stack, _ = stack.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	stack, cmd := stack.Update(probeQuitMsg{})
	require.NotNil(t, cmd)
	stack, _ = stack.Update(cmd())

	form, ok := (<-formDone).(probeScreen)
	require.True(t, ok)
	assert.Equal(t, []tea.Msg{
		tea.WindowSizeMsg{Width: 80, Height: 24},
		tea.KeyPressMsg{Code: 'x', Text: "x"},
		probeQuitMsg{},
	}, form.received)
	assert.Equal(t, "table", stack.View().Content)

	stack, cmd = stack.Update(probeQuitMsg{})
	stack, _ = stack.Update(cmd())

	table, ok := (<-tableDone).(probeScreen)
	require.True(t, ok)
	assert.Equal(t, []tea.Msg{
		tea.WindowSizeMsg{Width: 80, Height: 24},
		probeEchoMsg{text: "ping result"},
		probeQuitMsg{},
	}, table.received)
	assert.Equal(t, "table", stack.View().Content)
}
