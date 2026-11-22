package ui

import (
	"image/color"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
)

type appearanceRow struct {
	label   string
	options []string
	field   func(*config.AppearanceSettings) *string
	preview func(value string) span
}

var appearanceRows = []appearanceRow{
	{
		label:   "border",
		options: config.BorderStyles,
		field:   func(a *config.AppearanceSettings) *string { return &a.Border },
		preview: func(value string) span {
			box := borderStyles[value]
			return span{text: box.cornerTL + box.lineH + box.lineH + box.cornerTR, fg: cSoft}
		},
	},
	{
		label:   "accent",
		options: config.AccentColors,
		field:   func(a *config.AppearanceSettings) *string { return &a.Accent },
		preview: func(value string) span { return span{text: "    ", bg: accentColors[value]} },
	},
	{
		label:   "glyphs",
		options: config.GlyphSets,
		field:   func(a *config.AppearanceSettings) *string { return &a.Glyphs },
		preview: func(value string) span {
			g := glyphSets[value]
			return span{text: strings.Join([]string{g.caret, g.statusOK, g.statusPending, g.statusErr, g.dotOn, g.ellipsis}, " "), fg: cSoft}
		},
	},
}

func (m settingsModel) handleAppearanceKey(key string) (tea.Model, tea.Cmd, bool) {
	n := len(appearanceRows)
	switch {
	case keyMap.Down.matches(key):
		m.appearCursor = (m.appearCursor + 1) % n
	case keyMap.Up.matches(key):
		m.appearCursor = (m.appearCursor - 1 + n) % n
	case keyMap.Right.matches(key):
		return m, m.cycleAppearance(1), true
	case keyMap.Left.matches(key):
		return m, m.cycleAppearance(-1), true
	default:
		return m, nil, false
	}

	m.setSettingsStatus(statusReady, kindIdle)
	return m, nil, true
}

func (m settingsModel) cycleAppearance(step int) tea.Cmd {
	row := appearanceRows[m.appearCursor]
	next := m.conm.Appearance
	value := row.field(&next)
	n := len(row.options)
	*value = row.options[(slices.Index(row.options, *value)+step+n)%n]

	return m.saveAppearanceCmd(next, row.label+" → "+*value)
}

func (m settingsModel) appearanceLines() []string {
	lines := make([]string, 0, 1+len(appearanceRows))
	lines = append(lines, frameLine(m.inner(), nil, nil))
	for i, row := range appearanceRows {
		lines = append(lines, m.appearanceLine(row, i == m.appearCursor))
	}
	return lines
}

func (m settingsModel) appearanceLine(row appearanceRow, on bool) string {
	value := *row.field(&m.conm.Appearance)

	var bg color.Color
	if on {
		bg = cAccent
	}

	spans := frameSelectBox(formBoxW, selectBox{
		value:  value,
		at:     slices.Index(row.options, value) + 1,
		n:      len(row.options),
		active: on,
	}, bg)
	spans = append(spans, span{text: "  ", bg: bg}, row.preview(value))

	return m.settingsLine(row.label, spans, on)
}
