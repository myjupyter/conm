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

	lines := []string{
		frameTop(initInner, m.initCrumb(), fmt.Sprintf(" %d of %d available "+gLineH, m.availableCount(), n)),
		frameLine(initInner, []span{
			{text: "  ", fg: cDim},
			{text: m.initIntro(), fg: cFg},
		}, nil),
		frameLine(initInner, nil, nil),
		m.initHeaderLine(),
		frameRule(initInner, gTeeL, gTeeR),
	}

	for i := range n {
		lines = append(lines, m.initRowLine(i))
	}

	lines = append(lines, frameRule(initInner, gTeeL, gTeeR))
	lines = append(lines, m.initFooterLines(n)...)
	return strings.Join(append(lines, frameBottom(initInner)), "\n")
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
	return frameLine(initInner, []span{
		{text: truncPad("", wMark, false), fg: cDim},
		{text: " " + truncPad("DATABASE", wInitType, false), fg: cDim},
		{text: " " + truncPad("CLIENTS", wInitClient, false), fg: cDim},
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

	return frameLine(initInner, spans, bg)
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

	return append(spans, span{text: strings.Repeat(" ", max(wInitClient+1-used, 1)), bg: bg})
}

func (m initModel) initFooterLines(n int) []string {
	var lines []string
	if m.statusKind != kindIdle {
		lines = append(lines, m.initStatusLine(), frameRule(initInner, gTeeL, gTeeR))
	}
	return append(lines, keyhintLines(initInner, keyhintFooter{
		groups: m.keyhints(),
		open:   m.help,
		pos:    cursorPos(m.cursor, n),
	})...)
}

func (m initModel) initStatusLine() string {
	icon, col := statusGlyph(m.statusKind)
	return frameStatus(initInner, icon, col, m.status, "")
}
