package ui

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/myjupyter/conm/internal/config"
)

func typeColor(t config.ConnType) color.Color {
	switch t {
	case config.PostgresConnType:
		return cPostgres
	default:
		return cAccent
	}
}

const inner = 108

const (
	wMark = 2
	wName = 17
	wUser = 14
	wHost = 39
	wPort = 6
	wDB   = 15
	wPing = 9
)

var (
	cBorder   = lipgloss.Color("#3a362e")
	cDim      = lipgloss.Color("#6d6759")
	cMuted    = lipgloss.Color("#5f594c")
	cFaint    = lipgloss.Color("#4d483e")
	cSoft     = lipgloss.Color("#8b8474")
	cFg       = lipgloss.Color("#ddd6c8")
	cDead     = lipgloss.Color("#7d7668")
	cAccent   = lipgloss.Color("#63d18a")
	cRed      = lipgloss.Color("#e5654a")
	cAmber    = lipgloss.Color("#e3b34a")
	cInvFg    = lipgloss.Color("#0f0f0d")
	cPostgres = lipgloss.Color("#5aa0d6")

	cErrBorder = lipgloss.Color("#7a3f34")
	cErrTagBg  = lipgloss.Color("#9e3b2a")
	cErrTagFg  = lipgloss.Color("#fdf3f1")
	cErrCode   = lipgloss.Color("#f0a48f")
	cErrMuted  = lipgloss.Color("#8d7a76")
	cErrValue  = lipgloss.Color("#c8b8b4")
	cHint      = lipgloss.Color("#a2938f")
)

var keybinds = []struct{ key, label string }{
	{"↑/k", "up"}, {"↓/j", "down"}, {"enter", "connect"},
	{"a", "add"}, {"e", "edit"}, {"d", "del"},
	{"p", "ping"}, {"q/esc", "quit"},
}

type span struct {
	text string
	fg   color.Color
	bg   color.Color
	bold bool
	raw  string
}

func (s span) render() string {
	if s.raw != "" {
		return s.raw
	}
	st := lipgloss.NewStyle().Foreground(s.fg)
	if s.bg != nil {
		st = st.Background(s.bg)
	}
	if s.bold {
		st = st.Bold(true)
	}
	return st.Render(s.text)
}

func spanWidth(s span) int { return len([]rune(s.text)) }

func border(s string) string {
	return lipgloss.NewStyle().Foreground(cBorder).Render(s)
}

func truncPad(s string, w int, right bool) string {
	if w <= 0 {
		return ""
	}
	r := []rune(s)
	switch {
	case len(r) > w:
		if w == 1 {
			return string(r[:1])
		}
		return string(r[:w-1]) + "…"
	case right:
		return strings.Repeat(" ", w-len(r)) + s
	default:
		return s + strings.Repeat(" ", w-len(r))
	}
}

func (m Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m Model) render() string {
	n := m.reg.Len()

	lines := []string{m.topLine(n), m.tabsLine(n), rule("├", "┤"), m.headerLine(), rule("├", "┤")}

	if n == 0 {
		lines = append(lines, boxLine([]span{
			{text: truncPad("  no connections · press a to add one", inner, false), fg: cDim},
		}, nil))
	} else {
		for i := range n {
			lines = append(lines, m.rowLine(i))
		}
	}

	if e := m.currentErr(); e != nil {
		lines = append(lines, m.errPanelLines(e)...)
	}

	lines = append(lines,
		rule("├", "┤"),
		m.statusLine(n),
		m.keybindLine(),
		border("└")+border(strings.Repeat("─", inner))+border("┘"),
	)
	return strings.Join(lines, "\n")
}

func boxLine(spans []span, bg color.Color) string {
	w := 0
	for _, s := range spans {
		w += spanWidth(s)
	}
	if w < inner {
		spans = append(spans, span{text: strings.Repeat(" ", inner-w), bg: bg})
	}
	var b strings.Builder
	b.WriteString(border("│"))
	for _, s := range spans {
		b.WriteString(s.render())
	}
	b.WriteString(border("│"))
	return b.String()
}

func rule(left, right string) string {
	return border(left) + border(strings.Repeat("─", inner)) + border(right)
}

func (m Model) topLine(n int) string {
	left := []span{
		{text: "─ ", fg: cBorder},
		{text: "conm", fg: cFg, bold: true},
		{text: " · connections ", fg: cDim},
	}
	right := fmt.Sprintf(" %d %s ─", n, plural(n, "connection", "connections"))

	used := len([]rune(right))
	for _, s := range left {
		used += spanWidth(s)
	}
	fill := max(inner-used, 0)

	var b strings.Builder
	b.WriteString(border("┌"))
	for _, s := range left {
		b.WriteString(s.render())
	}
	b.WriteString(border(strings.Repeat("─", fill)))
	b.WriteString((span{text: right, fg: cDim}).render())
	b.WriteString(border("┐"))
	return b.String()
}

func (m Model) tabsLine(n int) string {
	tabs := []struct {
		name   string
		count  int
		color  color.Color
		active bool
	}{
		{name: "postgres", count: n, color: typeColor(config.PostgresConnType), active: true},
	}

	spans := []span{{text: " ", fg: cDim}}
	for _, t := range tabs {
		label := fmt.Sprintf(" %s %d ", t.name, t.count)
		if t.active {
			spans = append(spans, span{text: label, fg: cInvFg, bg: t.color, bold: true})
		} else {
			spans = append(spans, span{text: label, fg: t.color})
		}
		spans = append(spans, span{text: " ", fg: cDim})
	}
	return boxLine(spans, nil)
}

func (m Model) headerLine() string {
	return boxLine([]span{
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
	c, ok := m.reg.Get(i)
	if !ok {
		return boxLine(nil, nil)
	}
	sel := i == m.cursor
	st := m.states[i]
	failed := st.connErr != nil || st.pingStatus == pingFailed || !c.IsValid()

	var bg, fg, soft, markC color.Color
	switch {
	case sel:
		bg, fg, soft, markC = typeColor(c.ConnType()), cInvFg, cInvFg, cInvFg
	case failed:
		fg, soft, markC = cDead, cSoft, cRed
	default:
		fg, soft, markC = cFg, cSoft, cFaint
	}

	caret := " "
	if sel {
		caret = "❯"
	}
	mark := "○"
	if failed {
		mark = "✕"
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
	return boxLine(spans, bg)
}

func (m Model) pingSpan(i int, bg color.Color, sel bool) span {
	st := m.states[i]
	if st.pingStatus == pingPinging {
		raw := " " + strings.Repeat(" ", wPing-1) + st.pingSpinner.View()
		return span{text: strings.Repeat(" ", wPing+1), raw: raw}
	}

	text, color := "—", cFaint
	switch {
	case st.connErr != nil:
		text, color = st.connErr.code, cRed
	case st.pingStatus == pingFailed:
		text, color = "timeout", cRed
	case st.pingStatus == pingOK:
		ms := st.pingResult.PingTime.Milliseconds()
		text = fmt.Sprintf("%dms", ms)
		color = cAmber
		if ms < 40 {
			color = cAccent
		}
	}
	if sel {
		color = cInvFg
	}
	return span{text: " " + truncPad(text, wPing, true), fg: color, bg: bg}
}

func (m Model) errPanelLines(e *connError) []string {
	tag := " " + strings.ToUpper(e.action) + " FAILED "
	code := " " + e.code + " "
	right := " " + e.conn + " "
	fill := max(inner-1-len([]rune(tag))-len([]rune(code))-len([]rune(right)), 0)

	var top strings.Builder
	top.WriteString(border("├"))
	top.WriteString((span{text: "─", fg: cErrBorder}).render())
	top.WriteString((span{text: tag, fg: cErrTagFg, bg: cErrTagBg, bold: true}).render())
	top.WriteString((span{text: code, fg: cErrCode, bold: true}).render())
	top.WriteString((span{text: strings.Repeat("─", fill), fg: cErrBorder}).render())
	top.WriteString((span{text: right, fg: cErrMuted}).render())
	top.WriteString(border("┤"))

	lines := []string{top.String()}
	if e.target != "" {
		lines = append(lines, errKV("target", e.target, cErrValue))
	}
	if e.op != "" {
		lines = append(lines, errKV("during", e.op, cErrValue))
	}
	lines = append(lines, errKV("error", e.detail, cErrCode))
	if e.hint != "" {
		lines = append(lines, boxLine([]span{
			{text: "  ", fg: cDim},
			{text: "→ ", fg: cAmber},
			{text: truncPad(e.hint, inner-4, false), fg: cHint},
		}, nil))
	}
	lines = append(lines, boxLine([]span{
		{text: "  ", fg: cDim},
		{text: "r", fg: cFg, bold: true},
		{text: " retry · ", fg: cErrMuted},
		{text: "esc", fg: cFg, bold: true},
		{text: " dismiss", fg: cErrMuted},
	}, nil))
	return lines
}

func errKV(label, value string, valColor color.Color) string {
	return boxLine([]span{
		{text: " ", fg: cDim},
		{text: " " + truncPad(label, 8, false), fg: cErrMuted},
		{text: truncPad(value, inner-11, false), fg: valColor},
	}, nil)
}

func (m Model) statusLine(n int) string {
	icon, color, text := "›", cDim, m.status
	switch {
	case m.confirming:
		icon, color, text = "!", cAmber, fmt.Sprintf("delete %q? y/n", m.cursorLabel())
	default:
		icon, color = statusGlyph(m.statusKind)
	}

	pos := "0/0"
	if n > 0 {
		pos = fmt.Sprintf("%d/%d", m.cursor+1, n)
	}
	textW := max(inner-4-len([]rune(pos)), 0)

	return boxLine([]span{
		{text: " ", fg: cDim},
		{text: icon, fg: color},
		{text: " " + truncPad(text, textW, false), fg: color},
		{text: pos + " ", fg: cMuted},
	}, nil)
}

func statusGlyph(k statusKind) (string, color.Color) {
	switch k {
	case kindOK:
		return "✓", cAccent
	case kindPending:
		return "◐", cAmber
	case kindWarn:
		return "!", cAmber
	case kindErr:
		return "✗", cRed
	default:
		return "›", cDim
	}
}

func (m Model) keybindLine() string {
	spans := []span{{text: " ", fg: cDim}}
	for i, b := range keybinds {
		sep := " · "
		if i == len(keybinds)-1 {
			sep = ""
		}
		spans = append(spans,
			span{text: b.key, fg: cFg, bold: true},
			span{text: " " + b.label + sep, fg: cDim},
		)
	}
	return boxLine(spans, nil)
}
