package ui

import (
	"image/color"
	"maps"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/rivo/uniseg"

	"github.com/myjupyter/conm/internal/config"
)

var (
	cBorder = lipgloss.Color("#3a362e")
	cDim    = lipgloss.Color("#6d6759")
	cMuted  = lipgloss.Color("#5f594c")
	cFaint  = lipgloss.Color("#4d483e")
	cSoft   = lipgloss.Color("#8b8474")
	cFg     color.Color
	cDead   = lipgloss.Color("#7d7668")
	cAccent color.Color
	cOK     = lipgloss.Color("#63d18a")
	cRed    = lipgloss.Color("#e5654a")
	cAmber  = lipgloss.Color("#e3b34a")
	cInvFg  = lipgloss.Color("#0f0f0d")
	cValue  = lipgloss.Color("#c4bcac")
	cInput  color.Color

	baseFg      = lipgloss.Color("#ddd6c8")
	baseInput   = lipgloss.Color("#c4bcac")
	cHeadBg     = lipgloss.Color("#1a1912")
	cRail       = lipgloss.Color("#33302a")
	cPairBg     = lipgloss.Color("#151a13")
	cPong       = lipgloss.Color("#b9cdbd")
	cPostgres   = lipgloss.Color("#5aa0d6")
	cMySQL      = lipgloss.Color("#c98f2f")
	cMSSQL      = lipgloss.Color("#a07fd0")
	cClickHouse = lipgloss.Color("#e0c341")
	cRedis      = lipgloss.Color("#c4574d")
	cMongoDB    = lipgloss.Color("#4faa41")
	cSSH        = lipgloss.Color("#9ea6b4")

	cErrTagBg = lipgloss.Color("#9e3b2a")
	cErrTagFg = lipgloss.Color("#fdf3f1")
	cErrCode  = lipgloss.Color("#f0a48f")
	cErrMuted = lipgloss.Color("#8d7a76")
	cErrValue = lipgloss.Color("#c8b8b4")
	cHint     = lipgloss.Color("#a2938f")
)

var (
	gCornerTL, gCornerTR, gCornerBL, gCornerBR string
	gTeeL, gTeeR                               string
	gLineTop, gLineBottom, gLineH              string
	gLineLeft, gLineRight                      string

	gStatusIdle, gStatusOK, gStatusPending, gStatusWarn, gStatusErr string

	gImportNew                          string
	gCaret, gDotOn, gDotOff, gDotFail   string
	gFieldBad, gInputCursor, gInputMask string
	gEmpty, gEllipsis                   string
	gHintArrow, gScrollUp, gScrollDown  string
	gSelectPrev, gSelectNext            string
)

const (
	gEmptyList = "none"
)

var borderVars = map[string]*string{
	"top_left": &gCornerTL, "top_right": &gCornerTR, "bottom_left": &gCornerBL, "bottom_right": &gCornerBR,
	"tee_left": &gTeeL, "tee_right": &gTeeR,
	"top": &gLineTop, "bottom": &gLineBottom, "rule": &gLineH, "left": &gLineLeft, "right": &gLineRight,
}

var glyphCell int

var (
	gradientsOn     bool
	borderGradients map[string][]color.Color
	borderColors    map[string]color.Color
	glyphColors     map[string]color.Color
)

var (
	tTitle, tTab, tHeading, tSelection, tKey, tValue, tNote textStyle
)

var textVars = map[string]*textStyle{
	roleTitle: &tTitle, "tab": &tTab, roleHeading: &tHeading, "selection": &tSelection,
	"key": &tKey, "value": &tValue, "note": &tNote,
}

var glyphVars = map[string]*string{
	"status_idle": &gStatusIdle, "status_ok": &gStatusOK, "status_pending": &gStatusPending,
	"status_warn": &gStatusWarn, "status_error": &gStatusErr, "import_new": &gImportNew,
	"caret": &gCaret, elDotOn: &gDotOn, elDotOff: &gDotOff, elDotFail: &gDotFail,
	"field_bad": &gFieldBad, "input_cursor": &gInputCursor, "input_mask": &gInputMask,
	"empty": &gEmpty, "ellipsis": &gEllipsis, "hint_arrow": &gHintArrow,
	"scroll_up": &gScrollUp, "scroll_down": &gScrollDown,
	"select_prev": &gSelectPrev, "select_next": &gSelectNext,
}

const (
	elDotOn     = "dot_on"
	elDotOff    = "dot_off"
	elDotFail   = "dot_fail"
	colorAccent = config.AccentColor
	colorSoft   = "soft"
	roleHeading = "heading"
	roleTitle   = "title"
)

const (
	defaultBorder  = "single"
	asciiBorder    = "ascii"
	defaultGlyphs  = "unicode"
	defaultAccent  = "green"
	defaultText    = "default"
	defaultPalette = "default"
)

func zip(elements []string, values ...string) map[string]string {
	runes := make(map[string]string, len(elements))
	for i, value := range values {
		runes[elements[i]] = value
	}
	return runes
}

func box(tl, tr, bl, br, teeL, teeR, h, v string) map[string]string {
	return frame(tl, tr, bl, br, teeL, teeR, h, h, h, v, v)
}

func frame(tl, tr, bl, br, teeL, teeR, top, bottom, rule, left, right string) map[string]string {
	return zip(config.BorderElements, tl, tr, bl, br, teeL, teeR, top, bottom, rule, left, right)
}

var borderPresets = map[string]map[string]string{
	"single":    box("┌", "┐", "└", "┘", "├", "┤", "─", "│"),
	"rounded":   box("╭", "╮", "╰", "╯", "├", "┤", "─", "│"),
	"double":    box("╔", "╗", "╚", "╝", "╠", "╣", "═", "║"),
	"heavy":     box("┏", "┓", "┗", "┛", "┣", "┫", "━", "┃"),
	"dashed":    box("┌", "┐", "└", "┘", "├", "┤", "╌", "╎"),
	"dotted":    box("┌", "┐", "└", "┘", "├", "┤", "┄", "┆"),
	"mixed":     box("╒", "╕", "╘", "╛", "╞", "╡", "═", "│"),
	"block":     frame("▛", "▜", "▙", "▟", "▌", "▐", "▀", "▄", "▀", "▌", "▐"),
	"solid":     box("█", "█", "█", "█", "█", "█", "█", "█"),
	"shade":     box("░", "░", "░", "░", "░", "░", "░", "░"),
	asciiBorder: box("+", "+", "+", "+", "+", "+", "-", "|"),
	"hidden":    box(" ", " ", " ", " ", " ", " ", " ", " "),
}

func glyphs(values ...string) map[string]string {
	return zip(config.GlyphElements, values...)
}

var glyphPresets = map[string]map[string]string{
	"unicode": glyphs(
		"›", "✓", "◐", "!", "✗",
		"○", "❯", "●", "○", "✕", "✗",
		"█", "•", "—", "…", "→", "↑", "↓",
		"<", ">",
	),
	"heavy": glyphs(
		"›", "✔", "◑", "!", "✘",
		"◌", "➜", "◉", "◎", "✘", "✘",
		"▌", "●", "–", "…", "➜", "▲", "▼",
		"◀", "▶",
	),
	"geometric": glyphs(
		"▹", "◆", "◈", "▲", "■",
		"◇", "▶", "◆", "◇", "■", "■",
		"█", "▪", "—", "…", "▸", "▴", "▾",
		"◂", "▸",
	),
	"minimal": glyphs(
		"·", "•", "∘", "!", "×",
		"∘", "▸", "•", "◦", "×", "×",
		"▏", "∙", "–", "…", "›", "˄", "˅",
		"‹", "›",
	),
	"emoji": glyphs(
		"🔹", "✅", "⏳", "❗", "❌",
		"🆕", "👉", "🟢", "⚪", "🔴", "❌",
		"█", "•", "➖", "…", "👉", "🔼", "🔽",
		"⏪", "⏩",
	),
	"nerd": glyphs(
		"\uf105", "\uf00c", "\uf110", "\uf071", "\uf00d",
		"\uf10c", "\uf054", "\uf111", "\uf10c", "\uf057", "\uf00d",
		"█", "•", "—", "…", "\uf061", "\uf062", "\uf063",
		"\uf053", "\uf054",
	),
	config.ASCIIGlyphs: glyphs(
		">", "+", "~", "!", "x",
		"o", ">", "*", "o", "x", "x",
		"_", "*", "-", "~", ">", "^", "v",
		"<", ">",
	),
}

func styles(values ...string) map[string]string {
	return zip(config.TextElements, values...)
}

var textPresets = map[string]map[string]string{
	defaultText:      styles("bold", "bold", "bold", "bold", "bold", "bold", config.PlainText),
	config.PlainText: styles(config.PlainText, config.PlainText, config.PlainText, config.PlainText, config.PlainText, config.PlainText, config.PlainText),
	"emphasis":       styles("bold+underline", "bold", "bold+underline", "bold", "bold", "bold", "italic"),
	colorSoft:        styles("bold", config.PlainText, "bold", config.PlainText, config.PlainText, config.PlainText, "italic+faint"),
}

func colors(elements []string, value string, except map[string]string) map[string]string {
	palette := make(map[string]string, len(elements))
	for _, element := range elements {
		palette[element] = value
		if v, ok := except[element]; ok {
			palette[element] = v
		}
	}
	return palette
}

var glyphTinted = []string{"caret", elDotOn, "import_new", "input_cursor", "hint_arrow", "select_prev", "select_next"}

func tint(value, dotOff string) map[string]string {
	except := map[string]string{elDotOff: dotOff}
	for _, element := range glyphTinted {
		except[element] = value
	}
	return colors(config.GlyphColorElements, config.AutoColor, except)
}

var borderPalettes = map[string]map[string]string{
	"default":   colors(config.BorderElements, config.AutoColor, nil),
	colorAccent: colors(config.BorderElements, colorAccent, nil),
	colorSoft:   colors(config.BorderElements, colorSoft, nil),
	"bright":    colors(config.BorderElements, "fg", nil),
	"duotone":   colors(config.BorderElements, colorAccent, map[string]string{"left": config.AutoColor, "right": config.AutoColor}),
	"sunset":    gradientFrame("amber", "pink", "purple", "blue"),
	"ocean":     gradientFrame("teal", "blue", "purple", "pink"),
}

func gradientFrame(topLeft, topRight, bottomLeft, bottomRight string) map[string]string {
	via := func(from, to string) string { return from + config.GradientSeparator + to }
	return map[string]string{
		"top_left": topLeft, "top": via(topLeft, topRight), "top_right": topRight,
		"left": via(topLeft, bottomLeft), "right": via(topRight, bottomRight),
		"bottom_left": bottomLeft, "bottom": via(bottomLeft, bottomRight), "bottom_right": bottomRight,
		"tee_left": topLeft, "rule": via(topLeft, topRight), "tee_right": topRight,
	}
}

var glyphPalettes = map[string]map[string]string{
	"default":   colors(config.GlyphColorElements, config.AutoColor, nil),
	colorAccent: tint(colorAccent, config.AutoColor),
	"bright":    tint("fg", colorSoft),
	"muted":     tint(colorSoft, "dim"),
}

var textPalettes = map[string]map[string]string{
	defaultPalette: colors(config.TextColorElements, config.AutoColor, nil),
	colorAccent:    colors(config.TextColorElements, config.AutoColor, map[string]string{roleTitle: colorAccent, roleHeading: colorAccent, "key": colorAccent}),
	"muted":        colors(config.TextColorElements, config.AutoColor, map[string]string{roleHeading: colorSoft, "key": colorSoft, "value": colorSoft, "note": "dim"}),
	"sunset":       colors(config.TextColorElements, config.AutoColor, map[string]string{roleTitle: "amber>pink", roleHeading: "amber>pink>purple", "tab": "amber>pink"}),
	"ocean":        colors(config.TextColorElements, config.AutoColor, map[string]string{roleTitle: "teal>blue", roleHeading: "teal>blue>purple", "tab": "teal>blue"}),
}

var accentColors = map[string]color.Color{
	"green":  lipgloss.Color("#63d18a"),
	"lime":   lipgloss.Color("#b6df62"),
	"teal":   lipgloss.Color("#55cec0"),
	"blue":   lipgloss.Color("#55bef0"),
	"purple": lipgloss.Color("#b995f6"),
	"pink":   lipgloss.Color("#ec86cc"),
	"red":    lipgloss.Color("#f47b74"),
	"orange": lipgloss.Color("#fb9d59"),
	"amber":  lipgloss.Color("#eebb58"),
}

func accentColor(value string) color.Color {
	if config.IsHexColor(value) {
		return lipgloss.Color(value)
	}
	if c, ok := accentColors[value]; ok {
		return c
	}
	return accentColors[defaultAccent]
}

func borderLocked(a config.AppearanceSettings) bool {
	return a.Glyphs == config.ASCIIGlyphs
}

func borderRunes(a config.AppearanceSettings) map[string]string {
	if borderLocked(a) {
		return maps.Clone(borderPresets[asciiBorder])
	}
	return resolveRunes(borderPresets, a.Border, defaultBorder, a.BorderRunes)
}

func glyphRunes(a config.AppearanceSettings) map[string]string {
	return resolveRunes(glyphPresets, a.Glyphs, defaultGlyphs, a.GlyphRunes)
}

func borderColorNames(a config.AppearanceSettings) map[string]string {
	return resolveRunes(borderPalettes, a.BorderColor, defaultPalette, a.BorderColors)
}

func glyphColorNames(a config.AppearanceSettings) map[string]string {
	return resolveRunes(glyphPalettes, a.GlyphColor, defaultPalette, a.GlyphColors)
}

func textColorNames(a config.AppearanceSettings) map[string]string {
	return resolveRunes(textPalettes, a.TextColor, defaultPalette, a.TextColors)
}

func colorStops(value string) []color.Color {
	if value == "" || value == config.AutoColor {
		return nil
	}
	parts := strings.Split(value, config.GradientSeparator)
	stops := make([]color.Color, 0, len(parts))
	for _, part := range parts {
		if c := namedColor(part); c != nil {
			stops = append(stops, c)
		}
	}
	return stops
}

func colorOr(c, fallback color.Color) color.Color {
	if c == nil {
		return fallback
	}
	return c
}

func firstStop(value string) color.Color {
	if stops := colorStops(value); len(stops) > 0 {
		return stops[0]
	}
	return nil
}

func namedColor(name string) color.Color {
	switch name {
	case colorAccent:
		return cAccent
	case "border":
		return cBorder
	case "dim":
		return cDim
	case colorSoft:
		return cSoft
	case "fg":
		return cFg
	case "ok":
		return cOK
	case "warn":
		return cAmber
	case "error":
		return cRed
	}
	if config.IsHexColor(name) {
		return lipgloss.Color(name)
	}
	return accentColors[name]
}

func textStyles(a config.AppearanceSettings) map[string]string {
	return resolveRunes(textPresets, a.Text, defaultText, a.TextStyle)
}

func resolveRunes(presets map[string]map[string]string, preset, fallback string, overrides map[string]string) map[string]string {
	base, ok := presets[preset]
	if !ok {
		base = presets[fallback]
	}
	runes := maps.Clone(base)
	maps.Copy(runes, overrides)
	return runes
}

func init() {
	applyAppearance(config.AppearanceSettings{
		Border: defaultBorder, Accent: defaultAccent, Glyphs: defaultGlyphs, Text: defaultText,
		BorderColor: defaultPalette, GlyphColor: defaultPalette, TextColor: defaultPalette,
	})
}

func applyAppearance(a config.AppearanceSettings) {
	cAccent = accentColor(a.Accent)
	for element, value := range borderRunes(a) {
		*borderVars[element] = value
	}
	glyphCell = 1
	for element, value := range glyphRunes(a) {
		*glyphVars[element] = value
		glyphCell = max(glyphCell, ansi.StringWidth(value))
	}
	gradientsOn = a.Gradients
	textColors := textColorNames(a)
	cFg = colorOr(firstStop(textColors["body"]), baseFg)
	cInput = colorOr(firstStop(textColors["input"]), baseInput)
	for element, value := range textStyles(a) {
		*textVars[element] = roleStyle(value, textColors[element])
	}
	borderColors = make(map[string]color.Color, len(config.BorderElements))
	borderGradients = make(map[string][]color.Color)
	for element, name := range borderColorNames(a) {
		borderColors[element] = cBorder
		stops := colorStops(name)
		if len(stops) > 0 {
			borderColors[element] = stops[0]
		}
		if len(stops) > 1 && gradientsOn {
			borderGradients[element] = stops
		}
	}
	glyphColors = make(map[string]color.Color, len(config.GlyphColorElements))
	for element, name := range glyphColorNames(a) {
		if c := namedColor(name); c != nil {
			glyphColors[element] = c
		}
	}
}

func roleStyle(style, colour string) textStyle {
	t := parseTextStyle(style)
	stops := colorStops(colour)
	if len(stops) > 0 {
		t.fg = stops[0]
	}
	if len(stops) > 1 && gradientsOn {
		t.stops = stops
	}
	return t
}

func edge(element, runes string) string {
	return edgeAt(element, runes, 0, ansi.StringWidth(runes))
}

func edgeAt(element, runes string, from, total int) string {
	stops := borderGradients[element]
	if len(stops) < 2 || total < 2 {
		return flatEdge(element, runes)
	}
	colors := lipgloss.Blend1D(total, stops...)
	var b strings.Builder
	at := from
	for _, cell := range graphemes(runes) {
		b.WriteString(lipgloss.NewStyle().Foreground(colors[min(max(at, 0), total-1)]).Render(cell))
		at += ansi.StringWidth(cell)
	}
	return b.String()
}

func gradientText(text string, stops []color.Color, base lipgloss.Style) string {
	cells := graphemes(text)
	if len(cells) < 2 {
		return base.Foreground(stops[0]).Render(text)
	}
	var b strings.Builder
	for i, c := range lipgloss.Blend1D(len(cells), stops...) {
		b.WriteString(base.Foreground(c).Render(cells[i]))
	}
	return b.String()
}

func graphemes(text string) []string {
	var cells []string
	g := uniseg.NewGraphemes(text)
	for g.Next() {
		cells = append(cells, g.Str())
	}
	return cells
}

func sideGradients(lines []string) []string {
	left, right := borderGradients["left"], borderGradients["right"]
	if len(left) < 2 && len(right) < 2 {
		return lines
	}
	lefts, rights := sideColors(left, len(lines)), sideColors(right, len(lines))
	out := make([]string, len(lines))
	for i, line := range lines {
		out[i] = repaintSides(line, lefts[i], rights[i])
	}
	return out
}

func sideColors(stops []color.Color, n int) []color.Color {
	if len(stops) < 2 || n == 0 {
		return make([]color.Color, n)
	}
	return lipgloss.Blend1D(n, stops...)
}

func repaintSides(line string, left, right color.Color) string {
	if prefix := flatEdge("left", gLineLeft); left != nil && strings.HasPrefix(line, prefix) {
		line = lipgloss.NewStyle().Foreground(left).Render(gLineLeft) + strings.TrimPrefix(line, prefix)
	}
	if suffix := flatEdge("right", gLineRight); right != nil && strings.HasSuffix(line, suffix) {
		line = strings.TrimSuffix(line, suffix) + lipgloss.NewStyle().Foreground(right).Render(gLineRight)
	}
	return line
}

func flatEdge(element, runes string) string {
	return lipgloss.NewStyle().Foreground(borderColors[element]).Render(runes)
}

func tinted(element string, contextual color.Color) color.Color {
	if c, ok := glyphColors[element]; ok {
		return c
	}
	return contextual
}

func cell(glyph string) string {
	return glyph + strings.Repeat(" ", max(glyphCell-ansi.StringWidth(glyph), 0))
}

func blankCell() string {
	return strings.Repeat(" ", glyphCell)
}

type textStyle struct {
	bold, italic, underline, faint bool
	fg                             color.Color
	stops                          []color.Color
}

func parseTextStyle(value string) textStyle {
	var t textStyle
	for attr := range strings.SplitSeq(value, "+") {
		switch attr {
		case "bold":
			t.bold = true
		case "italic":
			t.italic = true
		case "underline":
			t.underline = true
		case "faint":
			t.faint = true
		}
	}
	return t
}

func styleIf(on bool, t textStyle) textStyle {
	if on {
		return t
	}
	return textStyle{}
}

func applyConm(c config.Conm) {
	keyMap = newKeymap(c.Keybindings)
	applyAppearance(c.Appearance)
}

func typeColor(t config.ConnType) color.Color {
	switch t {
	case config.PostgresConnType:
		return cPostgres
	case config.MySQLConnType:
		return cMySQL
	case config.MSSQLConnType:
		return cMSSQL
	case config.ClickHouseConnType:
		return cClickHouse
	case config.RedisConnType:
		return cRedis
	case config.MongoDBConnType:
		return cMongoDB
	case config.SSHConnType:
		return cSSH
	default:
		return cAccent
	}
}
