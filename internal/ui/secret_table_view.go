package ui

import (
	"fmt"
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/i18n"
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
	n := m.secrets.Len()
	total := m.secrets.CountOf(m.secrets.Active())

	lines := []string{
		frameTop(secretInner, m.crumb(), " "+i18n.Sprintf("%d entries", total)+" "+gLineH),
		m.storeTabsLine(total),
		frameRule(secretInner, gTeeL, gTeeR),
		m.secretHeaderLine(),
		frameRule(secretInner, gTeeL, gTeeR),
	}

	if n == 0 {
		lines = append(lines, frameLine(secretInner, []span{
			{text: truncPad(m.emptySecretText(), secretInner, false), fg: cFaint},
		}, nil))
	} else {
		for i := range n {
			lines = append(lines, m.secretRowLine(i))
		}
	}

	lines = append(lines, frameRule(secretInner, gTeeL, gTeeR))
	lines = append(lines, m.secretFooterLines(n)...)
	lines = append(lines, frameBottom(secretInner))
	return strings.Join(lines, "\n")
}

func (m secretModel) secretFooterLines(n int) []string {
	var lines []string
	if m.secrets.Confirming() || m.statusKind != kindIdle {
		lines = append(lines, m.secretStatusLine(), frameRule(secretInner, gTeeL, gTeeR))
	}
	return append(lines, keyhintLines(secretInner, keyhintFooter{
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
		return "   " + i18n.Sprintf("no entry matches the filter")
	}
	return "   " + i18n.Sprintf("%s is empty · press %s to add an entry", m.secrets.Active(), keyMap.Add.hint)
}

func (m secretModel) crumb() string {
	if m.secrets.Picking() {
		return i18n.Sprintf("keyring · pick an entry")
	}
	return "keyring"
}

func (m secretModel) storeTabsLine(n int) string {
	label := fmt.Sprintf(" %s %d ", m.secrets.Active(), n)
	return frameLine(secretInner, []span{
		{text: " ", fg: cDim},
		{text: label, fg: cInvFg, bg: cAccent, bold: true},
		{text: "  ", fg: cDim},
		{text: i18n.Sprintf("tab to switch store"), fg: cFaint},
	}, nil)
}

func (m secretModel) secretHeaderLine() string {
	return frameLine(secretInner, []span{
		{text: truncPad("", wMark, false), fg: cDim},
		{text: " " + truncPad("ID", wSecID, false), fg: cDim},
		{text: " " + truncPad(i18n.Sprintf("LOCATION"), wSecLoc, false), fg: cDim},
		{text: " " + truncPad(i18n.Sprintf("USED BY"), wSecUsed, false), fg: cDim},
		{text: " " + truncPad(i18n.Sprintf("DESCRIPTION"), wSecDesc, false), fg: cDim},
	}, nil)
}

func (m secretModel) secretRowLine(i int) string {
	s, ok := m.secrets.At(i)
	if !ok {
		return frameLine(secretInner, nil, nil)
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

	usedText := i18n.Sprintf("unused")
	usedFg := soft
	if used > 0 {
		usedText = i18n.Sprintf("%d connections", used)
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

func (m secretModel) secretStatusLine() string {
	icon, col := statusGlyph(m.statusKind)
	text := m.status
	if m.secrets.Confirming() {
		icon, col, text = gStatusWarn, cAmber, i18n.Sprintf("delete %q from the keyring? y/n", m.cursorSecretLabel())
	}
	return frameStatus(secretInner, icon, col, text, "")
}
