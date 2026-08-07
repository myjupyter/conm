package ui

import (
	"context"
	"io"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/registry"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#336791")).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#336791")).
			Padding(0, 1)

	itemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCCCCC")).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#336791")).
			Padding(0, 1)

	borderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F5F")).
			MarginTop(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)
)

type Model struct {
	reg    registry.Registry[config.Postgres]
	states []ConnState

	cursor int

	pinged bool

	runErr error
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

func New(reg registry.Registry[config.Postgres]) Model {
	states := make([]ConnState, reg.Len())
	for i := range states {
		s := spinner.New()
		s.Spinner = spinner.Globe
		states[i] = ConnState{
			pingSpinner: s,
		}
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
		switch msg.String() {
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
		case "p":
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

	case pingResultMsg:
		if msg.err != nil {
			m.states[msg.index].connPing = pingFailedPingState
		} else {
			m.states[msg.index].connPing = pingSuccessPingState
		}

	case runResultMsg:
		m.runErr = msg.err

	case spinner.TickMsg:
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

	return m, nil
}

func (m Model) pingCmd(i int) tea.Cmd {
	conn, ok := m.reg.Get(i)
	if !ok {
		return nil
	}
	return func() tea.Msg {
		return pingResultMsg{index: i, err: conn.Ping(context.Background())}
	}
}

func (m Model) runCmd(i int) tea.Cmd {
	conn, ok := m.reg.Get(i)
	if !ok {
		return nil
	}
	return tea.Exec(
		runExec{run: func() error { return conn.Run(context.Background()) }},
		func(err error) tea.Msg { return runResultMsg{err: err} },
	)
}

type runExec struct {
	run func() error
}

func (r runExec) Run() error        { return r.run() }
func (runExec) SetStdin(io.Reader)  {}
func (runExec) SetStdout(io.Writer) {}
func (runExec) SetStderr(io.Writer) {}

func (m Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m Model) render() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Postgres connections"))
	b.WriteByte('\n')

	if m.reg.Len() == 0 {
		b.WriteString(itemStyle.Render("No connections found."))
		b.WriteByte('\n')
		b.WriteString(helpStyle.Render("q/esc: quit"))
		return b.String()
	}

	headers := []string{"NAME", "USERNAME", "HOST", "PORT", "DBNAME", "TAGS"}
	if m.pinged {
		headers = append([]string{"PING"}, headers...)
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(borderStyle).
		Headers(headers...).
		StyleFunc(func(row, _ int) lipgloss.Style {
			switch row {
			case table.HeaderRow:
				return headerStyle
			case m.cursor:
				return selectedStyle
			default:
				return itemStyle
			}
		})

	for i := 0; i < m.reg.Len(); i++ {
		c, ok := m.reg.Get(i)
		if !ok {
			continue
		}
		row := []string{
			c.Name(),
			c.Username(),
			c.Host(),
			strconv.Itoa(c.Port()),
			c.Database(),
			strings.Join(c.Tags(), ", "),
		}

		if m.pinged {
			status := string(m.states[i].connPing)
			if m.states[i].connPing == pingingPingState {
				status = m.states[i].pingSpinner.View()
			}
			row = append([]string{status}, row...)
		}

		t.Row(row...)
	}

	b.WriteString(t.String())
	b.WriteByte('\n')
	if m.runErr != nil {
		b.WriteString(errorStyle.Render("run failed: " + m.runErr.Error()))
		b.WriteByte('\n')
	}
	b.WriteString(helpStyle.Render("↑/k up · ↓/j down · enter · p ping · q/esc quit"))
	return b.String()
}
