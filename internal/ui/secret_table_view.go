package ui

import (
	"fmt"
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// The secrets table is narrower than the connections table: four short columns
// instead of six, and no live state to show on the right.
const secretInner = 88

const (
	wSecID   = 18
	wSecLoc  = 24
	wSecUsed = 16
	wSecDesc = 22
)

func (m secretModel) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m secretModel) render() string {
	n := m.repo.Len()

	lines := []string{
		frameTop(secretInner, m.crumb(), fmt.Sprintf(" %d %s "+gLineH, n, plural(n, "entry", "entries"))),
		m.storeTabsLine(n),
		frameRule(secretInner, gTeeL, gTeeR),
		m.secretHeaderLine(),
		frameRule(secretInner, gTeeL, gTeeR),
	}

	if n == 0 {
		lines = append(lines, frameLine(secretInner, []span{
			{text: truncPad("   keyring is empty · press "+keyMap.Add.hint+" to add an entry", secretInner, false), fg: cFaint},
		}, nil))
	} else {
		for i := range n {
			lines = append(lines, m.secretRowLine(i))
		}
	}

	lines = append(lines,
		frameRule(secretInner, gTeeL, gTeeR),
		m.secretStatusLine(n),
	)
	lines = append(lines, keyhintLines(secretInner, m.keyhints(), m.help)...)
	lines = append(lines, frameBottom(secretInner))
	return strings.Join(lines, "\n")
}

func (m secretModel) crumb() string {
	if m.picking {
		return "keyring · pick an entry"
	}
	return "keyring"
}

func (m secretModel) storeTabsLine(n int) string {
	label := fmt.Sprintf(" %s %d ", m.repo.Kind(), n)
	return frameLine(secretInner, []span{
		{text: " ", fg: cDim},
		{text: label, fg: cInvFg, bg: cAccent, bold: true},
		{text: "  ", fg: cDim},
		{text: "tab to switch store", fg: cFaint},
	}, nil)
}

func (m secretModel) secretHeaderLine() string {
	return frameLine(secretInner, []span{
		{text: truncPad("", wMark, false), fg: cDim},
		{text: " " + truncPad("ID", wSecID, false), fg: cDim},
		{text: " " + truncPad("LOCATION", wSecLoc, false), fg: cDim},
		{text: " " + truncPad("USED BY", wSecUsed, false), fg: cDim},
		{text: " " + truncPad("DESCRIPTION", wSecDesc, false), fg: cDim},
	}, nil)
}

func (m secretModel) secretRowLine(i int) string {
	s, ok := m.repo.Get(i)
	if !ok {
		return frameLine(secretInner, nil, nil)
	}
	sel := i == m.cursor
	used := len(m.repo.UsagesAt(i))

	var bg, fg, soft, markC color.Color
	switch {
	case sel:
		bg, fg, soft, markC = cAccent, cInvFg, cInvFg, cInvFg
	case !s.IsValid():
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
	case !s.IsValid():
		mark = gDotFail
	case used > 0:
		mark = gDotOn
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

	return frameLine(secretInner, []span{
		{text: caret + mark, fg: markC, bg: bg},
		{text: " " + truncPad(s.ID(), wSecID, false), fg: fg, bg: bg},
		{text: " " + truncPad(s.Location(), wSecLoc, false), fg: soft, bg: bg},
		{text: " " + truncPad(usedText, wSecUsed, false), fg: usedFg, bg: bg},
		{text: " " + truncPad(s.Description(), wSecDesc, false), fg: soft, bg: bg},
	}, bg)
}

func (m secretModel) secretStatusLine(n int) string {
	icon, col := statusGlyph(m.statusKind)
	text := m.status
	if m.confirming {
		icon, col, text = gStatusWarn, cAmber, fmt.Sprintf("delete %q from the keyring? y/n", m.cursorSecretLabel())
	}
	return frameStatus(secretInner, icon, col, text, cursorPos(m.cursor, n))
}
