package ui

import (
	"fmt"
	"image/color"
	"slices"
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

func (m infoModel) inner() int {
	return m.size.modalWidth(infoInner)
}

func (m infoModel) render() string {
	body, _ := m.infoRowLines()
	start, end := m.body.Range(len(body))
	lines := append(m.headLines(scrollNote(start, len(body)-end)), body[start:end]...)
	return m.size.join(append(lines, m.tailLines()...))
}

func (m infoModel) headLines(scroll string) []string {
	return []string{
		frameTop(m.inner(), "connection info", topNote(scroll)),
		m.titleLine(),
		frameRule(m.inner(), gTeeL, gTeeR),
		m.infoTabsLine(),
		frameLine(m.inner(), nil, nil),
	}
}

func (m infoModel) tailLines() []string {
	lines := []string{frameLine(m.inner(), nil, nil), frameRule(m.inner(), gTeeL, gTeeR)}

	if m.statusKind != kindIdle {
		icon, col := statusGlyph(m.statusKind)
		lines = append(lines,
			frameStatus(m.inner(), icon, col, m.status, ""),
			frameRule(m.inner(), gTeeL, gTeeR),
		)
	}

	lines = append(lines, keyhintLines(m.inner(), keyhintFooter{groups: m.keyhints(), open: m.help})...)

	return append(lines, frameBottom(m.inner()))
}

func (m infoModel) titleLine() string {
	cfg, ok := m.conns.Config()
	if !ok {
		return frameLine(m.inner(), nil, nil)
	}

	kind := cfg.ConnType()

	return frameLine(m.inner(), []span{
		{text: "  ", fg: cDim},
		{text: gDotOn + " ", fg: typeColor(kind)},
		{text: connLabel(cfg), fg: cFg, bold: true},
		{text: "  [" + kind.String() + "]", fg: typeColor(kind)},
	}, nil)
}

func (m infoModel) infoTabsLine() string {
	return frameTabs(m.inner(), m.tabTitles(), slices.Index(m.tabs(), m.tab), "")
}

func (m infoModel) infoRowLines() (lines []string, focus focusSpan) {
	pick, head, headed := 0, 0, false
	for _, r := range m.rows() {
		if r.head {
			head, headed = len(lines), true
			lines = append(lines, frameHead(m.inner(), r.label, r.note)...)
			continue
		}

		on := pick == m.cursor
		pick++
		if r.hidden {
			continue
		}

		at := len(lines)
		for i, chunk := range wrapValue(r.value, m.inner()-4-wInfoLbl) {
			lines = append(lines, infoRowLine(m.inner(), r, chunk, i == 0, on))
		}
		if on {
			focus = focusSpan{from: at, row: at, to: len(lines) - 1}
			if headed {
				focus.from = head
			}
		}
		headed = false
	}

	return lines, focus
}

func infoRowLine(w int, r infoRow, chunk string, first, on bool) string {
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
		spans = append(spans, infoHintSpans(w, r, len([]rune(chunk)), bg)...)
	}

	return frameLine(w, spans, bg)
}

func infoHintSpans(w int, r infoRow, valueWidth int, bg color.Color) []span {
	hint := keyMap.Copy.hint + " copies"
	if r.link {
		hint = keyMap.Confirm.hint + " opens · " + hint
	}

	pad := w - 4 - wInfoLbl - valueWidth - len([]rune(hint)) - 1
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
