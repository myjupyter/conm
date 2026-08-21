package ui

import (
	"fmt"
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui/spec"
)

const (
	formInner  = 88
	formLabelW = 12
	formBoxW   = 46
)

func (m formModel) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m formModel) render() string {
	lines := []string{
		m.formTopLine(),
		m.badgeLine(),
		frameLine(formInner, nil, nil),
		m.tabLine(),
		m.dividerLine(),
		m.noteLine(),
	}

	for _, i := range m.sectionFields(m.section) {
		lines = append(lines, m.fieldRow(i), m.messageRow(i))
	}
	lines = append(lines, frameLine(formInner, nil, nil))

	if m.pong != "" {
		lines = append(lines, frameRule(formInner, gTeeL, gTeeR), m.pongLine())
	}
	if m.ping != nil {
		lines = append(lines, frameErrPanel(formInner, m.ping)...)
	}

	lines = append(lines, frameRule(formInner, gTeeL, gTeeR), m.formStatusLine())
	lines = append(lines, frameKeybar(formInner, m.formBinds())...)
	lines = append(lines, frameBottom(formInner))
	return strings.Join(lines, "\n")
}

func (m formModel) formTopLine() string {
	crumb := "new connection"
	if m.isEdit {
		crumb = "edit connection"
	}
	return frameTop(formInner, crumb, "")
}

func (m formModel) badgeLine() string {
	title := "new connection"
	if m.isEdit {
		name := strings.TrimSpace(m.vals["name"])
		if name == "" {
			name = strings.TrimSpace(m.vals["host"])
		}
		title = "editing " + name
	}
	return frameBadge(formInner, "postgres", typeColor(config.PostgresConnType), title)
}

func (m formModel) tabLine() string {
	spans := []span{{text: "  ", fg: cDim}}
	for si, s := range m.sections {
		if si == m.section {
			spans = append(spans, span{text: " " + s.Title + " ", fg: cInvFg, bg: cAccent, bold: true})
		} else {
			spans = append(spans, span{text: " " + s.Title + " ", fg: cSoft})
		}
		if m.attempted {
			if n := m.sectionErrCount(si); n > 0 {
				spans = append(spans, span{text: fmt.Sprintf(" %d ", n), fg: cErrTagFg, bg: cErrTagBg, bold: true})
			}
		}
		spans = append(spans, span{text: "  ", fg: cDim})
	}
	spans = append(spans, span{text: "tab to switch", fg: cFaint})
	return frameLine(formInner, spans, nil)
}

func (m formModel) dividerLine() string {
	return frameLine(formInner, []span{
		{text: "  ", fg: cDim},
		{text: strings.Repeat(gLineH, formInner-4), fg: cBorder},
	}, nil)
}

func (m formModel) noteLine() string {
	return frameLine(formInner, []span{
		{text: "  ", fg: cDim},
		{text: m.sections[m.section].Note, fg: cFaint},
	}, nil)
}

func (m formModel) fieldRow(i int) string {
	f := m.spec.Fields[i]
	active := i == m.currentField()

	var bg color.Color
	if active {
		bg = cAccent
	}

	caret := "   "
	if active {
		caret = " " + gCaret + " "
	}
	gutter, gc := m.gutter(i, active)

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
	spans = append(spans, m.inputBox(i, active, bg)...)
	return frameLine(formInner, spans, bg)
}

func (m formModel) inputBox(i int, active bool, bg color.Color) []span {
	f := m.spec.Fields[i]

	borderFg := cBorder
	if active {
		borderFg = cInvFg
	}
	if m.attempted && m.fieldError(i) != "" {
		borderFg = cRed
	}

	if f.Kind == spec.SelectFieldKind {
		value := m.vals[f.Key]
		at := 1
		for j, opt := range f.Options {
			if opt == value {
				at = j + 1
				break
			}
		}
		pos := fmt.Sprintf("%d/%d", at, len(f.Options))
		mid := max(formBoxW-8-len([]rune(pos)), 0)
		value = truncPad(value, mid, false)
		left := max((mid-len([]rune(value)))/2, 0)
		right := max(mid-left-len([]rune(value)), 0)

		valFg, posFg := cValue, cMuted
		if active {
			valFg, posFg = cInvFg, cInvFg
		}
		return []span{
			{text: gLineV + " " + gSelectPrev + " ", fg: borderFg, bg: bg},
			{text: strings.Repeat(" ", left), bg: bg},
			{text: value, fg: valFg, bg: bg, bold: true},
			{text: strings.Repeat(" ", right), bg: bg},
			{text: pos + " ", fg: posFg, bg: bg},
			{text: gSelectNext + " " + gLineV, fg: borderFg, bg: bg},
		}
	}

	disp, ghost := m.display(i, active)
	if active && m.insert {
		disp += gInputCursor
	}
	valFg := cValue
	if ghost {
		valFg = cFaint
	}
	if active {
		valFg = cInvFg
	}
	return []span{
		{text: gLineV, fg: borderFg, bg: bg},
		{text: truncPad(disp, formBoxW-2, false), fg: valFg, bg: bg},
		{text: gLineV, fg: borderFg, bg: bg},
	}
}

func (m formModel) display(i int, active bool) (string, bool) {
	f := m.spec.Fields[i]
	raw := m.vals[f.Key]

	// A reference is a name, not a password: it is shown as typed, never
	// masked, because hiding which entry is attached helps nobody.
	if m.isSecretField(i) && m.isRef() {
		if raw == "" {
			return "no entry — ←/→ to pick", true
		}
		return raw, false
	}

	if raw != "" {
		if f.Kind == spec.HiddenFieldKind && !m.reveal {
			return strings.Repeat(gInputMask, len([]rune(raw))), false
		}
		return raw, false
	}
	switch {
	case f.DefaultValue != "":
		return f.DefaultValue + " (default)", true
	case active:
		return f.Example, true
	case f.Property == spec.RequiredFieldProperty:
		return "", true
	default:
		return gEmpty, true
	}
}

func (m formModel) gutter(i int, active bool) (string, color.Color) {
	f := m.spec.Fields[i]
	raw := m.vals[f.Key]

	var g string
	var c color.Color
	switch {
	case m.attempted && m.fieldError(i) != "":
		g, c = gFieldBad, cRed
	case m.isSecretField(i) && m.isRef() && raw == "":
		g, c = gDotOff, cFaint
	case raw != "" || f.Kind == spec.SelectFieldKind || f.DefaultValue != "":
		g, c = gDotOn, cAccent
	default:
		g, c = gDotOff, cFaint
	}
	if active {
		c = cInvFg
	}
	return g, c
}

func (m formModel) messageRow(i int) string {
	text := ""
	col := cRed
	if m.attempted {
		if e := m.fieldError(i); e != "" {
			text = gFieldBad + " " + e
		}
	}
	prefix := strings.Repeat(" ", 5+formLabelW+1)
	return frameLine(formInner, []span{
		{text: prefix, fg: cDim},
		{text: truncPad(text, formInner-len([]rune(prefix))-2, false), fg: col},
	}, nil)
}

func (m formModel) pongLine() string {
	return framePong(formInner, m.pong)
}

func (m formModel) formStatusLine() string {
	icon, col := statusGlyph(m.statusKind)
	return frameStatus(formInner, icon, col, m.status, cursorPos(m.idx, len(m.sectionFields(m.section))))
}

func (m formModel) formBinds() []keybind {
	if m.insert {
		return []keybind{
			{"esc", "done"}, {"enter", "next field"}, {"tab", "section"},
		}
	}

	binds := []keybind{
		{"↑↓/jk", "move"}, {"tab", "section"}, {"←/→", "select"}, {"e", "edit"},
	}
	switch cur := m.currentField(); {
	case m.isProviderField(cur) && m.isRef():
		binds = append(binds, keybind{"s", m.storeLabel()})
	case m.spec.Fields[cur].Kind == spec.HiddenFieldKind && !m.isRef():
		label := "show"
		if m.reveal {
			label = "hide"
		}
		binds = append(binds, keybind{"s", label})
	}
	return append(binds,
		keybind{"enter", "submit"},
		keybind{"p", "ping"},
		keybind{"esc", "cancel"},
	)
}

func caretColor(active bool) color.Color {
	if active {
		return cInvFg
	}
	return cAccent
}
