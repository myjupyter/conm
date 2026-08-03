package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
)

func RunSelectConnForm(title string, conns []config.ConnectionConfig) (int, bool, error) {
	m, err := tea.NewProgram(newSelectConnModel(title, conns)).Run()
	if err != nil {
		return 0, false, err
	}

	sm := m.(selectConnModel)
	if !sm.chosen {
		return 0, false, nil
	}

	return sm.cursor, true, nil
}

type selectConnModel struct {
	title string
	conns []config.ConnectionConfig

	cursor int
	chosen bool
}

func newSelectConnModel(title string, conns []config.ConnectionConfig) selectConnModel {
	return selectConnModel{
		title: title,
		conns: conns,
	}
}

func (m selectConnModel) Init() tea.Cmd {
	return nil
}

func (m selectConnModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.cursor < len(m.conns)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.conns) > 0 {
				m.chosen = true
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m selectConnModel) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m selectConnModel) render() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(m.title))
	b.WriteByte('\n')

	for i, c := range m.conns {
		line := connLabel(c)
		if i == m.cursor {
			b.WriteString(selectedStyle.Render("> " + line))
		} else {
			b.WriteString(itemStyle.Render("  " + line))
		}
		b.WriteByte('\n')
	}

	b.WriteString(helpStyle.Render("↑/k up · ↓/j down · enter select · q/esc quit"))
	return b.String()
}

func connLabel(c config.ConnectionConfig) string {
	label := c.Name()
	if label == "" {
		label = fmt.Sprintf("%s:%d/%s", c.Host(), c.Port(), c.Database())
	}
	if !c.IsValid() {
		label += "  " + errorStyle.Render("(broken)")
	}
	return label
}
