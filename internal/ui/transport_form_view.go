package ui

import (
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/charmbracelet/x/ansi"

	"github.com/myjupyter/conm/internal/i18n"
	"github.com/myjupyter/conm/internal/ui/spec"
)

func (m transportFormModel) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m transportFormModel) render() string {
	crumb := i18n.Sprintf("new transport")
	if m.isEdit {
		crumb = i18n.Sprintf("edit transport")
	}

	lines := []string{
		frameTop(formInner, crumb, ""),
		frameBadge(formInner, i18n.Sprintf("transport"), cAccent, m.badgeTitle()),
	}

	at := 0
	for _, s := range m.spec.Sections {
		lines = append(lines, frameHead(formInner, s.Title, s.Note)...)
		for range s.Fields {
			lines = append(lines, m.transportFieldRow(at), m.transportMessageRow(m.order[at]))
			at++
		}
	}

	icon, col := statusGlyph(m.statusKind)
	lines = append(lines,
		frameLine(formInner, nil, nil),
		frameRule(formInner, gTeeL, gTeeR),
		frameStatus(formInner, icon, col, m.status, cursorPos(m.idx, len(m.order))),
	)
	if m.insert {
		lines = append(lines, keyhintBar(formInner, m.keyhints())...)
	} else {
		lines = append(lines, keyhintLines(formInner, keyhintFooter{groups: m.keyhints(), open: m.help})...)
	}
	lines = append(lines, frameBottom(formInner))
	return strings.Join(lines, "\n")
}

func (m transportFormModel) badgeTitle() string {
	if name := strings.TrimSpace(m.vals[spec.TransportNameKey]); m.isEdit && name != "" {
		return i18n.Sprintf("editing %s", name)
	}
	return m.title
}

func (m transportFormModel) transportFieldRow(at int) string {
	i := m.order[at]
	f := m.spec.Fields[i]
	active := at == m.idx

	var bg color.Color
	if active {
		bg = cAccent
	}

	caret := caretBlank
	if active {
		caret = " " + gCaret + " "
	}
	gutter, gc := m.transportGutter(i, active)

	label := strings.ToLower(f.Label)
	if f.Property == spec.RequiredFieldProperty {
		label += " *"
	}
	labelFg := cSoft
	if active {
		labelFg = cInvFg
	}

	box := m.transportInputBox(i, active, bg)
	spans := make([]span, 0, 3+len(box))
	spans = append(spans,
		span{text: caret, fg: caretColor(active), bg: bg, bold: true},
		span{text: gutter + " ", fg: gc, bg: bg},
		span{text: truncPad(label, formLabelW, false), fg: labelFg, bg: bg, bold: active},
	)
	spans = append(spans, box...)
	return frameLine(formInner, spans, bg)
}

func (m transportFormModel) transportInputBox(i int, active bool, bg color.Color) []span {
	f := m.spec.Fields[i]

	borderFg := cBorder
	if active {
		borderFg = cInvFg
	}
	if m.attempted && m.fieldError(i) != "" {
		borderFg = cRed
	}

	if isSelect(f) {
		return selectSpans(m.vals[f.Key], m.options(f), m.isMissing(f), active, borderFg, bg)
	}

	disp, ghost := m.vals[f.Key], false
	if disp == "" {
		disp, ghost = m.placeholder(f, active), true
	}
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

func (m transportFormModel) placeholder(f spec.FormField, active bool) string {
	switch {
	case f.DefaultValue != "":
		return i18n.Sprintf("%s (default)", f.DefaultValue)
	case active:
		return f.Example
	case f.Property == spec.RequiredFieldProperty:
		return ""
	default:
		return gEmpty
	}
}

func (m transportFormModel) transportGutter(i int, active bool) (string, color.Color) {
	f := m.spec.Fields[i]

	var g string
	var c color.Color
	switch {
	case m.attempted && m.fieldError(i) != "":
		g, c = gFieldBad, cRed
	case m.vals[f.Key] != "" || f.DefaultValue != "":
		g, c = gDotOn, cAccent
	default:
		g, c = gDotOff, cFaint
	}
	if active {
		c = cInvFg
	}
	return g, c
}

func (m transportFormModel) transportMessageRow(i int) string {
	text := ""
	if m.attempted {
		if e := m.fieldError(i); e != "" {
			text = gFieldBad + " " + e
		}
	}
	prefix := strings.Repeat(" ", 5+formLabelW+1)
	return frameLine(formInner, []span{
		{text: prefix, fg: cDim},
		{text: truncPad(text, formInner-ansi.StringWidth(prefix)-2, false), fg: cRed},
	}, nil)
}
