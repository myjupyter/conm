package ui

import (
	"fmt"
	"image/color"

	tea "charm.land/bubbletea/v2"
)

const workspaceInner = 60

const (
	wWsName   = 44
	wWsStatus = 9
)

func (m workspaceModel) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m workspaceModel) render() string {
	n := m.workspaces.Len()
	start, end := m.rows.Range(n)
	lines := m.headLines(scrollNote(start, n-end))

	if n == 0 {
		lines = append(lines, frameLine(m.inner(), []span{
			{text: truncPad("   no workspace matches the filter", m.inner(), false), fg: cFaint},
		}, nil))
	}
	for i := start; i < end; i++ {
		lines = append(lines, m.workspaceRowLine(i))
	}

	return m.size.join(append(lines, m.tailLines()...))
}

func (m workspaceModel) inner() int {
	return m.size.frameWidth(workspaceInner)
}

func (m workspaceModel) nameWidth() int {
	return wWsName + m.size.extra(workspaceInner)
}

func (m workspaceModel) headLines(scroll string) []string {
	total := m.workspaces.Total()
	return []string{
		frameTop(m.inner(), "workspaces", topNote(fmt.Sprintf("%d %s", total, plural(total, "workspace", "workspaces")), scroll)),
		m.workspaceHeaderLine(),
		frameRule(m.inner(), gTeeL, gTeeR),
	}
}

func (m workspaceModel) tailLines() []string {
	lines := []string{frameRule(m.inner(), gTeeL, gTeeR)}
	lines = append(lines, m.workspaceFooterLines(m.workspaces.Len())...)
	return append(lines, frameBottom(m.inner()))
}

func (m workspaceModel) workspaceFooterLines(n int) []string {
	var lines []string
	if m.workspaces.Confirming() || m.statusKind != kindIdle {
		lines = append(lines, m.workspaceStatusLine(), frameRule(m.inner(), gTeeL, gTeeR))
	}
	return append(lines, keyhintLines(m.inner(), keyhintFooter{
		groups: m.keyhints(),
		open:   m.workspaces.Help(),
		search: m.workspaces.Searching() || m.workspaces.Filtered(),
		typing: m.workspaces.Searching(),
		query:  m.workspaces.Query(),
		pos:    cursorPos(m.workspaces.Cursor(), n),
	})...)
}

func (m workspaceModel) workspaceHeaderLine() string {
	return frameLine(m.inner(), []span{
		{text: truncPad("", wMark, false), fg: cDim},
		{text: " " + truncPad("NAME", m.nameWidth(), false), fg: cDim},
		{text: " " + truncPad("STATUS", wWsStatus, false), fg: cDim},
	}, nil)
}

func (m workspaceModel) workspaceRowLine(i int) string {
	name, ok := m.workspaces.At(i)
	if !ok {
		return frameLine(m.inner(), nil, nil)
	}
	sel := i == m.workspaces.Cursor()
	current := name == m.workspaces.Current()

	var bg, fg, soft, markC color.Color
	switch {
	case sel:
		bg, fg, soft, markC = cAccent, cInvFg, cInvFg, cInvFg
	case current:
		fg, soft, markC = cFg, cSoft, cAccent
	default:
		fg, soft, markC = cFg, cSoft, cFaint
	}

	caret, mark, note := " ", gDotOff, ""
	if sel {
		caret = gCaret
	}
	if current {
		mark, note = gDotOn, "current"
	}

	return frameLine(m.inner(), []span{
		{text: caret + mark, fg: markC, bg: bg},
		{text: " " + truncPad(name, m.nameWidth(), false), fg: fg, bg: bg},
		{text: " " + truncPad(note, wWsStatus, false), fg: soft, bg: bg},
	}, bg)
}

func (m workspaceModel) workspaceStatusLine() string {
	icon, col := statusGlyph(m.statusKind)
	text := m.status
	if m.workspaces.Confirming() {
		icon, col, text = gStatusWarn, cAmber, fmt.Sprintf("delete workspace %q with its connections and secrets? y/n", m.cursorWorkspaceLabel())
	}
	return frameStatus(m.inner(), icon, col, text, "")
}
