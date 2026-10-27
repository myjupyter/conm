package ui

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

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
	total := m.transports.CountOf(m.transports.Active())

	lines := []string{
		frameTop(transportInner, "transports", fmt.Sprintf(" %d %s "+gLineH, total, plural(total, "transport", "transports"))),
		m.kindTabsLine(total),
		frameRule(transportInner, gTeeL, gTeeR),
		m.transportHeaderLine(),
		frameRule(transportInner, gTeeL, gTeeR),
	}

	if n == 0 {
		lines = append(lines, frameLine(transportInner, []span{
			{text: truncPad(m.emptyTransportText(), transportInner, false), fg: cFaint},
		}, nil))
	} else {
		for i := range n {
			lines = append(lines, m.transportRowLine(i))
		}
	}

	lines = append(lines, frameRule(transportInner, gTeeL, gTeeR))
	lines = append(lines, m.transportFooterLines(n)...)
	lines = append(lines, frameBottom(transportInner))
	return strings.Join(lines, "\n")
}

func (m transportModel) transportFooterLines(n int) []string {
	var lines []string
	if m.transports.Confirming() || m.statusKind != kindIdle {
		lines = append(lines, m.transportStatusLine(), frameRule(transportInner, gTeeL, gTeeR))
	}
	return append(lines, keyhintLines(transportInner, keyhintFooter{
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
	return frameLine(transportInner, []span{
		{text: " ", fg: cDim},
		{text: label, fg: cInvFg, bg: cAccent, bold: true},
	}, nil)
}

func (m transportModel) transportHeaderLine() string {
	return frameLine(transportInner, []span{
		{text: truncPad("", wMark, false), fg: cDim},
		{text: " " + truncPad("NAME", wTrName, false), fg: cDim},
		{text: " " + truncPad("KIND", wTrKind, false), fg: cDim},
		{text: " " + truncPad("VIA", wTrVia, false), fg: cDim},
		{text: " " + truncPad("LOCAL PORT", wTrPort, false), fg: cDim},
		{text: " " + truncPad("USED BY", wTrUsed, false), fg: cDim},
	}, nil)
}

func (m transportModel) transportRowLine(i int) string {
	t, ok := m.transports.At(i)
	if !ok {
		return frameLine(transportInner, nil, nil)
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

	return frameLine(transportInner, []span{
		{text: caret + mark, fg: markC, bg: bg},
		{text: " " + truncPad(t.Name, wTrName, false), fg: fg, bg: bg},
		{text: " " + truncPad(t.Kind, wTrKind, false), fg: soft, bg: bg},
		{text: " " + truncPad(m.viaText(t), wTrVia, false), fg: soft, bg: bg},
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
	return frameStatus(transportInner, icon, col, text, "")
}
