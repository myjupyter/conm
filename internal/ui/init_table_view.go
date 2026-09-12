package ui

import (
	"fmt"
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui/spec"
)

// The setup table is as narrow as the forms: three columns and no live state.
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
	n := len(m.dbs)

	lines := []string{
		frameTop(initInner, m.initCrumb(), fmt.Sprintf(" %d of %d enabled "+gLineH, m.enabledCount(), n)),
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
	switch {
	case m.adding:
		return "add a database"
	case m.enabledCount() == 0:
		return "first run"
	default:
		return "databases"
	}
}

func (m initModel) initIntro() string {
	switch {
	case m.adding:
		return "add another database · " + keyMap.Toggle.hint + " to enable it"
	case m.enabledCount() == 0:
		return "no database enabled yet — pick the one you are connecting to"
	default:
		return "the databases conm manages · " + keyMap.Edit.hint + " to change one"
	}
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
	db := m.dbs[i]
	sel := i == m.cursor
	installed := m.installedClient(db.Type) != ""

	var bg, fg, markC, stateC color.Color
	switch {
	case sel:
		bg, fg, markC, stateC = cAccent, cInvFg, cInvFg, cInvFg
	case !installed:
		fg, markC, stateC = cDead, cFaint, cDim
	case db.Enabled:
		fg, markC, stateC = cFg, typeColor(db.Type), cAccent
	default:
		fg, markC, stateC = cFg, typeColor(db.Type), cDim
	}

	caret := " "
	if sel {
		caret = gCaret
	}

	clients := m.clientSpans(db, sel, bg)
	spans := make([]span, 0, 3+len(clients))
	spans = append(spans,
		span{text: caret + gDotOn, fg: markC, bg: bg},
		span{text: " " + truncPad(db.Type.String(), wInitType, false), fg: fg, bg: bg},
	)
	spans = append(spans, clients...)
	spans = append(spans, span{
		text: " " + truncPad(spec.DatabaseState(db.Enabled), wInitState, false),
		fg:   stateC, bg: bg, bold: db.Enabled,
	})

	return frameLine(initInner, spans, bg)
}

// clientSpans lists every client conm can drive this database with: the
// configured one accented, the ones found on this host readable, the rest
// muted — the whole list padded out to one column.
func (m initModel) clientSpans(db config.Database, sel bool, bg color.Color) []span {
	clients := m.clients[db.Type]

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
		case cli.Name == db.CLI:
			fg = cAccent
		case cli.Installed:
			fg = cFg
		}

		spans = append(spans, span{text: text, fg: fg, bg: bg, bold: cli.Name == db.CLI})
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
