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

func (m secretModel) secretBinds() []struct{ key, label string } {
	use := "info"
	if m.picking {
		use = "use"
	}
	return []struct{ key, label string }{
		{"↑/k", "up"}, {"↓/j", "down"}, {"enter", use},
		{"n", "new"}, {"e", "edit"}, {"d", "del"},
		{"esc", "back"},
	}
}

func (m secretModel) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m secretModel) render() string {
	n := m.repo.Len()

	lines := []string{
		topLineW(secretInner, m.crumb(), fmt.Sprintf(" %d %s ─", n, plural(n, "entry", "entries"))),
		m.storeTabsLine(n),
		ruleW(secretInner, "├", "┤"),
		m.secretHeaderLine(),
		ruleW(secretInner, "├", "┤"),
	}

	if n == 0 {
		lines = append(lines, boxLineW(secretInner, []span{
			{text: truncPad("   keyring is empty · press n to add an entry", secretInner, false), fg: cFaint},
		}, nil))
	} else {
		for i := range n {
			lines = append(lines, m.secretRowLine(i))
		}
	}

	lines = append(lines,
		ruleW(secretInner, "├", "┤"),
		m.secretStatusLine(n),
	)
	lines = append(lines, keybarLinesW(secretInner, m.secretBinds())...)
	lines = append(lines, border("└")+border(strings.Repeat("─", secretInner))+border("┘"))
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
	return boxLineW(secretInner, []span{
		{text: " ", fg: cDim},
		{text: label, fg: cInvFg, bg: cAccent, bold: true},
		{text: "  ", fg: cDim},
		{text: "tab to switch store", fg: cFaint},
	}, nil)
}

func (m secretModel) secretHeaderLine() string {
	return boxLineW(secretInner, []span{
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
		return boxLineW(secretInner, nil, nil)
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
		caret = "❯"
	}
	mark := "○"
	switch {
	case !s.IsValid():
		mark = "✕"
	case used > 0:
		mark = "●"
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

	return boxLineW(secretInner, []span{
		{text: caret + mark, fg: markC, bg: bg},
		{text: " " + truncPad(s.ID(), wSecID, false), fg: fg, bg: bg},
		{text: " " + truncPad(s.Location(), wSecLoc, false), fg: soft, bg: bg},
		{text: " " + truncPad(usedText, wSecUsed, false), fg: usedFg, bg: bg},
		{text: " " + truncPad(s.Description(), wSecDesc, false), fg: soft, bg: bg},
	}, bg)
}

func (m secretModel) secretStatusLine(n int) string {
	icon, col, text := "›", cDim, m.status
	if m.confirming {
		icon, col, text = "!", cAmber, fmt.Sprintf("delete %q from the keyring? y/n", m.cursorSecretLabel())
	} else {
		icon, col = statusGlyph(m.statusKind)
	}

	pos := "0/0"
	if n > 0 {
		pos = fmt.Sprintf("%d/%d", m.cursor+1, n)
	}
	textW := max(secretInner-4-len([]rune(pos)), 0)

	return boxLineW(secretInner, []span{
		{text: " ", fg: cDim},
		{text: icon, fg: col},
		{text: " " + truncPad(text, textW, false), fg: col},
		{text: pos + " ", fg: cMuted},
	}, nil)
}
