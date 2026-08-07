package ui

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/registry"
)

type Model struct {
	reg    registry.Registry[config.Postgres]
	states []ConnState

	cursor int

	pinged     bool
	confirming bool

	runErr    error
	changeErr error
}

type ConnState struct {
	pingSpinner spinner.Model

	connPing pingState
}

type pingState string

const (
	undefinedPingState   = "•"
	pingingPingState     = "…"
	pingSuccessPingState = "🟢"
	pingFailedPingState  = "🔴"
)

type pingResultMsg struct {
	index int
	err   error
}

type runResultMsg struct {
	err error
}

type connChangedMsg struct {
	err error
}

func newConnState() ConnState {
	s := spinner.New()
	s.Spinner = spinner.Globe
	return ConnState{
		pingSpinner: s,
		connPing:    undefinedPingState,
	}
}

func New(reg registry.Registry[config.Postgres]) Model {
	states := make([]ConnState, reg.Len())
	for i := range states {
		states[i] = newConnState()
	}
	return Model{
		reg:    reg,
		states: states,
	}
}

func (m Model) Init() tea.Cmd {
	for i := range m.states {
		m.states[i].pingSpinner.Tick()
		m.states[i].connPing = undefinedPingState
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.confirming {
			return m.handleConfirmKey(msg.String())
		}
		return m.handleKey(msg.String())

	case pingResultMsg:
		if msg.index < len(m.states) {
			m.states[msg.index].connPing = pingStateForErr(msg.err)
		}

	case runResultMsg:
		m.runErr = msg.err

	case connChangedMsg:
		return m.applyChange(msg.err), nil

	case spinner.TickMsg:
		return m.tickSpinners(msg)
	}

	return m, nil
}

func (m Model) handleKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "ctrl+c", "q", "esc":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < m.reg.Len()-1 {
			m.cursor++
		}
	case "enter":
		m.runErr = nil
		return m, m.runCmd(m.cursor)
	case "a":
		m.changeErr = nil
		return m, m.addCmd()
	case "e":
		if m.reg.Len() > 0 {
			m.changeErr = nil
			return m, m.editCmd(m.cursor)
		}
	case "d":
		if m.reg.Len() > 0 {
			m.changeErr = nil
			m.confirming = true
		}
	case "p":
		return m.pingAll()
	}
	return m, nil
}

func (m Model) handleConfirmKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "ctrl+c":
		return m, tea.Quit
	case "y":
		m.confirming = false
		m.changeErr = nil
		return m, m.removeCmd(m.cursor)
	case "n", "esc":
		m.confirming = false
	}
	return m, nil
}

func (m Model) pingAll() (tea.Model, tea.Cmd) {
	m.pinged = true
	var cmds []tea.Cmd
	for i := range m.states {
		if m.states[i].connPing == pingingPingState {
			continue
		}
		m.states[i].connPing = pingingPingState
		cmds = append(cmds, m.states[i].pingSpinner.Tick, m.pingCmd(i))
	}
	return m, tea.Batch(cmds...)
}

func (m Model) tickSpinners(msg spinner.TickMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	for i := range m.states {
		if m.states[i].connPing != pingingPingState {
			continue
		}
		var cmd tea.Cmd
		m.states[i].pingSpinner, cmd = m.states[i].pingSpinner.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m Model) applyChange(err error) Model {
	m.changeErr = err
	m = m.syncStates()
	if m.cursor > m.reg.Len()-1 {
		m.cursor = max(m.reg.Len()-1, 0)
	}
	return m
}

func (m Model) syncStates() Model {
	n := m.reg.Len()
	for len(m.states) < n {
		m.states = append(m.states, newConnState())
	}
	if len(m.states) > n {
		m.states = m.states[:n]
	}
	return m
}

func (m Model) cursorLabel() string {
	c, ok := m.reg.Get(m.cursor)
	if !ok {
		return ""
	}
	if name := c.Name(); name != "" {
		return name
	}
	return c.Host()
}

func pingStateForErr(err error) pingState {
	if err != nil {
		return pingFailedPingState
	}
	return pingSuccessPingState
}
