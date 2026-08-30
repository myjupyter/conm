package ui

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// Every screen is a single box of a fixed inner width, drawn line by line: a
// top border carrying the breadcrumb, rules between sections, content lines, a
// status line, the key bar, then the bottom border. This file owns that
// vocabulary and nothing else — the colours it draws with live in theme.go.
//
// Every function takes the inner width w explicitly, so the same primitives
// serve the wide connections table and the narrower forms and keyring screens.

// keybind is one hint in the key bar at the bottom of a screen.
type keybind struct{ key, label string }

// statusReady is the resting status message every screen falls back to.
const statusReady = "ready"

// statusKind picks the glyph and colour of the status line; the message itself
// is free text the screen sets.
type statusKind int

const (
	kindIdle statusKind = iota
	kindOK
	kindPending
	kindWarn
	kindErr
)

// span is a run of text with a single style. Content lines are built from
// spans so a line can mix colours and still be padded to an exact width.
type span struct {
	text string
	fg   color.Color
	bg   color.Color
	bold bool

	// raw is pre-rendered output used for widgets that style themselves (a
	// spinner, say). text is then only there to account for the width.
	raw string
}

func (s span) render() string {
	if s.raw != "" {
		return s.raw
	}
	st := lipgloss.NewStyle().Foreground(s.fg)
	if s.bg != nil {
		st = st.Background(s.bg)
	}
	if s.bold {
		st = st.Bold(true)
	}
	return st.Render(s.text)
}

func (s span) width() int { return len([]rune(s.text)) }

// border paints box-drawing runes in the frame's own colour.
func border(s string) string {
	return lipgloss.NewStyle().Foreground(cBorder).Render(s)
}

// truncPad fits s into exactly w cells, eliding with an ellipsis when it is too
// long and padding on the left instead of the right when right is set.
func truncPad(s string, w int, right bool) string {
	if w <= 0 {
		return ""
	}
	r := []rune(s)
	switch {
	case len(r) > w:
		if w == 1 {
			return string(r[:1])
		}
		return string(r[:w-1]) + gEllipsis
	case right:
		return strings.Repeat(" ", w-len(r)) + s
	default:
		return s + strings.Repeat(" ", w-len(r))
	}
}

// frameTop draws the top border with the app name, the screen's breadcrumb on
// the left and an optional summary pushed against the right corner.
func frameTop(w int, crumb, right string) string {
	left := []span{
		{text: gLineH + " ", fg: cBorder},
		{text: "conm", fg: cFg, bold: true},
		{text: " · " + crumb + " ", fg: cDim},
	}

	used := len([]rune(right))
	for _, s := range left {
		used += s.width()
	}

	var b strings.Builder
	b.WriteString(border(gCornerTL))
	for _, s := range left {
		b.WriteString(s.render())
	}
	b.WriteString(border(strings.Repeat(gLineH, max(w-used, 0))))
	b.WriteString((span{text: right, fg: cDim}).render())
	b.WriteString(border(gCornerTR))
	return b.String()
}

// frameLine draws one content line, padding it out to w so the right border
// always lands in the same column.
func frameLine(w int, spans []span, bg color.Color) string {
	used := 0
	for _, s := range spans {
		used += s.width()
	}
	if used < w {
		spans = append(spans, span{text: strings.Repeat(" ", w-used), bg: bg})
	}

	var b strings.Builder
	b.WriteString(border(gLineV))
	for _, s := range spans {
		b.WriteString(s.render())
	}
	b.WriteString(border(gLineV))
	return b.String()
}

// frameRule draws a horizontal divider; left and right are its end runes.
func frameRule(w int, left, right string) string {
	return border(left) + border(strings.Repeat(gLineH, w)) + border(right)
}

func frameBottom(w int) string { return frameRule(w, gCornerBL, gCornerBR) }

// frameBadge draws a screen's type tag and title under the top border.
func frameBadge(w int, label string, labelBg color.Color, title string) string {
	return frameLine(w, []span{
		{text: "  ", fg: cDim},
		{text: " " + label + " ", fg: cInvFg, bg: labelBg, bold: true},
		{text: "  " + title, fg: cFg},
	}, nil)
}

// frameStatus draws the status line: a glyph, a message truncated to fit, and
// a position counter pinned to the right border.
func frameStatus(w int, icon string, fg color.Color, text, pos string) string {
	textW := max(w-4-len([]rune(pos)), 0)
	return frameLine(w, []span{
		{text: " ", fg: cDim},
		{text: icon, fg: fg},
		{text: " " + truncPad(text, textW, false), fg: fg},
		{text: pos + " ", fg: cMuted},
	}, nil)
}

func statusGlyph(k statusKind) (string, color.Color) {
	switch k {
	case kindOK:
		return gStatusOK, cAccent
	case kindPending:
		return gStatusPending, cAmber
	case kindWarn:
		return gStatusWarn, cAmber
	case kindErr:
		return gStatusErr, cRed
	default:
		return gStatusIdle, cDim
	}
}

// cursorPos formats the "3/12" counter frameStatus ends with.
func cursorPos(cur, n int) string {
	if n == 0 {
		return "0/0"
	}
	return fmt.Sprintf("%d/%d", cur+1, n)
}

// frameKeybar lays out key/label hints, wrapping onto extra lines rather than
// spilling past the right border.
func frameKeybar(w int, binds []keybind) []string {
	// One indent span plus a key/label pair per bind, if nothing wraps.
	newLine := func() []span {
		return append(make([]span, 0, 1+2*len(binds)), span{text: " ", fg: cDim})
	}

	var lines []string
	spans := newLine()
	used := 1
	for i, b := range binds {
		sep := " · "
		if i == len(binds)-1 {
			sep = ""
		}
		seg := b.key + " " + b.label + sep
		if used+len([]rune(seg)) > w && len(spans) > 1 {
			lines = append(lines, frameLine(w, spans, nil))
			spans = newLine()
			used = 1
		}
		spans = append(spans,
			span{text: b.key, fg: cFg, bold: true},
			span{text: " " + b.label + sep, fg: cDim},
		)
		used += len([]rune(seg))
	}
	return append(lines, frameLine(w, spans, nil))
}

func frameSearch(w int, query string, typing bool) string {
	cursor := " "
	if typing {
		cursor = gInputCursor
	}

	room := max(w-3-len([]rune(keyMap.Search.hint)), 0)
	shown := []rune(query)
	if len(shown) > room {
		shown = shown[len(shown)-room:]
	}

	return frameLine(w, []span{
		{text: "  ", fg: cDim},
		{text: keyMap.Search.hint, fg: cAccent, bold: true},
		{text: string(shown), fg: cFg},
		{text: cursor, fg: cAccent},
		{text: strings.Repeat(" ", room-len(shown)), fg: cDim},
	}, nil)
}

func framePong(w int, text string) string {
	return frameLine(w, []span{
		{text: " ", fg: cDim},
		{text: " PONG ", fg: cInvFg, bg: cAccent, bold: true},
		{text: " " + text, fg: cPong},
	}, nil)
}

// frameErrPanel draws the failure panel: a titled rule, the key/value lines
// describing what went wrong, an optional hint and the way out.
func frameErrPanel(w int, e *connError) []string {
	tag := " " + strings.ToUpper(string(e.op)) + " FAILED "
	code := " " + e.code + " "
	right := " " + e.label + " "
	fill := max(w-1-len([]rune(tag))-len([]rune(code))-len([]rune(right)), 0)

	var top strings.Builder
	top.WriteString(border(gTeeL))
	top.WriteString((span{text: gLineH, fg: cErrBorder}).render())
	top.WriteString((span{text: tag, fg: cErrTagFg, bg: cErrTagBg, bold: true}).render())
	top.WriteString((span{text: code, fg: cErrCode, bold: true}).render())
	top.WriteString((span{text: strings.Repeat(gLineH, fill), fg: cErrBorder}).render())
	top.WriteString((span{text: right, fg: cErrMuted}).render())
	top.WriteString(border(gTeeR))

	lines := []string{top.String()}
	if e.target != "" {
		lines = append(lines, frameErrKV(w, "target", e.target, cErrValue))
	}
	if e.during != "" {
		lines = append(lines, frameErrKV(w, "during", e.during, cErrValue))
	}
	lines = append(lines, frameErrKV(w, "error", e.detail, cErrCode))
	if e.hint != "" {
		lines = append(lines, frameLine(w, []span{
			{text: "  ", fg: cDim},
			{text: gHintArrow + " ", fg: cAmber},
			{text: truncPad(e.hint, w-4, false), fg: cHint},
		}, nil))
	}
	return append(lines, frameLine(w, []span{
		{text: "  ", fg: cDim},
		{text: keyMap.Retry.hint, fg: cFg, bold: true},
		{text: " retry · ", fg: cErrMuted},
		{text: keyMap.Cancel.hint, fg: cFg, bold: true},
		{text: " dismiss", fg: cErrMuted},
	}, nil))
}

func frameErrKV(w int, label, value string, valColor color.Color) string {
	return frameLine(w, []span{
		{text: " ", fg: cDim},
		{text: " " + truncPad(label, 8, false), fg: cErrMuted},
		{text: truncPad(value, w-11, false), fg: valColor},
	}, nil)
}

// selectBox is a select field's value as the frame draws it: the value itself,
// its position among the options, and the two states that repaint the box.
type selectBox struct {
	value  string
	at, n  int
	active bool
	bad    bool
}

// frameSelectBox draws a select field's < value > box w cells wide: the value
// centred between the cycle arrows, its position pinned to the right.
func frameSelectBox(w int, s selectBox, bg color.Color) []span {
	borderFg := cBorder
	switch {
	case s.bad:
		borderFg = cRed
	case s.active:
		borderFg = cInvFg
	}

	valFg, posFg := cValue, cMuted
	if s.active {
		valFg, posFg = cInvFg, cInvFg
	}

	pos := fmt.Sprintf("%d/%d", s.at, s.n)
	mid := max(w-8-len([]rune(pos)), 0)
	value := truncPad(s.value, min(len([]rune(s.value)), mid), false)
	left := (mid - len([]rune(value))) / 2

	return []span{
		{text: gLineV + " " + gSelectPrev + " ", fg: borderFg, bg: bg},
		{text: strings.Repeat(" ", left), bg: bg},
		{text: value, fg: valFg, bg: bg, bold: true},
		{text: strings.Repeat(" ", mid-left-len([]rune(value))), bg: bg},
		{text: pos + " ", fg: posFg, bg: bg},
		{text: gSelectNext + " " + gLineV, fg: borderFg, bg: bg},
	}
}
