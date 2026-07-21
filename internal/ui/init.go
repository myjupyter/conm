package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/myjupyter/conm/internal/config"
)

var (
	inactiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Padding(0, 1)

	pathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)

func SelectPGClient(clients []config.PGClientInfo) (string, bool, error) {
	m, err := tea.NewProgram(newPGClientModel(clients)).Run()
	if err != nil {
		return "", false, err
	}

	pm := m.(pgClientModel)
	if !pm.chosen {
		return "", false, nil
	}
	return pm.clients[pm.cursor].Name, true, nil
}

type pgClientModel struct {
	clients []config.PGClientInfo

	cursor int
	chosen bool
}

func newPGClientModel(clients []config.PGClientInfo) pgClientModel {
	ordered := make([]config.PGClientInfo, 0, len(clients))
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
		clients: ordered,
		cursor:  firstSelectable(ordered),
	}
}

func firstSelectable(clients []config.PGClientInfo) int {
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
			if m.cursor < len(m.clients) && m.clients[m.cursor].Exists {
				m.chosen = true
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m pgClientModel) prevSelectable(from int) int {
	for i := from - 1; i >= 0; i-- {
		if m.clients[i].Exists {
			return i
		}
	}
	return from
}

func (m pgClientModel) nextSelectable(from int) int {
	for i := from + 1; i < len(m.clients); i++ {
		if m.clients[i].Exists {
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

	for i, c := range m.clients {
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
