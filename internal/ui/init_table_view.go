package ui

import (
	"fmt"
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
)

const initInner = 88

const (
	wInitType   = 14
	wInitClient = 44
	wInitState  = 12
)

func (m initModel) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m initModel) render() string {
	n := len(m.types)
	start, end := m.rows.Range(n)
	lines := m.headLines(scrollNote(start, n-end))

	for i := start; i < end; i++ {
		lines = append(lines, m.initRowLine(i))
	}

	return m.size.join(append(lines, m.tailLines()...))
}

func (m initModel) inner() int {
	return m.size.frameWidth(initInner)
}

func (m initModel) clientWidth() int {
	return wInitClient + m.size.extra(initInner)
}

func (m initModel) headLines(scroll string) []string {
	n := len(m.types)
	return []string{
		frameTop(m.inner(), m.initCrumb(), topNote(fmt.Sprintf("%d of %d available", m.availableCount(), n), scroll)),
		frameLine(m.inner(), []span{
			{text: "  ", fg: cDim},
			{text: m.initIntro(), fg: cFg},
		}, nil),
		frameLine(m.inner(), nil, nil),
		m.initHeaderLine(),
		frameRule(m.inner(), gTeeL, gTeeR),
	}
}

func (m initModel) tailLines() []string {
	lines := []string{frameRule(m.inner(), gTeeL, gTeeR)}
	lines = append(lines, m.initFooterLines(len(m.types))...)
	return append(lines, frameBottom(m.inner()))
}

func (m initModel) initCrumb() string {
	if m.adding {
		return "add a connection"
	}

	return "first run"
}

func (m initModel) initIntro() string {
	if m.adding {
		return "pick the type of the new connection"
	}

	return "no connections yet — pick the type of the first one"
}

func (m initModel) initHeaderLine() string {
	return frameLine(m.inner(), []span{
		{text: truncPad("", wMark, false), fg: cDim},
		{text: " " + truncPad("DATABASE", wInitType, false), fg: cDim},
		{text: " " + truncPad("CLIENTS", m.clientWidth(), false), fg: cDim},
		{text: " " + truncPad("STATE", wInitState, false), fg: cDim},
	}, nil)
}

func (m initModel) initRowLine(i int) string {
	t := m.types[i]
	sel := i == m.cursor
	available := m.available(t)

	var bg, fg, markC, stateC color.Color
	switch {
	case sel && available:
		bg, fg, markC, stateC = cAccent, cInvFg, cInvFg, cInvFg
	case sel:
		bg, fg, markC, stateC = cDead, cInvFg, cInvFg, cInvFg
	case !available:
		fg, markC, stateC = cDead, cFaint, cDim
	default:
		fg, markC, stateC = cFg, typeColor(t), cAccent
	}

	caret := " "
	if sel {
		caret = gCaret
	}

	state := "available"
	if !available {
		state = "no client"
	}

	clients := m.clientSpans(t, sel, bg)
	spans := make([]span, 0, 3+len(clients))
	spans = append(spans,
		span{text: caret + gDotOn, fg: markC, bg: bg},
		span{text: " " + truncPad(t.String(), wInitType, false), fg: fg, bg: bg},
	)
	spans = append(spans, clients...)
	spans = append(spans, span{
		text: " " + truncPad(state, wInitState, false),
		fg:   stateC, bg: bg, bold: available,
	})

	return frameLine(m.inner(), spans, bg)
}

func (m initModel) clientSpans(t config.ConnType, sel bool, bg color.Color) []span {
	clients := m.clients[t]

	spans := make([]span, 0, len(clients)+1)
	used := 0
	for i, cli := range clients {
		text := " " + cli.Name
		if i > 0 {
			text = ", " + cli.Name
		}

		fg := cMuted
		switch {
		case sel:
			fg = cInvFg
		case cli.Installed:
			fg = cFg
		}

		spans = append(spans, span{text: text, fg: fg, bg: bg})
		used += len([]rune(text))
	}

	return append(spans, span{text: strings.Repeat(" ", max(m.clientWidth()+1-used, 1)), bg: bg})
}

func (m initModel) initFooterLines(n int) []string {
	var lines []string
	if m.statusKind != kindIdle {
		lines = append(lines, m.initStatusLine(), frameRule(m.inner(), gTeeL, gTeeR))
	}
	return append(lines, keyhintLines(m.inner(), keyhintFooter{
		groups: m.keyhints(),
		open:   m.help,
		pos:    cursorPos(m.cursor, n),
	})...)
}

func (m initModel) initStatusLine() string {
	icon, col := statusGlyph(m.statusKind)
	return frameStatus(m.inner(), icon, col, m.status, "")
}
