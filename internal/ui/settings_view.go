package ui

import (
	"image/color"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui/spec"
)

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

func (m settingsModel) render() string {
	n := len(m.conm.Connections)
	start, end := m.window.Range(n)

	lines := []string{
		frameTop(infoInner, "settings", ""),
		frameTabs(infoInner, settingsTabTitles, 0, scrollNote(m.window.Above(), m.window.Below(n))),
		frameRule(infoInner, gTeeL, gTeeR),
	}

	for i, db := range m.conm.Connections[start:end] {
		lines = append(lines, m.sectionLines(db, start+i == m.cursor)...)
	}
	lines = append(lines, frameLine(infoInner, nil, nil), frameRule(infoInner, gTeeL, gTeeR))

	if m.statusKind != kindIdle {
		icon, col := statusGlyph(m.statusKind)
		lines = append(lines,
			frameStatus(infoInner, icon, col, m.status, ""),
			frameRule(infoInner, gTeeL, gTeeR),
		)
	}

	lines = append(lines, keyhintLines(infoInner, keyhintFooter{
		groups: m.keyhints(),
		open:   m.help,
		pos:    cursorPos(m.cursor, n),
	})...)

	return strings.Join(append(lines, frameBottom(infoInner)), "\n")
}

func (m settingsModel) sectionLines(db config.ConnectionSettings, on bool) []string {
	lines := frameHead(infoInner, db.Type.String(), spec.ConnectionState(db.Enabled))

	for _, r := range m.clientRows(db) {
		for i, chunk := range wrapValue(r.value, infoInner-4-wInfoLbl) {
			lines = append(lines, infoRowLine(r, chunk, i == 0, false))
		}
	}

	return append(lines,
		settingsLine(settingsSupportsLabel, m.supportsSpans(db), false),
		settingsLine(settingsClientLabel, m.selectorSpans(db, on), on),
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
		case c.Name == db.CLI:
			fg = cAccent
		case c.Installed:
			fg = cFg
		}

		spans = append(spans, span{text: text, fg: fg, bold: c.Name == db.CLI})
	}

	return spans
}

func (m settingsModel) selectorSpans(db config.ConnectionSettings, on bool) []span {
	value := db.CLI
	if value == "" {
		value = gEmpty
	}

	var bg color.Color
	if on {
		bg = cAccent
	}

	return frameSelectBox(formBoxW, selectBox{
		value:  value,
		at:     slices.Index(cli.Clients(db.Type), db.CLI) + 1,
		n:      len(cli.Clients(db.Type)),
		active: on,
	}, bg)
}

func settingsLine(label string, value []span, on bool) string {
	labelFg, bg := cDim, color.Color(nil)
	caret := "  "
	if on {
		labelFg, bg, caret = cInvFg, cAccent, " "+gCaret
	}

	spans := make([]span, 0, 3+len(value))
	spans = append(spans,
		span{text: caret, fg: labelFg, bg: bg, bold: true},
		span{text: truncPad(label, wInfoLbl, false), fg: labelFg, bg: bg, bold: on},
		span{text: "  ", bg: bg},
	)

	return frameLine(infoInner, append(spans, value...), bg)
}
