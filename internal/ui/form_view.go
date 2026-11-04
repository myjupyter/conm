package ui

import (
	"fmt"
	"image/color"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/ui/spec"
)

const (
	formInner  = 88
	formLabelW = 18
	formBoxW   = 46

	railW = 2

	commonHeadTitle = "common"
	linksHeadTitle  = "links"
)

func (m formModel) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m formModel) render() string {
	body, _ := m.bodyLines()
	start, end := m.body.Range(len(body))
	lines := append(m.headLines(scrollNote(start, len(body)-end)), body[start:end]...)
	return m.size.join(append(lines, m.tailLines()...))
}

func (m formModel) headLines(scroll string) []string {
	return []string{
		m.formTopLine(scroll),
		m.badgeLine(),
		frameLine(m.size.formInner(), nil, nil),
		m.tabLine(),
		m.dividerLine(),
	}
}

func (m formModel) bodyLines() (lines []string, focus focusSpan) {
	titled := m.hasLinks() && m.section == m.linkSect
	if titled {
		lines = append(lines, frameHead(m.size.formInner(), commonHeadTitle, "")...)
	}

	fields := m.sectionFields(m.section)
	for n, i := range fields {
		at := len(lines)
		if m.fields[i].Key == spec.LinkNameKey(0) {
			lines = append(lines, m.linkHeadLines()...)
		}
		if n == 0 {
			at = 0
		}
		if n == m.idx {
			focus = focusSpan{from: at, row: len(lines), to: len(lines) + 1}
		}
		lines = append(lines, m.fieldRow(i), m.messageRow(i))
	}
	if titled && m.linkCount() == 0 {
		lines = append(lines, m.linkHeadLines()...)
		if m.idx == len(fields)-1 {
			focus.to = len(lines) - 1
		}
	}
	return lines, focus
}

func (m formModel) tailLines() []string {
	lines := []string{frameLine(m.size.formInner(), nil, nil)}

	if m.pong != "" {
		lines = append(lines, frameRule(m.size.formInner(), gTeeL, gTeeR), m.pongLine())
	}
	if m.ping != nil {
		lines = append(lines, frameErrPanel(m.size.formInner(), m.ping)...)
	}

	lines = append(lines, frameRule(m.size.formInner(), gTeeL, gTeeR), m.formStatusLine())
	if m.insert {
		lines = append(lines, keyhintBar(m.size.formInner(), m.keyhints())...)
	} else {
		lines = append(lines, keyhintLines(m.size.formInner(), keyhintFooter{groups: m.keyhints(), open: m.help})...)
	}
	return append(lines, frameBottom(m.size.formInner()))
}

func (m formModel) formTopLine(scroll string) string {
	crumb := "new connection"
	if m.isEdit {
		crumb = "edit connection"
	}
	return frameTop(m.size.formInner(), crumb, topNote(scroll))
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
	return frameBadge(m.size.formInner(), m.kind.String(), typeColor(m.kind), title)
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
	return frameLine(m.size.formInner(), spans, nil)
}

func (m formModel) dividerLine() string {
	return frameLine(m.size.formInner(), []span{
		{text: "  ", fg: cDim},
		{text: strings.Repeat(gLineH, m.size.formInner()-4), fg: cBorder},
	}, nil)
}

func (m formModel) linkHeadLines() []string {
	note := ""
	if m.linkCount() == 0 {
		note = "nothing linked yet — " + keyMap.Form.AddItem.hint + " adds one"
	}

	return frameHead(m.size.formInner(), linksHeadTitle, note)
}

func (m formModel) fieldRow(i int) string {
	f := m.fields[i]
	active := i == m.currentField()
	pair, first, pairActive := m.linkPair(i)

	bg := m.rowBg(active, pair, pairActive)

	caret := caretBlank
	if active {
		caret = " " + gCaret + " "
	}
	gutter, gc := m.gutter(i, active)

	label := m.label(f)
	if f.Property == spec.RequiredFieldProperty {
		label += " *"
	}
	labelFg := cSoft
	if active {
		labelFg = cInvFg
	}

	box := m.inputBox(i, active, bg)
	labelW := formLabelW

	spans := make([]span, 0, 4+len(box))
	spans = append(spans, span{text: caret, fg: caretColor(active), bg: bg, bold: true})
	if pair {
		labelW -= railW
		spans = append(spans, span{text: railGlyph(first) + " ", fg: railColor(active, pairActive), bg: bg})
	}
	spans = append(spans,
		span{text: gutter + " ", fg: gc, bg: bg},
		span{text: truncPad(label, labelW, false), fg: labelFg, bg: bg, bold: active},
	)
	spans = append(spans, box...)
	return frameLine(m.size.formInner(), spans, bg)
}

func (m formModel) rowBg(active, pair, pairActive bool) color.Color {
	switch {
	case active:
		return cAccent
	case pair && pairActive:
		return cPairBg
	default:
		return nil
	}
}

func railGlyph(first bool) string {
	if first {
		return gCornerTL
	}
	return gCornerBL
}

func railColor(active, pairActive bool) color.Color {
	switch {
	case active:
		return cInvFg
	case pairActive:
		return cAccent
	default:
		return cRail
	}
}

func (m formModel) inputBox(i int, active bool, bg color.Color) []span {
	f := m.fields[i]

	borderFg := cBorder
	if active {
		borderFg = cInvFg
	}
	if m.attempted && m.fieldError(i) != "" {
		borderFg = cRed
	}

	if isSelect(f) && m.applies(f) {
		return m.selectBox(f, active, borderFg, bg)
	}
	if isArray(f) {
		return m.chipsBox(f, active, bg)
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

func (m formModel) chips(f spec.FormField, active bool) []chip {
	items := m.items(f)
	if active && m.insert {
		draft := m.chipDraft + gInputCursor
		if m.chipAt < len(items) {
			items[m.chipAt] = draft
		} else {
			items = append(items, draft)
		}
	}

	focus := -1
	switch {
	case active && m.insert:
		focus = m.chipAt
	case active:
		focus = m.chipAtCursor(f)
	}

	out := make([]chip, len(items))
	for j, item := range items {
		out[j] = chip{text: fitChip(item, active && m.insert && j == m.chipAt), focused: j == focus}
	}
	return out
}

func (m formModel) chipsBox(f spec.FormField, active bool, bg color.Color) []span {
	return frameChips(m.size.formBox(), m.chips(f, active), active, bg)
}

func (m formModel) selectBox(f spec.FormField, active bool, borderFg, bg color.Color) []span {
	return selectSpans(m.size.formBox(), m.vals[f.Key], m.options(f), m.isMissing(f), active, borderFg, bg)
}

func selectSpans(w int, value string, options []string, missing, active bool, borderFg, bg color.Color) []span {
	pos := fmt.Sprintf("-/%d", len(options))
	if !missing {
		pos = fmt.Sprintf("%d/%d", max(slices.Index(options, value), 0)+1, len(options))
	}
	if value == "" {
		value = gEmpty
	}
	mid := max(w-8-len([]rune(pos)), 0)
	value = truncPad(value, mid, false)
	left := max((mid-len([]rune(value)))/2, 0)
	right := max(mid-left-len([]rune(value)), 0)

	valFg, posFg := cValue, cMuted
	if missing {
		valFg = cFaint
	}
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

func (m formModel) display(i int, active bool) (string, bool) {
	f := m.fields[i]
	raw := m.vals[f.Key]

	if !m.applies(f) {
		return m.example(f), true
	}
	if m.isSecretField(i) && m.noSecret() {
		return "no password sent", true
	}

	if m.isSecretField(i) && m.isRef() {
		if raw == "" {
			return "no entry — " + keyMap.Cycle.hint + " to pick", true
		}
		return raw, false
	}

	if raw != "" {
		if f.Kind == spec.HiddenFieldKind && !m.reveal && m.isLiteral() {
			return strings.Repeat(gInputMask, len([]rune(raw))), false
		}
		return raw, false
	}
	switch {
	case f.DefaultValue != "":
		return f.DefaultValue + " (default)", true
	case active:
		return m.example(f), true
	case f.Property == spec.RequiredFieldProperty:
		return "", true
	default:
		return gEmpty, true
	}
}

func (m formModel) gutter(i int, active bool) (string, color.Color) {
	f := m.fields[i]
	raw := m.vals[f.Key]

	var g string
	var c color.Color
	switch {
	case m.attempted && m.fieldError(i) != "":
		g, c = gFieldBad, cRed
	case m.isSecretField(i) && m.noSecret():
		g, c = gDotOn, cAccent
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
	pair, _, pairActive := m.linkPair(i)
	bg := m.rowBg(false, pair, pairActive)

	prefix := strings.Repeat(" ", 5+formLabelW+1)
	return frameLine(m.size.formInner(), []span{
		{text: prefix, fg: cDim, bg: bg},
		{text: truncPad(text, m.size.formInner()-len([]rune(prefix))-2, false), fg: col, bg: bg},
	}, bg)
}

func (m formModel) pongLine() string {
	return framePong(m.size.formInner(), m.pong)
}

func (m formModel) formStatusLine() string {
	icon, col := statusGlyph(m.statusKind)
	return frameStatus(m.size.formInner(), icon, col, m.status, cursorPos(m.idx, len(m.sectionFields(m.section))))
}

const caretBlank = "   "

func caretColor(active bool) color.Color {
	if active {
		return cInvFg
	}
	return cAccent
}
