package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type importSelectOption int

const (
	importSelectSkip importSelectOption = iota
	importSelectAdd
	importSelectPGPass
)

func (m importSelectOption) String() string {
	switch m {
	case importSelectSkip:
		return "skip"
	case importSelectAdd:
		return "add"
	case importSelectPGPass:
		return "import from .pgpass"
	default:
		return ""
	}
}

func RunSelectImport() (importSelectOption, bool, error) {
	m, err := tea.NewProgram(newPGPassModel()).Run()
	if err != nil {
		return 0, false, err
	}

	pm, ok := m.(pgPassModel)
	if !ok {
		return 0, false, fmt.Errorf("import form returned an unexpected model %T", m)
	}
	if !pm.chosen {
		return 0, false, nil
	}

	return pm.options[pm.cursor], true, nil
}

type pgPassModel struct {
	options []importSelectOption

	cursor int
	chosen bool
}

func newPGPassModel() pgPassModel {
	return pgPassModel{
		options: []importSelectOption{
			importSelectSkip,
			importSelectAdd,
			importSelectPGPass,
		},
	}
}

func (m pgPassModel) Init() tea.Cmd {
	return nil
}

func (m pgPassModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}

	switch key.String() {
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
			b.WriteString(selectedStyle.Render("> " + opt.String()))
		} else {
			b.WriteString(itemStyle.Render("  " + opt.String()))
		}
		b.WriteByte('\n')
	}

	b.WriteString(helpStyle.Render("↑/k up · ↓/j down · enter select · q/esc quit"))
	return b.String()
}
