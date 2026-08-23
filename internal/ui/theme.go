package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"

	"github.com/myjupyter/conm/internal/config"
)

// The palette every screen draws with. It is kept apart from frame.go on
// purpose: frame.go decides how a line is assembled, this file decides what it
// looks like, so restyling the TUI never means touching the drawing code.
var (
	cBorder   = lipgloss.Color("#3a362e")
	cDim      = lipgloss.Color("#6d6759")
	cMuted    = lipgloss.Color("#5f594c")
	cFaint    = lipgloss.Color("#4d483e")
	cSoft     = lipgloss.Color("#8b8474")
	cFg       = lipgloss.Color("#ddd6c8")
	cDead     = lipgloss.Color("#7d7668")
	cAccent   = lipgloss.Color("#63d18a")
	cRed      = lipgloss.Color("#e5654a")
	cAmber    = lipgloss.Color("#e3b34a")
	cInvFg    = lipgloss.Color("#0f0f0d")
	cValue    = lipgloss.Color("#c4bcac")
	cPong     = lipgloss.Color("#b9cdbd")
	cPostgres = lipgloss.Color("#5aa0d6")

	cErrBorder = lipgloss.Color("#7a3f34")
	cErrTagBg  = lipgloss.Color("#9e3b2a")
	cErrTagFg  = lipgloss.Color("#fdf3f1")
	cErrCode   = lipgloss.Color("#f0a48f")
	cErrMuted  = lipgloss.Color("#8d7a76")
	cErrValue  = lipgloss.Color("#c8b8b4")
	cHint      = lipgloss.Color("#a2938f")
)

// The runes every screen is drawn from. Like the palette above, they say what
// the TUI looks like, not how a line is put together — swapping this block for
// a rounded or an ASCII set restyles the whole app without touching frame.go.
//
// Prose keeps its own punctuation: the "·" inside a status message is copy, not
// chrome, and reads better written out where it is used. The arrows inside a
// key hint ("↑/k", "←/→") are neither — they spell a key, so they come from the
// binding's hint in keybinding.go.
const (
	// The box every screen is framed in.
	gCornerTL = "┌"
	gCornerTR = "┐"
	gCornerBL = "└"
	gCornerBR = "┘"
	gTeeL     = "├"
	gTeeR     = "┤"
	gLineH    = "─"
	gLineV    = "│"

	// The status line's leading glyph, one per statusKind.
	gStatusIdle    = "›"
	gStatusOK      = "✓"
	gStatusPending = "◐"
	gStatusWarn    = "!"
	gStatusErr     = "✗"

	// Row and field markers: the cursor, then the state dot a connection, a
	// secret or a form field carries.
	gCaret    = "❯"
	gDotOn    = "●"
	gDotOff   = "○"
	gDotFail  = "✕"
	gFieldBad = "✗"

	// Inside an input: the caret while typing, the character a hidden value is
	// masked with, and the two ends of a select field's ‹ value › chrome.
	gInputCursor = "█"
	gInputMask   = "•"
	gSelectPrev  = "<"
	gSelectNext  = ">"

	// Stand-ins: an empty value, and the elision truncPad ends a clipped string
	// with.
	gEmpty    = "—"
	gEllipsis = "…"

	// The arrow the error panel points its hint with.
	gHintArrow = "→"
)

// typeColor is the accent a connection type is branded with, used for its tab
// and for the badge above a form.
func typeColor(t config.ConnType) color.Color {
	switch t {
	case config.PostgresConnType:
		return cPostgres
	default:
		return cAccent
	}
}
