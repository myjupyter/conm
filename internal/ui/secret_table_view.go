package ui

import (
	"fmt"
	"image/color"

	tea "charm.land/bubbletea/v2"
)

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
	n := m.secrets.Len()
	start, end := m.rows.Range(n)
	lines := m.headLines(scrollNote(start, n-end))

	if n == 0 {
		lines = append(lines, frameLine(m.inner(), []span{
			{text: truncPad(m.emptySecretText(), m.inner(), false), fg: cFaint},
		}, nil))
	}
	for i := start; i < end; i++ {
		lines = append(lines, m.secretRowLine(i))
	}

	return m.size.join(append(lines, m.tailLines()...))
}

func (m secretModel) inner() int {
	return m.size.frameWidth(secretInner)
}

func (m secretModel) descWidth() int {
	return wSecDesc + m.size.extra(secretInner)
}

func (m secretModel) headLines(scroll string) []string {
	total := m.secrets.CountOf(m.secrets.Active())
	return []string{
		frameTop(m.inner(), m.crumb(), topNote(fmt.Sprintf("%d %s", total, plural(total, "entry", "entries")), scroll)),
		m.storeTabsLine(total),
		frameRule(m.inner(), gTeeL, gTeeR),
		m.secretHeaderLine(),
		frameRule(m.inner(), gTeeL, gTeeR),
	}
}

func (m secretModel) tailLines() []string {
	lines := []string{frameRule(m.inner(), gTeeL, gTeeR)}
	lines = append(lines, m.secretFooterLines(m.secrets.Len())...)
	return append(lines, frameBottom(m.inner()))
}

func (m secretModel) secretFooterLines(n int) []string {
	var lines []string
	if m.secrets.Confirming() || m.statusKind != kindIdle {
		lines = append(lines, m.secretStatusLine(), frameRule(m.inner(), gTeeL, gTeeR))
	}
	return append(lines, keyhintLines(m.inner(), keyhintFooter{
		groups: m.keyhints(),
		open:   m.secrets.Help(),
		search: m.secrets.Searching() || m.secrets.Filtered(),
		typing: m.secrets.Searching(),
		query:  m.secrets.Query(),
		pos:    cursorPos(m.secrets.Cursor(), n),
	})...)
}

func (m secretModel) emptySecretText() string {
	if m.secrets.Filtered() {
		return "   no entry matches the filter"
	}
	return "   " + m.secrets.Active() + " is empty · press " + keyMap.Secret.Add.hint + " to add an entry"
}

func (m secretModel) crumb() string {
	if m.secrets.Picking() {
		return "keyring · pick an entry"
	}
	return "keyring"
}

func (m secretModel) storeTabsLine(n int) string {
	label := fmt.Sprintf(" %s %d ", m.secrets.Active(), n)
	return frameLine(m.inner(), []span{
		{text: " ", fg: cDim},
		{text: label, fg: cInvFg, bg: cAccent, bold: true},
		{text: "  ", fg: cDim},
		{text: "tab to switch store", fg: cFaint},
	}, nil)
}

func (m secretModel) secretHeaderLine() string {
	return frameLine(m.inner(), []span{
		{text: truncPad("", wMark, false), fg: cDim},
		{text: " " + truncPad("ID", wSecID, false), fg: cDim},
		{text: " " + truncPad("LOCATION", wSecLoc, false), fg: cDim},
		{text: " " + truncPad("USED BY", wSecUsed, false), fg: cDim},
		{text: " " + truncPad("DESCRIPTION", m.descWidth(), false), fg: cDim},
	}, nil)
}

func (m secretModel) secretRowLine(i int) string {
	s, ok := m.secrets.At(i)
	if !ok {
		return frameLine(m.inner(), nil, nil)
	}
	sel := i == m.secrets.Cursor()
	used := len(m.secrets.UsagesAt(i))

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

	return frameLine(m.inner(), []span{
		{text: caret + mark, fg: markC, bg: bg},
		{text: " " + truncPad(s.ID(), wSecID, false), fg: fg, bg: bg},
		{text: " " + truncPad(s.Location(), wSecLoc, false), fg: soft, bg: bg},
		{text: " " + truncPad(usedText, wSecUsed, false), fg: usedFg, bg: bg},
		{text: " " + truncPad(s.Description(), m.descWidth(), false), fg: soft, bg: bg},
	}, bg)
}

func (m secretModel) secretStatusLine() string {
	icon, col := statusGlyph(m.statusKind)
	text := m.status
	if m.secrets.Confirming() {
		icon, col, text = gStatusWarn, cAmber, fmt.Sprintf("delete %q from the keyring? y/n", m.cursorSecretLabel())
	}
	return frameStatus(m.inner(), icon, col, text, "")
}
