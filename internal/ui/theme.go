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
	cAccent     = lipgloss.Color("#63d18a")
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

const (
	gCornerTL = "┌"
	gCornerTR = "┐"
	gCornerBL = "└"
	gCornerBR = "┘"
	gTeeL     = "├"
	gTeeR     = "┤"
	gLineH    = "─"
	gLineV    = "│"

	gStatusIdle    = "›"
	gStatusOK      = "✓"
	gStatusPending = "◐"
	gStatusWarn    = "!"
	gStatusErr     = "✗"

	gCaret    = "❯"
	gDotOn    = "●"
	gDotOff   = "○"
	gDotFail  = "✕"
	gFieldBad = "✗"

	gInputCursor = "█"
	gInputMask   = "•"
	gSelectPrev  = "<"
	gSelectNext  = ">"

	gEmpty     = "—"
	gEmptyList = "none"
	gEllipsis  = "…"

	gHintArrow = "→"

	gScrollUp   = "↑"
	gScrollDown = "↓"
)

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
