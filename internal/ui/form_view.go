package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/ui/spec"
)

func (m formModel) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m formModel) render() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(m.title))
	b.WriteByte('\n')

	for i, f := range m.spec.Fields {
		focused := m.focus == i

		b.WriteString(" " + labelStyle.Render(f.Label) + requiredMark(f.Property))
		b.WriteByte('\n')

		if focused {
			b.WriteString(markerStyle.Render(">"))
		} else {
			b.WriteString(" ")
		}
		if f.Kind == spec.SelectFieldKind {
			b.WriteString(selectField(f.Options, m.selects[i], focused))
		} else {
			b.WriteString(m.inputs[i].View())
		}
		b.WriteString("\n\n")
	}

	if m.err != "" {
		b.WriteString(errorStyle.Render(m.err))
		b.WriteByte('\n')
	}

	b.WriteString(helpStyle.Render("tab/↑↓ move · ←/→ select · enter submit · esc cancel"))
	return b.String()
}

func requiredMark(p spec.FieldProperty) string {
	if p == spec.RequiredFieldProperty {
		return requiredStyle.Render(" *")
	}
	return ""
}

func selectField(options []string, cursor int, focused bool) string {
	value := "‹ " + options[cursor] + " ›"
	if focused {
		return markerStyle.Render(value)
	}
	return value
}
