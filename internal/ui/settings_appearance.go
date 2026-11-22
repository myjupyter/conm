package ui

import (
	"image/color"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
)

const hexDraftMax = len("#rrggbb")

type appearanceTab struct {
	cursor  int
	editing bool
	draft   string
}

type appearanceRow struct {
	label    string
	options  []string
	typed    bool
	field    func(*config.AppearanceSettings) *string
	lockedBy func(config.AppearanceSettings) string
	preview  func(value string, a config.AppearanceSettings) span
}

var appearanceRows = []appearanceRow{
	{
		label:   "border",
		options: config.BorderStyles,
		field:   func(a *config.AppearanceSettings) *string { return &a.Border },
		lockedBy: func(a config.AppearanceSettings) string {
			if a.Glyphs == config.ASCIIGlyphs {
				return "drawn by the ascii glyphs"
			}
			return ""
		},
		preview: func(value string, a config.AppearanceSettings) span {
			a.Border = value
			box := boxFor(a)
			return span{text: box.cornerTL + box.lineH + box.lineH + box.cornerTR, fg: cSoft}
		},
	},
	{
		label:   "accent",
		options: config.AccentColors,
		typed:   true,
		field:   func(a *config.AppearanceSettings) *string { return &a.Accent },
		preview: func(value string, _ config.AppearanceSettings) span {
			return span{text: "    ", bg: accentColor(value)}
		},
	},
	{
		label:   "glyphs",
		options: config.GlyphSets,
		field:   func(a *config.AppearanceSettings) *string { return &a.Glyphs },
		preview: func(value string, _ config.AppearanceSettings) span {
			g := glyphSets[value]
			return span{text: strings.Join([]string{g.caret, g.statusOK, g.statusPending, g.statusErr, g.dotOn, g.ellipsis}, " "), fg: cSoft}
		},
	},
}

func (r appearanceRow) locked(a config.AppearanceSettings) string {
	if r.lockedBy == nil {
		return ""
	}
	return r.lockedBy(a)
}

func (m settingsModel) appearanceRow() appearanceRow {
	return appearanceRows[m.appear.cursor]
}

func (m settingsModel) handleAppearanceKey(key string) (tea.Model, tea.Cmd, bool) {
	n := len(appearanceRows)
	switch {
	case keyMap.Down.matches(key):
		m.appear.cursor = (m.appear.cursor + 1) % n
	case keyMap.Up.matches(key):
		m.appear.cursor = (m.appear.cursor - 1 + n) % n
	case keyMap.Right.matches(key):
		model, cmd := m.cycleAppearance(1)
		return model, cmd, true
	case keyMap.Left.matches(key):
		model, cmd := m.cycleAppearance(-1)
		return model, cmd, true
	case keyMap.Settings.EditValue.matches(key) && m.appearanceRow().typed:
		m.appear.editing, m.appear.draft = true, "#"
		if value := *m.appearanceRow().field(&m.conm.Appearance); config.IsHexColor(value) {
			m.appear.draft = value
		}
		m.setSettingsStatus("type a #rrggbb colour", kindIdle)
		return m, nil, true
	default:
		return m, nil, false
	}

	m.setSettingsStatus(statusReady, kindIdle)
	return m, nil, true
}

func (m settingsModel) handleHexKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch {
	case keyMap.Interrupt.matches(key):
		return m, tea.Quit
	case keyMap.Cancel.matches(key):
		m.appear.editing = false
		m.setSettingsStatus(statusReady, kindIdle)
	case keyMap.Confirm.matches(key):
		return m.saveHexDraft()
	case keyMap.Backspace.matches(key):
		r := []rune(m.appear.draft)
		m.appear.draft = string(r[:max(len(r)-1, 0)])
	default:
		m = m.typeHex(msg.Text)
	}
	return m, nil
}

func (m settingsModel) typeHex(text string) settingsModel {
	draft := m.appear.draft + strings.Map(hexRune, text)
	m.appear.draft = draft[:min(len(draft), hexDraftMax)]
	return m
}

func hexRune(r rune) rune {
	if r == '#' || strings.ContainsRune("0123456789abcdefABCDEF", r) {
		return r
	}
	return -1
}

func (m settingsModel) saveHexDraft() (tea.Model, tea.Cmd) {
	if !config.IsHexColor(m.appear.draft) {
		m.setSettingsStatus(m.appear.draft+" is not a #rrggbb colour", kindWarn)
		return m, nil
	}

	row := m.appearanceRow()
	next := m.conm.Appearance
	*row.field(&next) = m.appear.draft
	m.appear.editing = false

	return m, m.saveAppearanceCmd(next, row.label+" → "+m.appear.draft)
}

func (m settingsModel) cycleAppearance(step int) (tea.Model, tea.Cmd) {
	row := m.appearanceRow()
	if reason := row.locked(m.conm.Appearance); reason != "" {
		m.setSettingsStatus(row.label+" is "+reason+" · pick other glyphs to choose one", kindWarn)
		return m, nil
	}

	next := m.conm.Appearance
	value := row.field(&next)
	n := len(row.options)
	*value = row.options[(slices.Index(row.options, *value)+step+n)%n]

	return m, m.saveAppearanceCmd(next, row.label+" → "+*value)
}

func (m settingsModel) appearanceLines() []string {
	lines := make([]string, 0, 1+len(appearanceRows))
	lines = append(lines, frameLine(m.inner(), nil, nil))
	for i, row := range appearanceRows {
		lines = append(lines, m.appearanceLine(row, i == m.appear.cursor))
	}
	return lines
}

func (m settingsModel) appearanceLine(row appearanceRow, on bool) string {
	a := m.conm.Appearance
	value := *row.field(&a)

	var bg color.Color
	if on {
		bg = cAccent
	}

	box := selectBox{value: value, at: slices.Index(row.options, value) + 1, n: len(row.options), active: on}
	preview := row.preview(value, a)
	reason := row.locked(a)
	switch {
	case reason != "":
		box.value, box.pos = config.ASCIIGlyphs, "locked"
		preview = span{text: reason, fg: cMuted}
	case on && m.appear.editing:
		box.value, box.pos = m.appear.draft+gInputCursor, "hex"
		if config.IsHexColor(m.appear.draft) {
			preview = row.preview(m.appear.draft, a)
		}
	case !slices.Contains(row.options, value):
		box.pos = "custom"
	}
	if on && preview.bg == nil {
		preview.fg, preview.bg = cInvFg, bg
	}

	spans := frameSelectBox(formBoxW, box, bg)
	spans = append(spans, span{text: "  ", bg: bg}, preview)

	return m.settingsLine(row.label, spans, on)
}
