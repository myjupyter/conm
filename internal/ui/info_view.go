package ui

import (
	"fmt"
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
)

const (
	infoInner = 88
	wInfoLbl  = 13
)

func (m infoModel) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true

	return v
}

func (m infoModel) render() string {
	lines := []string{
		frameTop(infoInner, "connection info", ""),
		m.titleLine(),
		frameRule(infoInner, gTeeL, gTeeR),
		m.infoTabsLine(),
		frameLine(infoInner, nil, nil),
	}

	lines = append(lines, m.infoRowLines()...)
	lines = append(lines, frameLine(infoInner, nil, nil), frameRule(infoInner, gTeeL, gTeeR))

	if m.statusKind != kindIdle {
		icon, col := statusGlyph(m.statusKind)
		lines = append(lines,
			frameStatus(infoInner, icon, col, m.status, ""),
			frameRule(infoInner, gTeeL, gTeeR),
		)
	}

	lines = append(lines, keyhintLines(infoInner, keyhintFooter{groups: m.keyhints(), open: m.help})...)

	return strings.Join(append(lines, frameBottom(infoInner)), "\n")
}

func (m infoModel) titleLine() string {
	cfg, ok := m.conns.Config()
	if !ok {
		return frameLine(infoInner, nil, nil)
	}

	kind := cfg.ConnType()

	return frameLine(infoInner, []span{
		{text: "  ", fg: cDim},
		{text: gDotOn + " ", fg: typeColor(kind)},
		{text: connLabel(cfg), fg: cFg, bold: true},
		{text: "  [" + kind.String() + "]", fg: typeColor(kind)},
	}, nil)
}

func (m infoModel) infoTabsLine() string {
	tabs := m.tabs()

	spans := make([]span, 0, 1+2*len(tabs))
	spans = append(spans, span{text: " ", fg: cDim})
	for _, tab := range tabs {
		s := span{text: " " + infoTabTitles[tab] + " ", fg: cDim}
		if tab == m.tab {
			s = span{text: s.text, fg: cInvFg, bg: cAccent, bold: true}
		}
		spans = append(spans, s, span{text: " ", fg: cDim})
	}

	return frameLine(infoInner, spans, nil)
}

func (m infoModel) infoRowLines() []string {
	var lines []string
	pick := 0
	for _, r := range m.rows() {
		if r.head {
			lines = append(lines, frameHead(infoInner, r.label, r.note)...)
			continue
		}

		on := pick == m.cursor
		pick++
		if r.hidden {
			continue
		}

		for i, chunk := range wrapValue(r.value, infoInner-4-wInfoLbl) {
			lines = append(lines, infoRowLine(r, chunk, i == 0, on))
		}
	}

	return lines
}

func infoRowLine(r infoRow, chunk string, first, on bool) string {
	fg, labelFg, bg := r.tone, cDim, color.Color(nil)
	if on {
		fg, labelFg, bg = cInvFg, cInvFg, cAccent
	}

	caret, label := "  ", ""
	if first {
		label = r.label
		if on {
			caret = " " + gCaret
		}
	}

	spans := []span{
		{text: caret, fg: labelFg, bg: bg, bold: true},
		{text: truncPad(label, wInfoLbl, false), fg: labelFg, bg: bg, bold: on},
		{text: "  " + chunk, fg: fg, bg: bg},
	}
	if on && first {
		spans = append(spans, infoHintSpans(r, len([]rune(chunk)), bg)...)
	}

	return frameLine(infoInner, spans, bg)
}

func infoHintSpans(r infoRow, valueWidth int, bg color.Color) []span {
	hint := keyMap.Yank.hint + " yanks"
	if r.link {
		hint = keyMap.Confirm.hint + " opens · " + hint
	}

	pad := infoInner - 4 - wInfoLbl - valueWidth - len([]rune(hint)) - 1
	if pad <= 0 {
		return nil
	}

	return []span{
		{text: strings.Repeat(" ", pad), bg: bg},
		{text: hint + " ", fg: cInvFg, bg: bg},
	}
}

func wrapValue(s string, w int) []string {
	r := []rune(s)
	if w <= 0 || len(r) <= w {
		return []string{s}
	}

	var out []string
	for len(r) > w {
		out, r = append(out, string(r[:w])), r[w:]
	}

	return append(out, string(r))
}

func scrollNote(above, below int) string {
	var parts []string
	if above > 0 {
		parts = append(parts, fmt.Sprintf("%s %d above", gScrollUp, above))
	}
	if below > 0 {
		parts = append(parts, fmt.Sprintf("%s %d below", gScrollDown, below))
	}

	return strings.Join(parts, " · ")
}
