package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
)

func RunSelectCLIForm(clients []config.CLIInfo) (config.CLIInfo, bool, error) {
	m, err := tea.NewProgram(newClientModel(clients)).Run()
	if err != nil {
		return config.CLIInfo{}, false, err
	}

	pm := m.(initCLIModel)
	if !pm.chosen {
		return config.CLIInfo{}, false, nil
	}

	return pm.clis[pm.cursor], true, nil
}

type initCLIModel struct {
	clis []config.CLIInfo

	cursor int
	chosen bool
}

func newClientModel(clients []config.CLIInfo) initCLIModel {
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

	return initCLIModel{
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

func (m initCLIModel) Init() tea.Cmd {
	return nil
}

func (m initCLIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m initCLIModel) prevSelectable(from int) int {
	for i := from - 1; i >= 0; i-- {
		if m.clis[i].Exists {
			return i
		}
	}
	return from
}

func (m initCLIModel) nextSelectable(from int) int {
	for i := from + 1; i < len(m.clis); i++ {
		if m.clis[i].Exists {
			return i
		}
	}
	return from
}

func (m initCLIModel) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m initCLIModel) render() string {
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
