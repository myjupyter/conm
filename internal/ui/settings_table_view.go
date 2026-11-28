package ui

import (
	"image/color"
	"slices"
	"strconv"
	"strings"

	"github.com/myjupyter/conm/internal/config"
)

func (m settingsModel) tableLines() []string {
	rows := m.tableRows()
	start, end := m.table.window.Range(len(rows))

	var lines []string
	for i := start; i < end; i++ {
		row := rows[i]
		newType := i == start || row.connType != rows[i-1].connType
		if newType {
			lines = append(lines, frameHead(m.inner(), row.connType.String(), m.columnsNote(row.connType))...)
		}
		if !row.isColor() && (newType || rows[i-1].isColor()) {
			if !newType {
				lines = append(lines, frameLine(m.inner(), nil, nil))
			}
			lines = append(lines, frameSubhead(m.inner(), "columns", ""))
		}
		lines = append(lines, m.tableLine(row, m.nested && i == m.table.cursor))
	}
	return lines
}

func (m settingsModel) columnsNote(t config.ConnType) string {
	s, _ := m.conm.Connection(t)
	titles := make([]string, 0, len(s.VisibleColumns()))
	for _, c := range s.VisibleColumns() {
		titles = append(titles, columnTitles[c])
	}
	return strings.Join(titles, " · ")
}

func (m settingsModel) tableLine(row tableRow, on bool) string {
	var bg color.Color
	if on {
		bg = cAccent
	}
	if row.isColor() {
		return m.settingsLabelLine(tabColorLabel, bindLabelW, m.tabColorSpans(row.connType, on, bg), on)
	}
	return m.settingsLabelLine("  "+row.column, bindLabelW, m.columnSpans(row, on, bg), on)
}

func (m settingsModel) tabColorSpans(t config.ConnType, on bool, bg color.Color) []span {
	s, _ := m.conm.Connection(t)
	value := s.ResolvedTabColor()
	box := selectBox{
		value:  value,
		at:     slices.Index(tabColorChoices, value) + 1,
		n:      len(tabColorChoices),
		tint:   tabColor(t),
		active: on,
	}
	switch {
	case on && m.table.editing:
		box.value, box.pos, box.tint = m.table.draft+gInputCursor, "type", nil
	case !slices.Contains(tabColorChoices, value):
		box.pos = "custom"
	}

	spans := frameSelectBox(formBoxW, box, bg)
	return append(spans, span{text: "  ", bg: bg}, span{text: swatch, bg: tabColor(t)})
}

func (m settingsModel) columnSpans(row tableRow, on bool, bg color.Color) []span {
	s, _ := m.conm.Connection(row.connType)
	at := slices.Index(s.VisibleColumns(), row.column)

	box, fg := "[ ]", cMuted
	order := ""
	if at >= 0 {
		box, fg, order = "[x]", cFg, strconv.Itoa(at+1)
	}
	if on {
		fg = cInvFg
	}
	spans := []span{{text: box + " " + truncPad(order, 2, true), fg: fg, bg: bg, style: tValue}}
	if row.column == config.NameColumn {
		spans = append(spans, span{text: "  always shown", fg: cMuted, bg: bg})
	}
	return spans
}
