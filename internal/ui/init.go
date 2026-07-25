package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/myjupyter/conm/internal/config"
)

var pathStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#626262"))

func SelectCLI(clients []config.CLIInfo) (string, bool, error) {
	m, err := tea.NewProgram(newPGClientModel(clients)).Run()
	if err != nil {
		return "", false, err
	}

	pm := m.(pgClientModel)
	if !pm.chosen {
		return "", false, nil
	}
	return pm.clis[pm.cursor].Name, true, nil
}

type pgClientModel struct {
	clis []config.CLIInfo

	cursor int
	chosen bool
}

func newPGClientModel(clients []config.CLIInfo) pgClientModel {
	ordered := make([]config.CLIInfo, 0, len(clients))
	for _, c := range clients {
		if c.Exists {
			ordered = append(ordered, c)
		}
	}
	for _, c := range clients {
		if !c.Exists {
			ordered = append(ordered, c)
		}
	}

	return pgClientModel{
		clis:   ordered,
		cursor: firstSelectable(ordered),
	}
}

func firstSelectable(clients []config.CLIInfo) int {
	for i, c := range clients {
		if c.Exists {
			return i
		}
	}
	return 0
}

func (m pgClientModel) Init() tea.Cmd {
	return nil
}

func (m pgClientModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "up", "k":
			m.cursor = m.prevSelectable(m.cursor)
		case "down", "j":
			m.cursor = m.nextSelectable(m.cursor)
		case "enter":
			if m.cursor < len(m.clis) && m.clis[m.cursor].Exists {
				m.chosen = true
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m pgClientModel) prevSelectable(from int) int {
	for i := from - 1; i >= 0; i-- {
		if m.clis[i].Exists {
			return i
		}
	}
	return from
}

func (m pgClientModel) nextSelectable(from int) int {
	for i := from + 1; i < len(m.clis); i++ {
		if m.clis[i].Exists {
			return i
		}
	}
	return from
}

func (m pgClientModel) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m pgClientModel) render() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Choose a Postgres client"))
	b.WriteByte('\n')

	for i, c := range m.clis {
		if c.Exists {
			line := c.Name + "  " + pathStyle.Render(c.Path)
			if i == m.cursor {
				b.WriteString(selectedStyle.Render("> " + line))
			} else {
				b.WriteString(itemStyle.Render("  " + line))
			}
		}
		b.WriteByte('\n')
	}

	b.WriteString(helpStyle.Render("↑/k up · ↓/j down · enter select · q/esc quit"))
	return b.String()
}

func SelectPGPassImport() (bool, bool, error) {
	m, err := tea.NewProgram(newPGPassModel()).Run()
	if err != nil {
		return false, false, err
	}

	pm := m.(pgPassModel)
	if !pm.chosen {
		return false, false, nil
	}
	return pm.cursor == 0, true, nil
}

type pgPassModel struct {
	options []string

	cursor int
	chosen bool
}

func newPGPassModel() pgPassModel {
	return pgPassModel{
		options: []string{"Import from .pgpass", "Skip"},
	}
}

func (m pgPassModel) Init() tea.Cmd {
	return nil
}

func (m pgPassModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case "enter":
			m.chosen = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m pgPassModel) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m pgPassModel) render() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Import connections from .pgpass?"))
	b.WriteByte('\n')

	for i, opt := range m.options {
		if i == m.cursor {
			b.WriteString(selectedStyle.Render("> " + opt))
		} else {
			b.WriteString(itemStyle.Render("  " + opt))
		}
		b.WriteByte('\n')
	}

	b.WriteString(helpStyle.Render("↑/k up · ↓/j down · enter select · q/esc quit"))
	return b.String()
}
