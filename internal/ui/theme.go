package ui

import (
	"image/color"

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
	gSelectPrev, gSelectNext            string
	gEmpty, gEmptyList, gEllipsis       string
	gHintArrow, gScrollUp, gScrollDown  string
)

type boxRunes struct {
	cornerTL, cornerTR, cornerBL, cornerBR string
	teeL, teeR, lineH, lineV               string
}

var borderStyles = map[string]boxRunes{
	"single":  {"┌", "┐", "└", "┘", "├", "┤", "─", "│"},
	"rounded": {"╭", "╮", "╰", "╯", "├", "┤", "─", "│"},
	"double":  {"╔", "╗", "╚", "╝", "╠", "╣", "═", "║"},
	"heavy":   {"┏", "┓", "┗", "┛", "┣", "┫", "━", "┃"},
}

var asciiBox = boxRunes{"+", "+", "+", "+", "+", "+", "-", "|"}

type glyphRunes struct {
	statusIdle, statusOK, statusPending, statusWarn, statusErr string
	importNew, caret, dotOn, dotOff, dotFail                   string
	fieldBad, inputCursor, inputMask                           string
	empty, ellipsis, hintArrow, scrollUp, scrollDown           string
}

var glyphSets = map[string]glyphRunes{
	"unicode": {
		statusIdle: "›", statusOK: "✓", statusPending: "◐", statusWarn: "!", statusErr: "✗",
		importNew: "○", caret: "❯", dotOn: "●", dotOff: "○", dotFail: "✕",
		fieldBad: "✗", inputCursor: "█", inputMask: "•",
		empty: "—", ellipsis: "…", hintArrow: "→", scrollUp: "↑", scrollDown: "↓",
	},
	"heavy": {
		statusIdle: "›", statusOK: "✔", statusPending: "◑", statusWarn: "!", statusErr: "✘",
		importNew: "◌", caret: "➜", dotOn: "◉", dotOff: "◎", dotFail: "✘",
		fieldBad: "✘", inputCursor: "▌", inputMask: "●",
		empty: "–", ellipsis: "…", hintArrow: "➜", scrollUp: "▲", scrollDown: "▼",
	},
	"geometric": {
		statusIdle: "▹", statusOK: "◆", statusPending: "◈", statusWarn: "▲", statusErr: "■",
		importNew: "◇", caret: "▶", dotOn: "◆", dotOff: "◇", dotFail: "■",
		fieldBad: "■", inputCursor: "█", inputMask: "▪",
		empty: "—", ellipsis: "…", hintArrow: "▸", scrollUp: "▴", scrollDown: "▾",
	},
	"nerd": {
		statusIdle: "\uf105", statusOK: "\uf00c", statusPending: "\uf110", statusWarn: "\uf071", statusErr: "\uf00d",
		importNew: "\uf10c", caret: "\uf054", dotOn: "\uf111", dotOff: "\uf10c", dotFail: "\uf057",
		fieldBad: "\uf00d", inputCursor: "█", inputMask: "•",
		empty: "—", ellipsis: "…", hintArrow: "\uf061", scrollUp: "\uf062", scrollDown: "\uf063",
	},
	config.ASCIIGlyphs: {
		statusIdle: ">", statusOK: "+", statusPending: "~", statusWarn: "!", statusErr: "x",
		importNew: "o", caret: ">", dotOn: "*", dotOff: "o", dotFail: "x",
		fieldBad: "x", inputCursor: "_", inputMask: "*",
		empty: "-", ellipsis: "~", hintArrow: ">", scrollUp: "^", scrollDown: "v",
	},
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
	return accentColors["green"]
}

func boxFor(a config.AppearanceSettings) boxRunes {
	if a.Glyphs == config.ASCIIGlyphs {
		return asciiBox
	}
	if box, ok := borderStyles[a.Border]; ok {
		return box
	}
	return borderStyles["single"]
}

func init() {
	applyAppearance(config.AppearanceSettings{Border: "single", Accent: "green", Glyphs: "unicode"})
}

func applyAppearance(a config.AppearanceSettings) {
	box := boxFor(a)
	glyphs, ok := glyphSets[a.Glyphs]
	if !ok {
		glyphs = glyphSets["unicode"]
	}

	cAccent = accentColor(a.Accent)
	gCornerTL, gCornerTR, gCornerBL, gCornerBR = box.cornerTL, box.cornerTR, box.cornerBL, box.cornerBR
	gTeeL, gTeeR, gLineH, gLineV = box.teeL, box.teeR, box.lineH, box.lineV
	gStatusIdle, gStatusOK, gStatusPending = glyphs.statusIdle, glyphs.statusOK, glyphs.statusPending
	gStatusWarn, gStatusErr, gImportNew = glyphs.statusWarn, glyphs.statusErr, glyphs.importNew
	gCaret, gDotOn, gDotOff, gDotFail = glyphs.caret, glyphs.dotOn, glyphs.dotOff, glyphs.dotFail
	gFieldBad, gInputCursor, gInputMask = glyphs.fieldBad, glyphs.inputCursor, glyphs.inputMask
	gSelectPrev, gSelectNext = "<", ">"
	gEmpty, gEmptyList, gEllipsis = glyphs.empty, "none", glyphs.ellipsis
	gHintArrow, gScrollUp, gScrollDown = glyphs.hintArrow, glyphs.scrollUp, glyphs.scrollDown
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
