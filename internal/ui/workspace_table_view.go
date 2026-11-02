package ui

import (
	"fmt"
	"image/color"
	"strings"

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
	total := m.workspaces.Total()

	lines := []string{
		frameTop(workspaceInner, "workspaces", fmt.Sprintf(" %d %s "+gLineH, total, plural(total, "workspace", "workspaces"))),
		m.workspaceHeaderLine(),
		frameRule(workspaceInner, gTeeL, gTeeR),
	}

	if n == 0 {
		lines = append(lines, frameLine(workspaceInner, []span{
			{text: truncPad("   no workspace matches the filter", workspaceInner, false), fg: cFaint},
		}, nil))
	} else {
		for i := range n {
			lines = append(lines, m.workspaceRowLine(i))
		}
	}

	lines = append(lines, frameRule(workspaceInner, gTeeL, gTeeR))
	lines = append(lines, m.workspaceFooterLines(n)...)
	lines = append(lines, frameBottom(workspaceInner))
	return strings.Join(lines, "\n")
}

func (m workspaceModel) workspaceFooterLines(n int) []string {
	var lines []string
	if m.workspaces.Confirming() || m.statusKind != kindIdle {
		lines = append(lines, m.workspaceStatusLine(), frameRule(workspaceInner, gTeeL, gTeeR))
	}
	return append(lines, keyhintLines(workspaceInner, keyhintFooter{
		groups: m.keyhints(),
		open:   m.workspaces.Help(),
		search: m.workspaces.Searching() || m.workspaces.Filtered(),
		typing: m.workspaces.Searching(),
		query:  m.workspaces.Query(),
		pos:    cursorPos(m.workspaces.Cursor(), n),
	})...)
}

func (m workspaceModel) workspaceHeaderLine() string {
	return frameLine(workspaceInner, []span{
		{text: truncPad("", wMark, false), fg: cDim},
		{text: " " + truncPad("NAME", wWsName, false), fg: cDim},
		{text: " " + truncPad("STATUS", wWsStatus, false), fg: cDim},
	}, nil)
}

func (m workspaceModel) workspaceRowLine(i int) string {
	name, ok := m.workspaces.At(i)
	if !ok {
		return frameLine(workspaceInner, nil, nil)
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

	return frameLine(workspaceInner, []span{
		{text: caret + mark, fg: markC, bg: bg},
		{text: " " + truncPad(name, wWsName, false), fg: fg, bg: bg},
		{text: " " + truncPad(note, wWsStatus, false), fg: soft, bg: bg},
	}, bg)
}

func (m workspaceModel) workspaceStatusLine() string {
	icon, col := statusGlyph(m.statusKind)
	text := m.status
	if m.workspaces.Confirming() {
		icon, col, text = gStatusWarn, cAmber, fmt.Sprintf("delete workspace %q with its connections and secrets? y/n", m.cursorWorkspaceLabel())
	}
	return frameStatus(workspaceInner, icon, col, text, "")
}
