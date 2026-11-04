package ui

import (
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/ui/spec"
)

func (m specFormModel[T]) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m specFormModel[T]) render() string {
	body := m.bodyLines()
	start, end := m.body.Range(len(body))
	lines := append(m.headLines(scrollNote(start, len(body)-end)), body[start:end]...)
	return m.size.join(append(lines, m.tailLines()...))
}

func (m specFormModel[T]) headLines(scroll string) []string {
	return []string{
		m.topLine(scroll),
		m.badgeLine(),
		frameLine(m.size.formInner(), nil, nil),
	}
}

func (m specFormModel[T]) bodyLines() []string {
	lines := make([]string, 0, 2*len(m.spec.Fields))
	for i := range m.spec.Fields {
		lines = append(lines, m.fieldRow(i), m.messageRow(i))
	}
	return lines
}

func (m specFormModel[T]) tailLines() []string {
	lines := []string{
		frameLine(m.size.formInner(), nil, nil),
		frameRule(m.size.formInner(), gTeeL, gTeeR),
		m.statusLine(),
	}
	if m.insert {
		lines = append(lines, keyhintBar(m.size.formInner(), m.keyhints())...)
	} else {
		lines = append(lines, keyhintLines(m.size.formInner(), keyhintFooter{groups: m.keyhints(), open: m.help})...)
	}
	return append(lines, frameBottom(m.size.formInner()))
}

func (m specFormModel[T]) topLine(scroll string) string {
	crumb := "new " + m.noun
	if m.isEdit {
		crumb = "edit " + m.noun
	}
	return frameTop(m.size.formInner(), crumb, topNote(scroll))
}

func (m specFormModel[T]) badgeLine() string {
	title := m.title
	if m.isEdit {
		if name := strings.TrimSpace(m.vals[m.spec.Fields[0].Key]); name != "" {
			title = "editing " + name
		}
	}
	return frameBadge(m.size.formInner(), m.badge, cAccent, title)
}

func (m specFormModel[T]) fieldRow(i int) string {
	f := m.spec.Fields[i]
	active := i == m.idx

	var bg color.Color
	if active {
		bg = cAccent
	}

	caret := caretBlank
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

	box := m.inputBox(i, active, bg)
	spans := make([]span, 0, 3+len(box))
	spans = append(spans,
		span{text: caret, fg: caretColor(active), bg: bg, bold: true},
		span{text: gutter + " ", fg: gc, bg: bg},
		span{text: truncPad(label, formLabelW, false), fg: labelFg, bg: bg, bold: active},
	)
	spans = append(spans, box...)
	return frameLine(m.size.formInner(), spans, bg)
}

func (m specFormModel[T]) inputBox(i int, active bool, bg color.Color) []span {
	borderFg := cBorder
	if active {
		borderFg = cInvFg
	}
	if m.attempted && m.fieldError(i) != "" {
		borderFg = cRed
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
		{text: truncPad(disp, m.size.formBox()-2, false), fg: valFg, bg: bg},
		{text: gLineV, fg: borderFg, bg: bg},
	}
}

func (m specFormModel[T]) display(i int, active bool) (string, bool) {
	f := m.spec.Fields[i]
	raw := m.vals[f.Key]
	if raw != "" {
		if f.Kind == spec.HiddenFieldKind && !m.reveal {
			return strings.Repeat(gInputMask, len([]rune(raw))), false
		}
		return raw, false
	}
	switch {
	case f.Kind == spec.HiddenFieldKind && m.isEdit:
		return "unchanged", true
	case f.Kind == spec.HiddenFieldKind:
		return "already in the keyring", true
	case active:
		return f.Example, true
	case f.Property == spec.RequiredFieldProperty:
		return "", true
	default:
		return gEmpty, true
	}
}

func (m specFormModel[T]) gutter(i int, active bool) (string, color.Color) {
	f := m.spec.Fields[i]

	var g string
	var c color.Color
	switch {
	case m.attempted && m.fieldError(i) != "":
		g, c = gFieldBad, cRed
	case m.vals[f.Key] != "":
		g, c = gDotOn, cAccent
	default:
		g, c = gDotOff, cFaint
	}
	if active {
		c = cInvFg
	}
	return g, c
}

func (m specFormModel[T]) messageRow(i int) string {
	text := ""
	if m.attempted {
		if e := m.fieldError(i); e != "" {
			text = gFieldBad + " " + e
		}
	}
	prefix := strings.Repeat(" ", 5+formLabelW+1)
	return frameLine(m.size.formInner(), []span{
		{text: prefix, fg: cDim},
		{text: truncPad(text, m.size.formInner()-len([]rune(prefix))-2, false), fg: cRed},
	}, nil)
}

func (m specFormModel[T]) statusLine() string {
	icon, col := statusGlyph(m.statusKind)
	return frameStatus(m.size.formInner(), icon, col, m.status, cursorPos(m.idx, len(m.spec.Fields)))
}
