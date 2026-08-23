package ui

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// The connections table is the widest screen: six columns plus the live ping
// state on the right.
const tableInner = 108

const (
	wMark = 2
	wName = 17
	wUser = 14
	wHost = 39
	wPort = 6
	wDB   = 15
	wPing = 9
)

func (m Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m Model) render() string {
	n := m.conns.Len()

	lines := []string{
		frameTop(tableInner, "connections", fmt.Sprintf(" %d %s "+gLineH, n, plural(n, "connection", "connections"))),
		m.tabsLine(),
		frameRule(tableInner, gTeeL, gTeeR),
		m.headerLine(),
		frameRule(tableInner, gTeeL, gTeeR),
	}

	if n == 0 {
		lines = append(lines, frameLine(tableInner, []span{
			{text: truncPad("  no connections · press "+keyMap.Add.hint+" to add one", tableInner, false), fg: cDim},
		}, nil))
	} else {
		for i := range n {
			lines = append(lines, m.rowLine(i))
		}
	}

	if e := m.currentErr(); e != nil {
		lines = append(lines, frameErrPanel(tableInner, e)...)
	} else if p := m.currentPong(); p != "" {
		lines = append(lines, frameRule(tableInner, gTeeL, gTeeR), framePong(tableInner, p))
	}

	lines = append(lines, frameRule(tableInner, gTeeL, gTeeR), m.statusLine(n))
	lines = append(lines, keyhintLines(tableInner, m.keyhints(), m.conns.Help())...)
	lines = append(lines, frameBottom(tableInner))
	return strings.Join(lines, "\n")
}

func (m Model) tabsLine() string {
	spans := []span{{text: " ", fg: cDim}}
	for _, kind := range m.conns.Kinds() {
		label := fmt.Sprintf(" %s %d ", kind, m.conns.CountOf(kind))
		if kind == m.conns.Active() {
			spans = append(spans, span{text: label, fg: cInvFg, bg: typeColor(kind), bold: true})
		} else {
			spans = append(spans, span{text: label, fg: typeColor(kind)})
		}
		spans = append(spans, span{text: " ", fg: cDim})
	}
	return frameLine(tableInner, spans, nil)
}

func (m Model) headerLine() string {
	return frameLine(tableInner, []span{
		{text: truncPad("", wMark, false), fg: cDim},
		{text: " " + truncPad("NAME", wName, false), fg: cDim},
		{text: " " + truncPad("USERNAME", wUser, false), fg: cDim},
		{text: " " + truncPad("HOST", wHost, false), fg: cDim},
		{text: " " + truncPad("PORT", wPort, true), fg: cDim},
		{text: " " + truncPad("DBNAME", wDB, false), fg: cDim},
		{text: " " + truncPad("PING", wPing, true), fg: cDim},
	}, nil)
}

func (m Model) rowLine(i int) string {
	c, ok := m.conns.ConnectionAt(i)
	st := m.pingAt(i)
	if !ok || st == nil {
		return frameLine(tableInner, nil, nil)
	}
	sel := i == m.conns.Cursor()
	failed := st.connErr != nil || st.pingStatus == pingFailed || !c.IsValid()

	var bg, fg, soft, markC color.Color
	switch {
	case sel:
		bg, fg, soft, markC = cAccent, cInvFg, cInvFg, cInvFg
	case failed:
		fg, soft, markC = cDead, cSoft, cRed
	default:
		fg, soft, markC = cFg, cSoft, cFaint
	}

	caret := " "
	if sel {
		caret = gCaret
	}
	mark := gDotOff
	if failed {
		mark = gDotFail
	}

	spans := []span{
		{text: caret + mark, fg: markC, bg: bg},
		{text: " " + truncPad(c.Name(), wName, false), fg: fg, bg: bg},
		{text: " " + truncPad(c.Username(), wUser, false), fg: soft, bg: bg},
		{text: " " + truncPad(c.Host(), wHost, false), fg: fg, bg: bg},
		{text: " " + truncPad(strconv.Itoa(c.Port()), wPort, true), fg: soft, bg: bg},
		{text: " " + truncPad(c.Database(), wDB, false), fg: fg, bg: bg},
		m.pingSpan(i, bg, sel),
	}
	return frameLine(tableInner, spans, bg)
}

func (m Model) pingSpan(i int, bg color.Color, sel bool) span {
	st := m.pingAt(i)
	if st == nil {
		return span{text: " " + truncPad(gEmpty, wPing, true), fg: cFaint, bg: bg}
	}
	if st.pingStatus == pingPinging {
		raw := " " + strings.Repeat(" ", wPing-1) + st.pingSpinner.View()
		return span{text: strings.Repeat(" ", wPing+1), raw: raw}
	}

	text, fg := gEmpty, cFaint
	switch {
	case st.connErr != nil:
		text, fg = st.connErr.code, cRed
	case st.pingStatus == pingFailed:
		text, fg = "timeout", cRed
	case st.pingStatus == pingOK:
		ms := st.pingResult.PingTime.Milliseconds()
		text = fmt.Sprintf("%dms", ms)
		fg = cAmber
		if ms < 40 {
			fg = cAccent
		}
	}
	if sel {
		fg = cInvFg
	}
	return span{text: " " + truncPad(text, wPing, true), fg: fg, bg: bg}
}

func (m Model) statusLine(n int) string {
	icon, col := statusGlyph(m.statusKind)
	text := m.status
	if m.conns.Confirming() {
		icon, col, text = gStatusWarn, cAmber, fmt.Sprintf("delete %q? y/n", m.cursorLabel())
	}
	return frameStatus(tableInner, icon, col, text, cursorPos(m.conns.Cursor(), n))
}
