package ui

import (
	"image/color"
	"slices"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/rivo/uniseg"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui/view"
)

const (
	appearanceWindow = 15
	hexDraftMax      = len("#rrggbb")
)

type appearanceInput int

const (
	inputNone appearanceInput = iota
	inputHex
	inputGlyph
	inputStyle
)

type appearanceTab struct {
	cursor  int
	window  view.Scroll
	editing bool
	draft   string
}

type elementGroup struct {
	order     []string
	presets   map[string]map[string]string
	preset    func(config.AppearanceSettings) string
	overrides func(*config.AppearanceSettings) *map[string]string
	resolve   func(config.AppearanceSettings) map[string]string
	locked    func(config.AppearanceSettings) bool
	input     appearanceInput
	accepts   func(string) bool
	choices   []string
}

func never(config.AppearanceSettings) bool { return false }

var borderGroup = elementGroup{
	order:     config.BorderStyles,
	presets:   borderPresets,
	preset:    func(a config.AppearanceSettings) string { return a.Border },
	overrides: func(a *config.AppearanceSettings) *map[string]string { return &a.BorderRunes },
	resolve:   borderRunes,
	locked:    borderLocked,
	input:     inputGlyph,
	accepts:   config.IsBorderRune,
}

var glyphGroup = elementGroup{
	order:     config.GlyphSets,
	presets:   glyphPresets,
	preset:    func(a config.AppearanceSettings) string { return a.Glyphs },
	overrides: func(a *config.AppearanceSettings) *map[string]string { return &a.GlyphRunes },
	resolve:   glyphRunes,
	locked:    never,
	input:     inputGlyph,
	accepts:   config.IsGlyph,
}

var textGroup = elementGroup{
	order:     config.TextStyles,
	presets:   textPresets,
	preset:    func(a config.AppearanceSettings) string { return a.Text },
	overrides: func(a *config.AppearanceSettings) *map[string]string { return &a.TextStyle },
	resolve:   textStyles,
	locked:    never,
	input:     inputStyle,
	accepts:   config.IsTextStyle,
	choices: []string{
		config.PlainText, "bold", "italic", "underline", "faint",
		"bold+italic", "bold+underline", "italic+underline", "italic+faint",
	},
}

func (g elementGroup) variants(element string) []string {
	if g.choices != nil {
		return g.choices
	}
	var variants []string
	for _, name := range g.order {
		if r := g.presets[name][element]; !slices.Contains(variants, r) {
			variants = append(variants, r)
		}
	}
	return variants
}

type appearanceRow struct {
	section string
	label   string
	input   appearanceInput

	options []string
	field   func(*config.AppearanceSettings) *string
	preview func(config.AppearanceSettings) []span

	group   *elementGroup
	element string
}

var appearanceRows = buildAppearanceRows()

func buildAppearanceRows() []appearanceRow {
	rows := []appearanceRow{
		{
			section: "accent",
			label:   "colour",
			input:   inputHex,
			options: config.AccentColors,
			field:   func(a *config.AppearanceSettings) *string { return &a.Accent },
			preview: func(a config.AppearanceSettings) []span { return []span{{text: "    ", bg: accentColor(a.Accent)}} },
		},
		{
			section: "text",
			label:   "style",
			options: config.TextStyles,
			field:   func(a *config.AppearanceSettings) *string { return &a.Text },
			preview: func(a config.AppearanceSettings) []span {
				t := textStyles(a)
				return []span{
					{text: "title", fg: cFg, style: parseTextStyle(t["title"])},
					{text: " heading", fg: cFg, style: parseTextStyle(t["heading"])},
					{text: " note", fg: cFaint, style: parseTextStyle(t["note"])},
				}
			},
			group: &textGroup,
		},
	}
	rows = append(rows, elementRows("text", &textGroup, config.TextElements)...)
	rows = append(rows, appearanceRow{
		section: "border",
		label:   "style",
		options: config.BorderStyles,
		field:   func(a *config.AppearanceSettings) *string { return &a.Border },
		preview: func(a config.AppearanceSettings) []span {
			r := borderRunes(a)
			return []span{{text: r["top_left"] + r["top"] + r["top"] + r["top_right"], fg: cSoft}}
		},
		group: &borderGroup,
	})
	rows = append(rows, elementRows("border", &borderGroup, config.BorderElements)...)
	rows = append(rows, appearanceRow{
		section: "glyphs",
		label:   "set",
		options: config.GlyphSets,
		field:   func(a *config.AppearanceSettings) *string { return &a.Glyphs },
		preview: func(a config.AppearanceSettings) []span {
			r := glyphRunes(a)
			sample := []string{r["caret"], r["status_ok"], r["status_pending"], r["status_error"], r["dot_on"], r["ellipsis"]}
			return []span{{text: strings.Join(sample, " "), fg: cSoft}}
		},
		group: &glyphGroup,
	})
	return append(rows, elementRows("glyphs", &glyphGroup, config.GlyphElements)...)
}

func elementRows(section string, group *elementGroup, elements []string) []appearanceRow {
	rows := make([]appearanceRow, 0, len(elements))
	for _, element := range elements {
		rows = append(rows, appearanceRow{
			section: section,
			label:   strings.ReplaceAll(element, "_", " "),
			input:   group.input,
			group:   group,
			element: element,
		})
	}
	return rows
}

func (r appearanceRow) name() string {
	return r.section + " " + r.label
}

func (r appearanceRow) value(a config.AppearanceSettings) string {
	if r.element != "" {
		return r.group.resolve(a)[r.element]
	}
	return *r.field(&a)
}

func (r appearanceRow) choices() []string {
	if r.element != "" {
		return r.group.variants(r.element)
	}
	return r.options
}

func (r appearanceRow) locked(a config.AppearanceSettings) bool {
	return r.group != nil && r.group.locked(a)
}

func (r appearanceRow) presetValue(a config.AppearanceSettings) string {
	return r.group.presets[r.group.preset(a)][r.element]
}

func (r appearanceRow) overridden(a config.AppearanceSettings) bool {
	_, ok := (*r.group.overrides(&a))[r.element]
	return ok
}

func (r appearanceRow) set(a *config.AppearanceSettings, value string) {
	if r.element == "" {
		*r.field(a) = value
		return
	}

	overrides := r.group.overrides(a)
	if value == r.presetValue(*a) {
		delete(*overrides, r.element)
		return
	}
	if *overrides == nil {
		*overrides = make(map[string]string)
	}
	(*overrides)[r.element] = value
}

func (r appearanceRow) accepts(draft string) bool {
	if r.input == inputHex {
		return config.IsHexColor(draft)
	}
	return r.group.accepts(draft)
}

func (r appearanceRow) typed(draft, text string) string {
	switch r.input {
	case inputHex:
		draft += strings.Map(hexRune, text)
		return draft[:min(len(draft), hexDraftMax)]
	case inputStyle:
		return draft + strings.Map(styleRune, strings.ToLower(text))
	default:
		if last := lastGrapheme(text); last != "" {
			return last
		}
		return draft
	}
}

func lastGrapheme(text string) string {
	var last string
	g := uniseg.NewGraphemes(text)
	for g.Next() {
		last = g.Str()
	}
	return last
}

func hexRune(r rune) rune {
	if r == '#' || strings.ContainsRune("0123456789abcdefABCDEF", r) {
		return r
	}
	return -1
}

func styleRune(r rune) rune {
	if r == '+' || (r >= 'a' && r <= 'z') {
		return r
	}
	return -1
}

func (m settingsModel) appearanceRow() appearanceRow {
	return appearanceRows[m.appear.cursor]
}

func (m settingsModel) handleAppearanceKey(key string) (tea.Model, tea.Cmd, bool) {
	n := len(appearanceRows)
	row := m.appearanceRow()
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
	case keyMap.Settings.EditValue.matches(key) && row.input != inputNone:
		return m.startAppearanceInput(), nil, true
	case keyMap.Settings.Reset.matches(key) && row.element != "":
		model, cmd := m.resetElement()
		return model, cmd, true
	default:
		return m, nil, false
	}

	m.setSettingsStatus(statusReady, kindIdle)
	return m, nil, true
}

func (m settingsModel) startAppearanceInput() settingsModel {
	row := m.appearanceRow()
	if row.locked(m.conm.Appearance) {
		m.setSettingsStatus(lockedStatus(row), kindWarn)
		return m
	}

	m.appear.editing, m.appear.draft = true, ""
	if row.input == inputHex {
		m.appear.draft = "#"
		if value := row.value(m.conm.Appearance); config.IsHexColor(value) {
			m.appear.draft = value
		}
		m.setSettingsStatus("type a #rrggbb colour", kindIdle)
		return m
	}

	if row.input == inputStyle {
		m.setSettingsStatus("type plain or attributes joined by +: "+strings.Join(config.TextAttributes, ", "), kindIdle)
		return m
	}

	m.setSettingsStatus("type one character for the "+row.name(), kindIdle)
	return m
}

func (m settingsModel) handleAppearanceInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch {
	case keyMap.Interrupt.matches(key):
		return m, tea.Quit
	case keyMap.Cancel.matches(key):
		m.appear.editing = false
		m.setSettingsStatus(statusReady, kindIdle)
	case keyMap.Confirm.matches(key):
		return m.saveAppearanceDraft()
	case keyMap.Backspace.matches(key):
		r := []rune(m.appear.draft)
		m.appear.draft = string(r[:max(len(r)-1, 0)])
	default:
		m = m.typeAppearance(msg.Text)
	}
	return m, nil
}

func (m settingsModel) typeAppearance(text string) settingsModel {
	m.appear.draft = m.appearanceRow().typed(m.appear.draft, text)
	return m
}

func (m settingsModel) saveAppearanceDraft() (tea.Model, tea.Cmd) {
	row := m.appearanceRow()
	if !row.accepts(m.appear.draft) {
		m.setSettingsStatus(strconv.Quote(m.appear.draft)+" does not fit the "+row.name(), kindWarn)
		return m, nil
	}

	m.appear.editing = false
	return m, m.setAppearance(row, m.appear.draft)
}

func (m settingsModel) cycleAppearance(step int) (tea.Model, tea.Cmd) {
	row := m.appearanceRow()
	if row.locked(m.conm.Appearance) {
		m.setSettingsStatus(lockedStatus(row), kindWarn)
		return m, nil
	}

	choices := row.choices()
	n := len(choices)
	at := slices.Index(choices, row.value(m.conm.Appearance))
	if at < 0 && step < 0 {
		at = 0
	}

	return m, m.setAppearance(row, choices[((at+step)%n+n)%n])
}

func (m settingsModel) resetElement() (tea.Model, tea.Cmd) {
	row := m.appearanceRow()
	if !row.overridden(m.conm.Appearance) {
		m.setSettingsStatus(row.name()+" already follows the preset", kindIdle)
		return m, nil
	}

	return m, m.setAppearance(row, row.presetValue(m.conm.Appearance))
}

func (m settingsModel) setAppearance(row appearanceRow, value string) tea.Cmd {
	next := m.conm.Appearance.Clone()
	row.set(&next, value)

	return m.saveAppearanceCmd(next, row.name()+" → "+value)
}

func lockedStatus(row appearanceRow) string {
	return row.section + " is drawn by the ascii glyphs · pick other glyphs to change it"
}

func (m settingsModel) appearanceLines() []string {
	start, end := m.appear.window.Range(len(appearanceRows))

	var lines []string
	for i := start; i < end; i++ {
		row := appearanceRows[i]
		if i == start || row.section != appearanceRows[i-1].section {
			lines = append(lines, frameHead(m.inner(), row.section, m.sectionNote(row))...)
		}
		lines = append(lines, m.appearanceLine(row, i == m.appear.cursor))
	}
	return lines
}

func (m settingsModel) sectionNote(row appearanceRow) string {
	a := m.conm.Appearance
	switch {
	case row.group == nil:
		return ""
	case row.group.locked(a):
		return "drawn by the ascii glyphs"
	}

	switch n := len(*row.group.overrides(&a)); n {
	case 0:
		return ""
	case 1:
		return "1 element customised"
	default:
		return strconv.Itoa(n) + " elements customised"
	}
}

func (m settingsModel) appearanceLine(row appearanceRow, on bool) string {
	a := m.conm.Appearance
	value := row.value(a)
	choices := row.choices()

	var bg color.Color
	if on {
		bg = cAccent
	}

	box := selectBox{value: value, at: slices.Index(choices, value) + 1, n: len(choices), active: on}
	preview := m.appearancePreview(row, a)
	switch {
	case row.locked(a):
		box.pos = "locked"
	case on && m.appear.editing:
		box.value, box.pos = m.appear.draft+gInputCursor, "type"
		if row.input == inputHex && config.IsHexColor(m.appear.draft) {
			preview = []span{{text: "    ", bg: accentColor(m.appear.draft)}}
		}
	case !slices.Contains(choices, value):
		box.pos = "custom"
	}

	spans := frameSelectBox(formBoxW, box, bg)
	spans = append(spans, span{text: "  ", bg: bg})
	for _, p := range preview {
		if on && p.bg == nil {
			p.fg, p.bg = cInvFg, bg
		}
		spans = append(spans, p)
	}

	return m.settingsLabelLine(row.label, bindLabelW, spans, on)
}

func (m settingsModel) appearancePreview(row appearanceRow, a config.AppearanceSettings) []span {
	if row.preview != nil {
		return row.preview(a)
	}

	var spans []span
	if row.group == &textGroup {
		spans = append(spans, span{text: "sample ", fg: cFg, style: parseTextStyle(row.value(a))})
	}
	if !row.locked(a) && row.overridden(a) {
		spans = append(spans, span{text: "preset " + row.presetValue(a), fg: cMuted})
	}
	return spans
}
