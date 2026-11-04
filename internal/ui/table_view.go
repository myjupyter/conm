package ui

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
)

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
	tail := append(m.panelLines(), m.footerTail()...)

	n := m.conns.Len()
	start, end := m.size.visibleRows(&m.rows, len(m.headLines(""))+len(tail), m.conns.Cursor(), n)
	lines := m.headLines(scrollNote(start, n-end))

	if n == 0 {
		lines = append(lines, frameLine(m.inner(), []span{
			{text: truncPad(m.emptyText(), m.inner(), false), fg: cDim},
		}, nil))
	}
	for i := start; i < end; i++ {
		lines = append(lines, m.rowLine(i))
	}

	return m.size.join(append(lines, tail...))
}

func (m Model) inner() int {
	return m.size.frameWidth(tableInner)
}

func (m Model) headLines(scroll string) []string {
	total := m.conns.CountOf(m.conns.Active())
	return []string{
		frameTop(m.inner(), "connections", topNote(fmt.Sprintf("%d %s", total, plural(total, "connection", "connections")), scroll)),
		m.tabsLine(),
		frameRule(m.inner(), gTeeL, gTeeR),
		m.headerLine(),
		frameRule(m.inner(), gTeeL, gTeeR),
	}
}

func (m Model) panelLines() []string {
	if e := m.currentErr(); e != nil {
		return frameErrPanel(m.inner(), e)
	}
	if p := m.currentPong(); p != "" {
		return []string{frameRule(m.inner(), gTeeL, gTeeR), framePong(m.inner(), p)}
	}
	return nil
}

func (m Model) footerTail() []string {
	lines := []string{frameRule(m.inner(), gTeeL, gTeeR)}
	lines = append(lines, m.footerLines(m.conns.Len())...)
	return append(lines, frameBottom(m.inner()))
}

func (m Model) footerLines(n int) []string {
	var lines []string
	if m.conns.Confirming() || m.statusKind != kindIdle {
		lines = append(lines, m.statusLine(), frameRule(m.inner(), gTeeL, gTeeR))
	}
	return append(lines, keyhintLines(m.inner(), keyhintFooter{
		groups: m.keyhints(),
		open:   m.conns.Help(),
		search: m.conns.Searching() || m.conns.Filtered(),
		typing: m.conns.Searching(),
		query:  m.conns.Query(),
		pos:    cursorPos(m.conns.Cursor(), n),
	})...)
}

func (m Model) emptyText() string {
	if m.conns.Filtered() {
		return "  no connection matches the filter"
	}
	return "  no connections · press " + keyMap.Screen.AddConnection.hint + " to add one"
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
	return frameLine(m.inner(), spans, nil)
}

func (m Model) hasDatabaseColumn() bool {
	return m.conns.Active().Kind() == config.DatabaseConnKind
}

func (m Model) hostWidth() int {
	host := wHost + m.size.extra(tableInner)
	if m.hasDatabaseColumn() {
		return host
	}
	return host + 1 + wDB
}

func (m Model) headerLine() string {
	spans := []span{
		{text: truncPad("", wMark, false), fg: cDim},
		{text: " " + truncPad("NAME", wName, false), fg: cDim},
		{text: " " + truncPad("USERNAME", wUser, false), fg: cDim},
		{text: " " + truncPad("HOST", m.hostWidth(), false), fg: cDim},
		{text: " " + truncPad("PORT", wPort, true), fg: cDim},
	}
	if m.hasDatabaseColumn() {
		spans = append(spans, span{text: " " + truncPad("DBNAME", wDB, false), fg: cDim})
	}
	spans = append(spans, span{text: " " + truncPad("PING", wPing, true), fg: cDim})

	return frameLine(m.inner(), spans, nil)
}

func (m Model) rowLine(i int) string {
	cfg, ok := m.conns.ConfigAt(i)
	st := m.pingAt(i)
	if !ok || st == nil {
		return frameLine(m.inner(), nil, nil)
	}
	sel := i == m.conns.Cursor()
	failed := st.connErr != nil || st.pingStatus == pingFailed || !cfg.IsValid()
	alive := st.connErr == nil && st.pingStatus == pingOK

	var bg, fg, soft, markC color.Color
	switch {
	case sel:
		bg, fg, soft, markC = cAccent, cInvFg, cInvFg, cInvFg
	case failed:
		fg, soft, markC = cDead, cSoft, cRed
	case alive:
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
	case failed:
		mark = gDotFail
	case alive:
		mark = gDotOn
	}

	spans := []span{
		{text: caret + mark, fg: markC, bg: bg},
		{text: " " + truncPad(cfg.Meta().Name, wName, false), fg: fg, bg: bg},
		{text: " " + truncPad(cfg.Username(), wUser, false), fg: soft, bg: bg},
		{text: " " + truncPad(cfg.Host(), m.hostWidth(), false), fg: fg, bg: bg},
		{text: " " + truncPad(strconv.Itoa(cfg.Port()), wPort, true), fg: soft, bg: bg},
	}
	if m.hasDatabaseColumn() {
		database := ""
		if db, ok := cfg.(config.DBConnection); ok {
			database = db.Database()
		}
		spans = append(spans, span{text: " " + truncPad(database, wDB, false), fg: fg, bg: bg})
	}
	spans = append(spans, m.pingSpan(i, bg, sel))

	return frameLine(m.inner(), spans, bg)
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

func (m Model) statusLine() string {
	icon, col := statusGlyph(m.statusKind)
	text := m.status
	switch {
	case m.conns.Confirming():
		icon, col, text = gStatusWarn, cAmber, fmt.Sprintf("delete %q? y/n", m.cursorLabel())
	case m.statusKind == kindPing:
		return framePing(m.inner(), text)
	}
	return frameStatus(m.inner(), icon, col, text, "")
}
