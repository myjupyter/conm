package ui

import (
	"image/color"
	"slices"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/rivo/uniseg"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui/view"
)

const (
	appearanceWindow = 15
	hexDraftMax      = len("#rrggbb")
	colorDraftMax    = 64
	gradientBar      = 8
	swatch           = "    "
)

type appearanceInput int

const (
	inputNone appearanceInput = iota
	inputHex
	inputColor
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
	auto      color.Color
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

func colorGroup(
	order []string,
	presets map[string]map[string]string,
	preset func(config.AppearanceSettings) string,
	overrides func(*config.AppearanceSettings) *map[string]string,
	resolve func(config.AppearanceSettings) map[string]string,
	accepts func(string) bool,
	auto color.Color,
) elementGroup {
	return elementGroup{
		order: order, presets: presets, preset: preset, overrides: overrides, resolve: resolve,
		locked: never, input: inputColor, accepts: accepts, choices: config.ColorNames, auto: auto,
	}
}

var (
	borderColorGroup = colorGroup(config.BorderPalettes, borderPalettes,
		func(a config.AppearanceSettings) string { return a.BorderColor },
		func(a *config.AppearanceSettings) *map[string]string { return &a.BorderColors },
		borderColorNames, config.IsColorOrGradient, cBorder)
	glyphColorGroup = colorGroup(config.GlyphPalettes, glyphPalettes,
		func(a config.AppearanceSettings) string { return a.GlyphColor },
		func(a *config.AppearanceSettings) *map[string]string { return &a.GlyphColors },
		glyphColorNames, config.IsColor, nil)
	textColorGroup = colorGroup(config.TextPalettes, textPalettes,
		func(a config.AppearanceSettings) string { return a.TextColor },
		func(a *config.AppearanceSettings) *map[string]string { return &a.TextColors },
		textColorNames, config.IsColorOrGradient, nil)
)

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
	subhead string
	label   string
	display string
	input   appearanceInput

	options []string
	field   func(*config.AppearanceSettings) *string
	flag    func(*config.AppearanceSettings) *bool
	preview func(config.AppearanceSettings) []span

	group   *elementGroup
	element string
}

const (
	paletteLabel = "palette"
	colorLabel   = "colour"
	flagOff      = "off"
	flagOn       = "on"
)

var appearanceRows = buildAppearanceRows()

func buildAppearanceRows() []appearanceRow {
	rows := []appearanceRow{
		{
			section: colorLabel,
			label:   colorAccent,
			input:   inputHex,
			options: config.AccentColors,
			field:   func(a *config.AppearanceSettings) *string { return &a.Accent },
			preview: func(a config.AppearanceSettings) []span { return []span{{text: swatch, bg: accentColor(a.Accent)}} },
		},
		{
			section: colorLabel,
			label:   "gradients",
			options: []string{flagOff, flagOn},
			flag:    func(a *config.AppearanceSettings) *bool { return &a.Gradients },
			preview: func(a config.AppearanceSettings) []span {
				if !a.Gradients {
					return []span{{text: "every gradient draws in its first colour", fg: cMuted}}
				}
				return gradientSwatch(colorStops("teal>blue>purple>pink"))
			},
		},
		{
			section: "text",
			label:   "style",
			options: config.TextStyles,
			field:   func(a *config.AppearanceSettings) *string { return &a.Text },
			preview: textSample,
			group:   &textGroup,
		},
		paletteRow("text", config.TextPalettes, func(a *config.AppearanceSettings) *string { return &a.TextColor }, textSample),
	}
	rows = append(rows, elementRows("text", "style", &textGroup, config.TextElements, &textColorGroup, config.TextElements)...)
	rows = append(rows,
		appearanceRow{
			section: "border",
			label:   "style",
			options: config.BorderStyles,
			field:   func(a *config.AppearanceSettings) *string { return &a.Border },
			preview: borderSample,
			group:   &borderGroup,
		},
		paletteRow("border", config.BorderPalettes, func(a *config.AppearanceSettings) *string { return &a.BorderColor }, borderSample),
	)
	rows = append(rows, elementRows("border", "character", &borderGroup, config.BorderElements, &borderColorGroup, config.BorderElements)...)
	rows = append(rows,
		appearanceRow{
			section: "glyphs",
			label:   "set",
			options: config.GlyphSets,
			field:   func(a *config.AppearanceSettings) *string { return &a.Glyphs },
			preview: glyphSample,
			group:   &glyphGroup,
		},
		paletteRow("glyphs", config.GlyphPalettes, func(a *config.AppearanceSettings) *string { return &a.GlyphColor }, glyphSample),
	)
	return append(rows, elementRows("glyphs", "glyph", &glyphGroup, config.GlyphElements, &glyphColorGroup, config.GlyphColorElements)...)
}

func paletteRow(section string, options []string, field func(*config.AppearanceSettings) *string, preview func(config.AppearanceSettings) []span) appearanceRow {
	return appearanceRow{
		section: section,
		label:   paletteLabel,
		display: paletteLabel,
		input:   inputNone,
		options: options,
		field:   field,
		preview: preview,
	}
}

func textSample(a config.AppearanceSettings) []span {
	styles, colors := textStyles(a), textColorNames(a)
	sample := func(element, text string, auto color.Color) span {
		return span{text: text, fg: auto, style: roleStyle(styles[element], colors[element])}
	}
	return []span{
		sample("title", "title", cFg),
		{text: " "},
		sample("heading", "heading", cFg),
		{text: " "},
		sample("key", "key", cFg),
		{text: " "},
		sample("note", "note", cFaint),
	}
}

func borderSample(a config.AppearanceSettings) []span {
	r, c := borderRunes(a), resolvedColors(borderColorNames(a), cBorder)
	return []span{
		{text: r["top_left"], fg: c["top_left"]},
		{text: r["top"] + r["top"], fg: c["top"]},
		{text: r["top_right"], fg: c["top_right"]},
		{text: " "},
		{text: r["left"], fg: c["left"]},
		{text: "  "},
		{text: r["right"], fg: c["right"]},
	}
}

func glyphSample(a config.AppearanceSettings) []span {
	r, c := glyphRunes(a), resolvedColors(glyphColorNames(a), cSoft)
	sample := []string{"caret", "status_ok", "status_pending", "status_error", elDotOn, "select_next"}
	spans := make([]span, 0, len(sample))
	for _, element := range sample {
		spans = append(spans, span{text: r[element] + " ", fg: c[element]})
	}
	return spans
}

func resolvedColors(names map[string]string, auto color.Color) map[string]color.Color {
	colors := make(map[string]color.Color, len(names))
	for element, name := range names {
		colors[element] = auto
		if c := firstStop(name); c != nil {
			colors[element] = c
		}
	}
	return colors
}

func elementRows(section, valueLabel string, group *elementGroup, elements []string, colors *elementGroup, colored []string) []appearanceRow {
	rows := make([]appearanceRow, 0, 2*len(elements))
	for _, element := range elements {
		label := strings.ReplaceAll(element, "_", " ")
		rows = append(rows, appearanceRow{
			section: section,
			subhead: label,
			label:   label,
			display: "  " + valueLabel,
			input:   group.input,
			group:   group,
			element: element,
		})
		if slices.Contains(colored, element) {
			rows = append(rows, appearanceRow{
				section: section,
				subhead: label,
				label:   label + " " + colorLabel,
				display: "  " + colorLabel,
				input:   colors.input,
				group:   colors,
				element: element,
			})
		}
	}
	return rows
}

func (r appearanceRow) name() string {
	return r.section + " " + r.label
}

func (r appearanceRow) value(a config.AppearanceSettings) string {
	if r.flag != nil {
		if *r.flag(&a) {
			return flagOn
		}
		return flagOff
	}
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
	if r.flag != nil {
		*r.flag(a) = value == flagOn
		return
	}
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
	if r.group == nil {
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
	case inputColor:
		draft += strings.Map(colorRune, strings.ToLower(text))
		return draft[:min(len(draft), colorDraftMax)]
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

func colorRune(r rune) rune {
	if r == '#' || r == '>' || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
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

	if row.input == inputColor {
		if value := row.value(m.conm.Appearance); value != config.AutoColor {
			m.appear.draft = value
		}
		hint := "type a colour name or #rrggbb"
		if row.group.accepts("teal>pink") {
			hint += ", or a gradient like teal>pink"
		}
		m.setSettingsStatus(hint, kindIdle)
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
		newSection := i == start || row.section != appearanceRows[i-1].section
		if newSection {
			lines = append(lines, frameHead(m.inner(), row.section, m.sectionNote(row))...)
		}
		if row.subhead != "" && (newSection || row.subhead != appearanceRows[i-1].subhead) {
			if !newSection {
				lines = append(lines, frameLine(m.inner(), nil, nil))
			}
			lines = append(lines, frameSubhead(m.inner(), row.subhead, m.subheadNote(row)))
		}
		lines = append(lines, m.appearanceLine(row, i == m.appear.cursor))
	}
	return lines
}

func (m settingsModel) sectionNote(row appearanceRow) string {
	a := m.conm.Appearance
	if borderLocked(a) && row.group == &borderGroup {
		return "drawn by the ascii glyphs"
	}

	var groups []*elementGroup
	for i := range appearanceRows {
		r := &appearanceRows[i]
		if r.section == row.section && r.element != "" && !slices.Contains(groups, r.group) {
			groups = append(groups, r.group)
		}
	}
	n := 0
	for _, g := range groups {
		n += len(*g.overrides(&a))
	}
	switch n {
	case 0:
		return ""
	case 1:
		return "1 change to the preset"
	default:
		return strconv.Itoa(n) + " changes to the preset"
	}
}

func (m settingsModel) subheadNote(row appearanceRow) string {
	a := m.conm.Appearance
	for i := range appearanceRows {
		r := &appearanceRows[i]
		if r.section == row.section && r.subhead == row.subhead && !r.locked(a) && r.overridden(a) {
			return "changed"
		}
	}
	return ""
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
	if row.input == inputHex || row.input == inputColor {
		box.tint = firstStop(value)
	}
	preview := m.appearancePreview(row, a)
	switch {
	case row.locked(a):
		box.pos = "locked"
	case on && m.appear.editing:
		box.value, box.pos, box.tint = m.appear.draft+gInputCursor, "type", nil
		switch {
		case row.input == inputHex && config.IsHexColor(m.appear.draft):
			preview = []span{{text: swatch, bg: lipgloss.Color(m.appear.draft)}}
		case row.input == inputColor && row.accepts(m.appear.draft):
			preview = colorSwatch(row.group, m.appear.draft)
		}
	case !slices.Contains(choices, value):
		box.pos = "custom"
	}

	spans := frameSelectBox(formBoxW, box, bg)
	spans = append(spans, span{text: "  ", bg: bg})
	for _, p := range preview {
		if on && p.bg == nil {
			p.fg, p.bg, p.style.fg, p.style.stops = cInvFg, bg, nil, nil
		}
		spans = append(spans, p)
	}

	label := row.label
	if row.display != "" {
		label = row.display
	}
	return m.settingsLabelLine(label, bindLabelW, spans, on)
}

func (m settingsModel) appearancePreview(row appearanceRow, a config.AppearanceSettings) []span {
	if row.preview != nil {
		return row.preview(a)
	}

	var spans []span
	switch {
	case row.input == inputColor:
		spans = append(spans, colorSwatch(row.group, row.value(a))...)
	case row.group == &textGroup:
		style := roleStyle(row.value(a), textColorNames(a)[row.element])
		spans = append(spans, span{text: "sample", fg: cFg, style: style}, span{text: " "})
	}
	if !row.locked(a) && row.overridden(a) {
		spans = append(spans, span{text: "preset " + row.presetValue(a), fg: cMuted})
	}
	return spans
}

func colorSwatch(g *elementGroup, name string) []span {
	stops := colorStops(name)
	switch {
	case len(stops) > 1 && gradientsOn:
		return append(gradientSwatch(stops), span{text: " "})
	case len(stops) > 1:
		return []span{{text: swatch, bg: stops[0]}, {text: " flat while gradients are off ", fg: cMuted}}
	case len(stops) == 1:
		return []span{{text: swatch, bg: stops[0]}, {text: " "}}
	case g.auto != nil:
		return []span{{text: swatch, bg: g.auto}, {text: " "}}
	}
	return []span{{text: "set by the screen ", fg: cMuted}}
}

func gradientSwatch(stops []color.Color) []span {
	spans := make([]span, 0, gradientBar)
	for _, c := range lipgloss.Blend1D(gradientBar, stops...) {
		spans = append(spans, span{text: " ", bg: c})
	}
	return spans
}
