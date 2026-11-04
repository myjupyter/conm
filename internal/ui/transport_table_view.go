package ui

import (
	"fmt"
	"image/color"
	"strconv"

	tea "charm.land/bubbletea/v2"
)

const transportInner = 88

const (
	wTrName = 18
	wTrKind = 5
	wTrVia  = 30
	wTrPort = 10
	wTrUsed = 14
)

func (m transportModel) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m transportModel) render() string {
	n := m.transports.Len()
	start, end := m.rows.Range(n)
	lines := m.headLines(scrollNote(start, n-end))

	if n == 0 {
		lines = append(lines, frameLine(m.inner(), []span{
			{text: truncPad(m.emptyTransportText(), m.inner(), false), fg: cFaint},
		}, nil))
	}
	for i := start; i < end; i++ {
		lines = append(lines, m.transportRowLine(i))
	}

	return m.size.join(append(lines, m.tailLines()...))
}

func (m transportModel) inner() int {
	return m.size.frameWidth(transportInner)
}

func (m transportModel) viaWidth() int {
	return wTrVia + m.size.extra(transportInner)
}

func (m transportModel) headLines(scroll string) []string {
	total := m.transports.CountOf(m.transports.Active())
	return []string{
		frameTop(m.inner(), "transports", topNote(fmt.Sprintf("%d %s", total, plural(total, "transport", "transports")), scroll)),
		m.kindTabsLine(total),
		frameRule(m.inner(), gTeeL, gTeeR),
		m.transportHeaderLine(),
		frameRule(m.inner(), gTeeL, gTeeR),
	}
}

func (m transportModel) tailLines() []string {
	lines := []string{frameRule(m.inner(), gTeeL, gTeeR)}
	lines = append(lines, m.transportFooterLines(m.transports.Len())...)
	return append(lines, frameBottom(m.inner()))
}

func (m transportModel) transportFooterLines(n int) []string {
	var lines []string
	if m.transports.Confirming() || m.statusKind != kindIdle {
		lines = append(lines, m.transportStatusLine(), frameRule(m.inner(), gTeeL, gTeeR))
	}
	return append(lines, keyhintLines(m.inner(), keyhintFooter{
		groups: m.keyhints(),
		open:   m.transports.Help(),
		search: m.transports.Searching() || m.transports.Filtered(),
		typing: m.transports.Searching(),
		query:  m.transports.Query(),
		pos:    cursorPos(m.transports.Cursor(), n),
	})...)
}

func (m transportModel) emptyTransportText() string {
	if m.transports.Filtered() {
		return "   no transport matches the filter"
	}
	return "   no transports yet · press " + keyMap.Transport.Add.hint + " to add one"
}

func (m transportModel) kindTabsLine(n int) string {
	label := fmt.Sprintf(" %s %d ", m.transports.Active(), n)
	return frameLine(m.inner(), []span{
		{text: " ", fg: cDim},
		{text: label, fg: cInvFg, bg: cAccent, bold: true},
	}, nil)
}

func (m transportModel) transportHeaderLine() string {
	return frameLine(m.inner(), []span{
		{text: truncPad("", wMark, false), fg: cDim},
		{text: " " + truncPad("NAME", wTrName, false), fg: cDim},
		{text: " " + truncPad("KIND", wTrKind, false), fg: cDim},
		{text: " " + truncPad("VIA", m.viaWidth(), false), fg: cDim},
		{text: " " + truncPad("LOCAL PORT", wTrPort, false), fg: cDim},
		{text: " " + truncPad("USED BY", wTrUsed, false), fg: cDim},
	}, nil)
}

func (m transportModel) transportRowLine(i int) string {
	t, ok := m.transports.At(i)
	if !ok {
		return frameLine(m.inner(), nil, nil)
	}
	sel := i == m.transports.Cursor()
	used := len(m.transports.UsagesAt(i))

	var bg, fg, soft, markC color.Color
	switch {
	case sel:
		bg, fg, soft, markC = cAccent, cInvFg, cInvFg, cInvFg
	case !t.IsValid():
		fg, soft, markC = cDead, cSoft, cRed
	case used > 0:
		fg, soft, markC = cFg, cSoft, cAccent
	default:
		fg, soft, markC = cFg, cSoft, cFaint
	}

	caret := " "
	if sel {
		caret = gCaret
	}
	mark := gDotOff
	switch {
	case !t.IsValid():
		mark = gDotFail
	case used > 0:
		mark = gDotOn
	}

	port := "auto"
	if t.LocalPort != 0 {
		port = strconv.Itoa(t.LocalPort)
	}

	usedText := "unused"
	usedFg := soft
	if used > 0 {
		usedText = fmt.Sprintf("%d %s", used, plural(used, "connection", "connections"))
		usedFg = cValue
		if sel {
			usedFg = cInvFg
		}
	}

	return frameLine(m.inner(), []span{
		{text: caret + mark, fg: markC, bg: bg},
		{text: " " + truncPad(t.Name, wTrName, false), fg: fg, bg: bg},
		{text: " " + truncPad(t.Kind, wTrKind, false), fg: soft, bg: bg},
		{text: " " + truncPad(m.viaText(t), m.viaWidth(), false), fg: soft, bg: bg},
		{text: " " + truncPad(port, wTrPort, false), fg: soft, bg: bg},
		{text: " " + truncPad(usedText, wTrUsed, false), fg: usedFg, bg: bg},
	}, bg)
}

func (m transportModel) transportStatusLine() string {
	icon, col := statusGlyph(m.statusKind)
	text := m.status
	if m.transports.Confirming() {
		icon, col, text = gStatusWarn, cAmber, fmt.Sprintf("delete transport %q? y/n", m.cursorTransportLabel())
	}
	return frameStatus(m.inner(), icon, col, text, "")
}
