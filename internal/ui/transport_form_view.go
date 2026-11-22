package ui

import (
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/ui/spec"
)

func (m transportFormModel) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m transportFormModel) render() string {
	body, _ := m.bodyLines()
	start, end := m.body.Range(len(body))
	lines := append(m.headLines(scrollNote(start, len(body)-end)), body[start:end]...)
	return m.size.join(append(lines, m.tailLines()...))
}

func (m transportFormModel) headLines(scroll string) []string {
	crumb := "new transport"
	if m.isEdit {
		crumb = "edit transport"
	}

	return []string{
		frameTop(m.size.formInner(), crumb, topNote(scroll)),
		frameBadge(m.size.formInner(), "transport", cAccent, m.badgeTitle()),
	}
}

func (m transportFormModel) bodyLines() (lines []string, focus focusSpan) {
	at := 0
	for _, s := range m.spec.Sections {
		head := len(lines)
		lines = append(lines, frameHead(m.size.formInner(), s.Title, "")...)
		for j := range s.Fields {
			if at == m.idx {
				focus = focusSpan{from: len(lines), row: len(lines), to: len(lines) + 1}
				if j == 0 {
					focus.from = head
				}
			}
			lines = append(lines, m.transportFieldRow(at), m.transportMessageRow(m.order[at]))
			at++
		}
	}
	return lines, focus
}

func (m transportFormModel) tailLines() []string {
	icon, col := statusGlyph(m.statusKind)
	lines := []string{
		frameLine(m.size.formInner(), nil, nil),
		frameRule(m.size.formInner(), gTeeL, gTeeR),
		frameStatus(m.size.formInner(), icon, col, m.status, cursorPos(m.idx, len(m.order))),
	}
	if m.insert {
		lines = append(lines, keyhintBar(m.size.formInner(), m.keyhints())...)
	} else {
		lines = append(lines, keyhintLines(m.size.formInner(), keyhintFooter{groups: m.keyhints(), open: m.help})...)
	}
	return append(lines, frameBottom(m.size.formInner()))
}

func (m transportFormModel) badgeTitle() string {
	if name := strings.TrimSpace(m.vals[spec.TransportNameKey]); m.isEdit && name != "" {
		return "editing " + name
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

	caret := caretBlank()
	if active {
		caret = " " + cell(gCaret) + " "
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
		span{text: caret, fg: caretColor(active), bg: bg, style: tSelection},
		span{text: cell(gutter) + " ", fg: gc, bg: bg},
		span{text: truncPad(label, formLabelW, false), fg: labelFg, bg: bg, style: styleIf(active, tSelection)},
	)
	spans = append(spans, box...)
	return frameLine(m.size.formInner(), spans, bg)
}

func (m transportFormModel) transportInputBox(i int, active bool, bg color.Color) []span {
	f := m.spec.Fields[i]

	edges := edgesFor(active, m.attempted && m.fieldError(i) != "")

	if isSelect(f) {
		return selectSpans(m.size.formBox(), m.vals[f.Key], m.options(f), m.isMissing(f), active, edges, bg)
	}

	disp, ghost := m.vals[f.Key], false
	if disp == "" {
		disp, ghost = m.placeholder(f, active), true
	}
	if active && m.insert {
		disp += gInputCursor
	}
	valFg := cInput
	if ghost {
		valFg = cFaint
	}
	if active {
		valFg = cInvFg
	}
	return []span{
		{text: gLineLeft, fg: edges.left, bg: bg},
		{text: truncPad(disp, m.size.formBox()-2, false), fg: valFg, bg: bg},
		{text: gLineRight, fg: edges.right, bg: bg},
	}
}

func (m transportFormModel) placeholder(f spec.FormField, active bool) string {
	switch {
	case f.DefaultValue != "":
		return f.DefaultValue + " (default)"
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
		g, c = gFieldBad, tinted("field_bad", cRed)
	case m.vals[f.Key] != "" || f.DefaultValue != "":
		g, c = gDotOn, tinted(elDotOn, cOK)
	default:
		g, c = gDotOff, tinted(elDotOff, cFaint)
	}
	if active {
		c = cInvFg
	}
	return g, c
}

func (m transportFormModel) transportMessageRow(i int) string {
	if !m.attempted {
		return frameFieldError(m.size.formInner(), "", nil)
	}
	return frameFieldError(m.size.formInner(), m.fieldError(i), nil)
}
