package ui

import (
	"image/color"
	"maps"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/myjupyter/conm/internal/config"
)

var (
	cBorder     = lipgloss.Color("#3a362e")
	cDim        = lipgloss.Color("#6d6759")
	cMuted      = lipgloss.Color("#5f594c")
	cFaint      = lipgloss.Color("#4d483e")
	cSoft       = lipgloss.Color("#8b8474")
	cFg         = lipgloss.Color("#ddd6c8")
	cDead       = lipgloss.Color("#7d7668")
	cAccent     color.Color
	cOK         = lipgloss.Color("#63d18a")
	cRed        = lipgloss.Color("#e5654a")
	cAmber      = lipgloss.Color("#e3b34a")
	cInvFg      = lipgloss.Color("#0f0f0d")
	cValue      = lipgloss.Color("#c4bcac")
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
	tTitle, tTab, tHeading, tSelection, tKey, tValue, tNote textStyle
)

var textVars = map[string]*textStyle{
	"title": &tTitle, "tab": &tTab, "heading": &tHeading, "selection": &tSelection,
	"key": &tKey, "value": &tValue, "note": &tNote,
}

var glyphVars = map[string]*string{
	"status_idle": &gStatusIdle, "status_ok": &gStatusOK, "status_pending": &gStatusPending,
	"status_warn": &gStatusWarn, "status_error": &gStatusErr, "import_new": &gImportNew,
	"caret": &gCaret, "dot_on": &gDotOn, "dot_off": &gDotOff, "dot_fail": &gDotFail,
	"field_bad": &gFieldBad, "input_cursor": &gInputCursor, "input_mask": &gInputMask,
	"empty": &gEmpty, "ellipsis": &gEllipsis, "hint_arrow": &gHintArrow,
	"scroll_up": &gScrollUp, "scroll_down": &gScrollDown,
	"select_prev": &gSelectPrev, "select_next": &gSelectNext,
}

const (
	defaultBorder = "single"
	asciiBorder   = "ascii"
	defaultGlyphs = "unicode"
	defaultAccent = "green"
	defaultText   = "default"
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
	"soft":           styles("bold", config.PlainText, "bold", config.PlainText, config.PlainText, config.PlainText, "italic+faint"),
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
	for element, value := range textStyles(a) {
		*textVars[element] = parseTextStyle(value)
	}
}

func cell(glyph string) string {
	return glyph + strings.Repeat(" ", max(glyphCell-ansi.StringWidth(glyph), 0))
}

func blankCell() string {
	return strings.Repeat(" ", glyphCell)
}

type textStyle struct {
	bold, italic, underline, faint bool
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
