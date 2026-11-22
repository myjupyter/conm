package ui

import (
	"image/color"
	"maps"

	"charm.land/lipgloss/v2"

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
	gTeeL, gTeeR, gLineH, gLineV               string

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
	"tee_left": &gTeeL, "tee_right": &gTeeR, "horizontal": &gLineH, "vertical": &gLineV,
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
)

func box(tl, tr, bl, br, teeL, teeR, h, v string) map[string]string {
	runes := make(map[string]string, len(config.BorderElements))
	for i, value := range []string{tl, tr, bl, br, teeL, teeR, h, v} {
		runes[config.BorderElements[i]] = value
	}
	return runes
}

var borderPresets = map[string]map[string]string{
	"single":    box("┌", "┐", "└", "┘", "├", "┤", "─", "│"),
	"rounded":   box("╭", "╮", "╰", "╯", "├", "┤", "─", "│"),
	"double":    box("╔", "╗", "╚", "╝", "╠", "╣", "═", "║"),
	"heavy":     box("┏", "┓", "┗", "┛", "┣", "┫", "━", "┃"),
	"dashed":    box("┌", "┐", "└", "┘", "├", "┤", "╌", "╎"),
	"dotted":    box("┌", "┐", "└", "┘", "├", "┤", "┄", "┆"),
	"mixed":     box("╒", "╕", "╘", "╛", "╞", "╡", "═", "│"),
	asciiBorder: box("+", "+", "+", "+", "+", "+", "-", "|"),
	"hidden":    box(" ", " ", " ", " ", " ", " ", " ", " "),
}

func glyphs(values ...string) map[string]string {
	runes := make(map[string]string, len(config.GlyphElements))
	for i, value := range values {
		runes[config.GlyphElements[i]] = value
	}
	return runes
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
	applyAppearance(config.AppearanceSettings{Border: defaultBorder, Accent: defaultAccent, Glyphs: defaultGlyphs})
}

func applyAppearance(a config.AppearanceSettings) {
	cAccent = accentColor(a.Accent)
	for element, value := range borderRunes(a) {
		*borderVars[element] = value
	}
	for element, value := range glyphRunes(a) {
		*glyphVars[element] = value
	}
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
