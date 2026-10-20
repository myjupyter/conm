package ui

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/i18n"
)

const transportInner = 88

const (
	wTrName = 18
	wTrKind = 5
	wTrVia  = 30
	wTrPort = 10
	wTrUsed = 14
)

func (m transportModel) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m transportModel) render() string {
	n := m.transports.Len()
	total := m.transports.CountOf(m.transports.Active())

	lines := []string{
		frameTop(transportInner, i18n.Sprintf("transports"), " "+i18n.Sprintf("%d transports", total)+" "+gLineH),
		m.kindTabsLine(total),
		frameRule(transportInner, gTeeL, gTeeR),
		m.transportHeaderLine(),
		frameRule(transportInner, gTeeL, gTeeR),
	}

	if n == 0 {
		lines = append(lines, frameLine(transportInner, []span{
			{text: truncPad(m.emptyTransportText(), transportInner, false), fg: cFaint},
		}, nil))
	} else {
		for i := range n {
			lines = append(lines, m.transportRowLine(i))
		}
	}

	lines = append(lines, frameRule(transportInner, gTeeL, gTeeR))
	lines = append(lines, m.transportFooterLines(n)...)
	lines = append(lines, frameBottom(transportInner))
	return strings.Join(lines, "\n")
}

func (m transportModel) transportFooterLines(n int) []string {
	var lines []string
	if m.transports.Confirming() || m.statusKind != kindIdle {
		lines = append(lines, m.transportStatusLine(), frameRule(transportInner, gTeeL, gTeeR))
	}
	return append(lines, keyhintLines(transportInner, keyhintFooter{
		groups: m.keyhints(),
		open:   m.transports.Help(),
		search: m.transports.Searching() || m.transports.Filtered(),
		typing: m.transports.Searching(),
		query:  m.transports.Query(),
		pos:    cursorPos(m.transports.Cursor(), n),
	})...)
}

func (m transportModel) emptyTransportText() string {
	if m.transports.Filtered() {
		return "   " + i18n.Sprintf("no transport matches the filter")
	}
	return "   " + i18n.Sprintf("no transports yet · press %s to add one", keyMap.Add.hint)
}

func (m transportModel) kindTabsLine(n int) string {
	label := fmt.Sprintf(" %s %d ", m.transports.Active(), n)
	return frameLine(transportInner, []span{
		{text: " ", fg: cDim},
		{text: label, fg: cInvFg, bg: cAccent, bold: true},
	}, nil)
}

func (m transportModel) transportHeaderLine() string {
	return frameLine(transportInner, []span{
		{text: truncPad("", wMark, false), fg: cDim},
		{text: " " + truncPad(i18n.Sprintf("NAME"), wTrName, false), fg: cDim},
		{text: " " + truncPad(i18n.Sprintf("KIND"), wTrKind, false), fg: cDim},
		{text: " " + truncPad(i18n.Sprintf("VIA"), wTrVia, false), fg: cDim},
		{text: " " + truncPad(i18n.Sprintf("LOCAL PORT"), wTrPort, false), fg: cDim},
		{text: " " + truncPad(i18n.Sprintf("USED BY"), wTrUsed, false), fg: cDim},
	}, nil)
}

func (m transportModel) transportRowLine(i int) string {
	t, ok := m.transports.At(i)
	if !ok {
		return frameLine(transportInner, nil, nil)
	}
	sel := i == m.transports.Cursor()
	used := len(m.transports.UsagesAt(i))

	var bg, fg, soft, markC color.Color
	switch {
	case sel:
		bg, fg, soft, markC = cAccent, cInvFg, cInvFg, cInvFg
	case !t.IsValid():
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
	case !t.IsValid():
		mark = gDotFail
	case used > 0:
		mark = gDotOn
	}

	port := i18n.Sprintf("auto")
	if t.LocalPort != 0 {
		port = strconv.Itoa(t.LocalPort)
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

	return frameLine(transportInner, []span{
		{text: caret + mark, fg: markC, bg: bg},
		{text: " " + truncPad(t.Name, wTrName, false), fg: fg, bg: bg},
		{text: " " + truncPad(t.Kind, wTrKind, false), fg: soft, bg: bg},
		{text: " " + truncPad(m.viaText(t), wTrVia, false), fg: soft, bg: bg},
		{text: " " + truncPad(port, wTrPort, false), fg: soft, bg: bg},
		{text: " " + truncPad(usedText, wTrUsed, false), fg: usedFg, bg: bg},
	}, bg)
}

func (m transportModel) transportStatusLine() string {
	icon, col := statusGlyph(m.statusKind)
	text := m.status
	if m.transports.Confirming() {
		icon, col, text = gStatusWarn, cAmber, i18n.Sprintf("delete transport %q? y/n", m.cursorTransportLabel())
	}
	return frameStatus(transportInner, icon, col, text, "")
}
