package ui

import (
	"image/color"
	"slices"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui/view"
)

const settingsSectionLines = 7

const (
	settingsPathLabel     = "path"
	settingsVersionLabel  = "version"
	settingsSupportsLabel = "supports"
	settingsClientLabel   = "client"
)

func (m settingsModel) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true

	return v
}

func (m settingsModel) inner() int {
	return m.size.modalWidth(infoInner)
}

func (m settingsModel) render() string {
	body := m.clientLines()
	if m.tab == settingsKeybindTab {
		body = m.bindsLines()
	}
	lines := append(m.headLines(), m.size.keepEnd(body, m.chrome())...)
	return m.size.join(append(lines, m.tailLines()...))
}

func (m settingsModel) chrome() int {
	return len(m.headLines()) + len(m.tailLines())
}

func (m settingsModel) tabWindow() (view.Scroll, int, int) {
	if m.tab == settingsKeybindTab {
		return m.binds.window, m.binds.cursor, len(m.bindings())
	}
	return m.window, m.cursor, len(m.conm.Connections)
}

func (m settingsModel) headLines() []string {
	window, _, n := m.tabWindow()
	return []string{
		frameTop(m.inner(), "settings", ""),
		frameTabs(m.inner(), settingsTabTitles, m.tab, scrollNote(window.Above(), window.Below(n))),
		frameRule(m.inner(), gTeeL, gTeeR),
	}
}

func (m settingsModel) tailLines() []string {
	_, cursor, n := m.tabWindow()
	lines := []string{frameLine(m.inner(), nil, nil), frameRule(m.inner(), gTeeL, gTeeR)}

	if m.statusKind != kindIdle {
		icon, col := statusGlyph(m.statusKind)
		lines = append(lines,
			frameStatus(m.inner(), icon, col, m.status, ""),
			frameRule(m.inner(), gTeeL, gTeeR),
		)
	}

	lines = append(lines, keyhintLines(m.inner(), keyhintFooter{
		groups: m.keyhints(),
		open:   m.help,
		pos:    cursorPos(cursor, n),
	})...)

	return append(lines, frameBottom(m.inner()))
}

func (m settingsModel) clientLines() []string {
	start, end := m.window.Range(len(m.conm.Connections))

	lines := make([]string, 0, (end-start)*settingsSectionLines)
	for i, db := range m.conm.Connections[start:end] {
		lines = append(lines, m.sectionLines(db, start+i == m.cursor)...)
	}
	return lines
}

func (m settingsModel) sectionLines(db config.ConnectionSettings, on bool) []string {
	note := ""
	if !m.installed(db.Type) {
		note = "no client installed"
	}
	lines := frameHead(m.inner(), db.Type.String(), note)

	for _, r := range m.clientRows(db) {
		for i, chunk := range wrapValue(r.value, m.inner()-4-wInfoLbl) {
			lines = append(lines, infoRowLine(m.inner(), r, chunk, i == 0, false))
		}
	}

	return append(lines,
		m.settingsLine(settingsSupportsLabel, m.supportsSpans(db), false),
		m.settingsLine(settingsClientLabel, m.selectorSpans(db, on), on),
	)
}

func (m settingsModel) clientRows(db config.ConnectionSettings) []infoRow {
	info, ok := m.clientInfo(db)
	version, probed := m.versions[info.Name]

	switch {
	case !ok:
		return []infoRow{
			{label: settingsPathLabel, value: gEmpty, tone: cFaint},
			{label: settingsVersionLabel, value: "no client configured", tone: cAmber},
		}
	case !info.Installed:
		return []infoRow{
			{label: settingsPathLabel, value: "not on PATH", tone: cAmber},
			{label: settingsVersionLabel, value: "not installed", tone: cAmber},
		}
	case !probed:
		version = gEllipsis
	case version == "":
		version = "unknown"
	}

	return []infoRow{
		{label: settingsPathLabel, value: info.Path, tone: cFg},
		{label: settingsVersionLabel, value: version, tone: cFg},
	}
}

func (m settingsModel) supportsSpans(db config.ConnectionSettings) []span {
	clients := m.clients[db.Type]

	spans := make([]span, 0, len(clients))
	for i, c := range clients {
		text := c.Name
		if i > 0 {
			text = ", " + c.Name
		}

		fg := cMuted
		switch {
		case c.Name == m.current(db):
			fg = cAccent
		case c.Installed:
			fg = cFg
		}

		spans = append(spans, span{text: text, fg: fg, bold: c.Name == m.current(db)})
	}

	return spans
}

func (m settingsModel) selectorSpans(db config.ConnectionSettings, on bool) []span {
	value := m.current(db)
	if value == "" {
		value = gEmpty
	}

	var bg color.Color
	if on {
		bg = cAccent
	}

	return frameSelectBox(formBoxW, selectBox{
		value:  value,
		at:     slices.Index(cli.Clients(db.Type), m.current(db)) + 1,
		n:      len(cli.Clients(db.Type)),
		active: on,
	}, bg)
}

func (m settingsModel) settingsLine(label string, value []span, on bool) string {
	return m.settingsLabelLine(label, wInfoLbl, value, on)
}

func (m settingsModel) settingsLabelLine(label string, labelW int, value []span, on bool) string {
	labelFg, bg := cDim, color.Color(nil)
	caret := "  "
	if on {
		labelFg, bg, caret = cInvFg, cAccent, " "+gCaret
	}

	spans := make([]span, 0, 3+len(value))
	spans = append(spans,
		span{text: caret, fg: labelFg, bg: bg, bold: true},
		span{text: truncPad(label, labelW, false), fg: labelFg, bg: bg, bold: on},
		span{text: "  ", bg: bg},
	)

	return frameLine(m.inner(), append(spans, value...), bg)
}
