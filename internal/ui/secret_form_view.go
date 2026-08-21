package ui

import (
	"fmt"
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/ui/spec"
)

func (m secretFormModel) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m secretFormModel) render() string {
	lines := []string{
		m.secretFormTopLine(),
		m.secretBadgeLine(),
		boxLineW(formInner, nil, nil),
	}

	for i := range m.spec.Fields {
		lines = append(lines, m.secretFieldRow(i), m.secretMessageRow(i))
	}

	lines = append(lines,
		boxLineW(formInner, nil, nil),
		ruleW(formInner, "├", "┤"),
		m.secretFormStatusLine(),
	)
	lines = append(lines, keybarLinesW(formInner, m.secretFormBinds())...)
	lines = append(lines, border("└")+border(strings.Repeat("─", formInner))+border("┘"))
	return strings.Join(lines, "\n")
}

func (m secretFormModel) secretFormTopLine() string {
	crumb := "new keyring entry"
	if m.isEdit {
		crumb = "edit keyring entry"
	}
	return topLineW(formInner, crumb, "")
}

func (m secretFormModel) secretBadgeLine() string {
	title := m.title
	if m.isEdit {
		if id := strings.TrimSpace(m.vals[spec.KeyringIDKey]); id != "" {
			title = "editing " + id
		}
	}
	return boxLineW(formInner, []span{
		{text: "  ", fg: cDim},
		{text: " keyring ", fg: cInvFg, bg: cAccent, bold: true},
		{text: "  " + title, fg: cFg},
	}, nil)
}

func (m secretFormModel) secretFieldRow(i int) string {
	f := m.spec.Fields[i]
	active := i == m.idx

	var bg color.Color
	if active {
		bg = cAccent
	}

	caret := "   "
	if active {
		caret = " ❯ "
	}
	gutter, gc := m.secretGutter(i, active)

	label := strings.ToLower(f.Label)
	if f.Property == spec.RequiredFieldProperty {
		label += " *"
	}
	labelFg := cSoft
	if active {
		labelFg = cInvFg
	}

	spans := []span{
		{text: caret, fg: caretColor(active), bg: bg, bold: true},
		{text: gutter + " ", fg: gc, bg: bg},
		{text: truncPad(label, formLabelW, false), fg: labelFg, bg: bg, bold: active},
	}
	spans = append(spans, m.secretInputBox(i, active, bg)...)
	return boxLineW(formInner, spans, bg)
}

func (m secretFormModel) secretInputBox(i int, active bool, bg color.Color) []span {
	borderFg := cBorder
	if active {
		borderFg = cInvFg
	}
	if m.attempted && m.fieldError(i) != "" {
		borderFg = cRed
	}

	disp, ghost := m.secretDisplay(i, active)
	if active && m.insert {
		disp += "█"
	}
	valFg := cValue
	if ghost {
		valFg = cFaint
	}
	if active {
		valFg = cInvFg
	}
	return []span{
		{text: "│", fg: borderFg, bg: bg},
		{text: truncPad(disp, formBoxW-2, false), fg: valFg, bg: bg},
		{text: "│", fg: borderFg, bg: bg},
	}
}

func (m secretFormModel) secretDisplay(i int, active bool) (string, bool) {
	f := m.spec.Fields[i]
	raw := m.vals[f.Key]
	if raw != "" {
		if f.Kind == spec.HiddenFieldKind && !m.reveal {
			return strings.Repeat("•", len([]rune(raw))), false
		}
		return raw, false
	}
	switch {
	// An edit leaves the password blank on purpose: the material is write-only,
	// and an untouched field must not look like an empty password.
	case f.Kind == spec.HiddenFieldKind && m.isEdit:
		return "unchanged", true
	case f.Kind == spec.HiddenFieldKind:
		return "already in the keyring", true
	case active:
		return f.Example, true
	case f.Property == spec.RequiredFieldProperty:
		return "", true
	default:
		return "—", true
	}
}

func (m secretFormModel) secretGutter(i int, active bool) (string, color.Color) {
	f := m.spec.Fields[i]

	var g string
	var c color.Color
	switch {
	case m.attempted && m.fieldError(i) != "":
		g, c = "✗", cRed
	case m.vals[f.Key] != "":
		g, c = "●", cAccent
	default:
		g, c = "○", cFaint
	}
	if active {
		c = cInvFg
	}
	return g, c
}

func (m secretFormModel) secretMessageRow(i int) string {
	text := ""
	if m.attempted {
		if e := m.fieldError(i); e != "" {
			text = "✗ " + e
		}
	}
	prefix := strings.Repeat(" ", 5+formLabelW+1)
	return boxLineW(formInner, []span{
		{text: prefix, fg: cDim},
		{text: truncPad(text, formInner-len([]rune(prefix))-2, false), fg: cRed},
	}, nil)
}

func (m secretFormModel) secretFormStatusLine() string {
	icon, col := statusGlyph(m.statusKind)
	pos := fmt.Sprintf("%d/%d", m.idx+1, len(m.spec.Fields))
	textW := max(formInner-4-len([]rune(pos)), 0)

	return boxLineW(formInner, []span{
		{text: " ", fg: cDim},
		{text: icon, fg: col},
		{text: " " + truncPad(m.status, textW, false), fg: col},
		{text: pos + " ", fg: cMuted},
	}, nil)
}

func (m secretFormModel) secretFormBinds() []struct{ key, label string } {
	if m.insert {
		return []struct{ key, label string }{
			{"esc", "done"}, {"enter", "next field"},
		}
	}

	binds := []struct{ key, label string }{
		{"↑↓/jk", "move"}, {"e", "edit"},
	}
	if m.spec.Fields[m.idx].Kind == spec.HiddenFieldKind {
		label := "show"
		if m.reveal {
			label = "hide"
		}
		binds = append(binds, struct{ key, label string }{"s", label})
	}
	return append(binds,
		struct{ key, label string }{"enter", "store"},
		struct{ key, label string }{"esc", "cancel"},
	)
}
