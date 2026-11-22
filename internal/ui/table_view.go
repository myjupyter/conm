package ui

import (
	"fmt"
	"image/color"
	"slices"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
)

const tableInner = 108

const (
	wName = 17
	wUser = 14
	wHost = 39
	wPort = 6
	wDB   = 15
	wPing = 9
	wTags = 16
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
			spans = append(spans, span{text: label, fg: cInvFg, bg: tabBg(typeColor(kind)), style: tTab})
		} else {
			spans = append(spans, span{text: label, fg: typeColor(kind)})
		}
		spans = append(spans, span{text: " ", fg: cDim})
	}
	return frameLine(m.inner(), spans, nil)
}

var (
	columnWidths = map[string]int{
		config.NameColumn: wName, config.UsernameColumn: wUser, config.HostColumn: wHost, config.PortColumn: wPort, config.DatabaseColumn: wDB, config.PingColumn: wPing, config.TagsColumn: wTags,
	}
	columnTitles = map[string]string{
		config.NameColumn: "NAME", config.UsernameColumn: "USERNAME", config.HostColumn: "HOST", config.PortColumn: "PORT", config.DatabaseColumn: "DBNAME", config.PingColumn: "PING", config.TagsColumn: "TAGS",
	}
	rightAligned = []string{config.PortColumn, config.PingColumn}
)

func (m Model) columns() []string {
	return settingsOf(m.conns.Active()).VisibleColumns()
}

func (m Model) columnWidth(column string) int {
	return layoutWidth(m.columns(), column, m.size.extra(tableInner))
}

func layoutWidth(columns []string, column string, extra int) int {
	flexible := config.HostColumn
	if !slices.Contains(columns, flexible) {
		flexible = config.NameColumn
	}
	if column != flexible {
		return columnWidths[column]
	}
	width := extra - 1
	for _, c := range config.DefaultColumns {
		width += columnWidths[c] + 1
	}
	for _, c := range columns {
		if c != flexible {
			width -= columnWidths[c] + 1
		}
	}
	return width
}

func (m Model) columnSpan(column, text string, fg, bg color.Color) span {
	return span{text: " " + truncPad(text, m.columnWidth(column), slices.Contains(rightAligned, column)), fg: fg, bg: bg}
}

func (m Model) headerLine() string {
	columns := m.columns()
	spans := make([]span, 0, 1+len(columns))
	spans = append(spans, span{text: blankCell() + blankCell(), fg: cDim})
	for _, column := range columns {
		spans = append(spans, m.columnSpan(column, columnTitles[column], cColumn, nil))
	}
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
		bg, fg, soft, markC = typeColor(m.conns.Active()), cInvFg, cInvFg, cInvFg
	case failed:
		fg, soft, markC = cDead, cSoft, cRed
	case alive:
		fg, soft, markC = cFg, cSoft, cAccent
	default:
		fg, soft, markC = cFg, cSoft, cFaint
	}

	mark, markEl := gDotOff, elDotOff
	switch {
	case failed:
		mark, markEl = gDotFail, elDotFail
	case alive:
		mark, markEl = gDotOn, elDotOn
	}

	spans := []span{
		{text: caretOf(sel), fg: tinted("caret", markC), bg: bg},
		{text: cell(mark), fg: markFg(sel, markEl, markC), bg: bg},
	}
	for _, column := range m.columns() {
		switch column {
		case config.NameColumn:
			spans = append(spans, m.columnSpan(column, cfg.Meta().Name, fg, bg))
		case config.UsernameColumn:
			spans = append(spans, m.columnSpan(column, cfg.Username(), soft, bg))
		case config.HostColumn:
			spans = append(spans, m.columnSpan(column, cfg.Host(), fg, bg))
		case config.PortColumn:
			spans = append(spans, m.columnSpan(column, strconv.Itoa(cfg.Port()), soft, bg))
		case config.DatabaseColumn:
			database := ""
			if db, ok := cfg.(config.DBConnection); ok {
				database = db.Database()
			}
			spans = append(spans, m.columnSpan(column, database, fg, bg))
		case config.PingColumn:
			spans = append(spans, m.pingSpan(i, bg, sel))
		case config.TagsColumn:
			spans = append(spans, m.columnSpan(column, strings.Join(cfg.Meta().Tags, ","), soft, bg))
		}
	}

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
			fg = cOK
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
