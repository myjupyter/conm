package ui

import (
	"fmt"
	"image/color"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/myjupyter/conm/internal/config"
)

type keybind struct{ key, label string }

const statusReady = "ready"

type statusKind int

const (
	kindIdle statusKind = iota
	kindOK
	kindPending
	kindWarn
	kindErr
	kindPing
)

type span struct {
	text string
	fg   color.Color
	bg   color.Color
	bold bool

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

func (s span) width() int { return ansi.StringWidth(s.text) }

func border(s string) string {
	return lipgloss.NewStyle().Foreground(cBorder).Render(s)
}

func truncPad(s string, w int, right bool) string {
	if w <= 0 {
		return ""
	}
	if ansi.StringWidth(s) > w {
		tail := gEllipsis
		if w == 1 {
			tail = ""
		}
		s = ansi.Truncate(s, w, tail)
	}
	padding := strings.Repeat(" ", w-ansi.StringWidth(s))
	if right {
		return padding + s
	}
	return s + padding
}

func clipSpans(spans []span, w int) []span {
	clipped := make([]span, 0, len(spans))
	for _, s := range spans {
		if w <= 0 {
			break
		}
		if s.width() > w {
			s.text, s.raw = ansi.Truncate(s.text, w, gEllipsis), ""
		}
		clipped = append(clipped, s)
		w -= s.width()
	}
	return clipped
}

func spansWidth(spans []span) int {
	used := 0
	for _, s := range spans {
		used += s.width()
	}
	return used
}

func fitNote(note string, w int) string {
	const separator = " · "
	for ansi.StringWidth(note) > w {
		_, rest, ok := strings.Cut(note, separator)
		if !ok {
			return ansi.TruncateLeft(note, ansi.StringWidth(note)-max(w, 0), "")
		}
		note = " " + rest
	}
	return note
}

func frameTop(w int, crumb, right string) string {
	left := []span{
		{text: gLineH + " ", fg: cBorder},
		{text: "conm", fg: cFg, bold: true},
		{text: ":" + config.Workspace(), fg: cAccent},
		{text: " · " + crumb + " ", fg: cDim},
	}

	left = clipSpans(left, w)
	right = fitNote(right, w-spansWidth(left))

	var b strings.Builder
	b.WriteString(border(gCornerTL))
	for _, s := range left {
		b.WriteString(s.render())
	}
	b.WriteString(border(strings.Repeat(gLineH, max(w-spansWidth(left)-ansi.StringWidth(right), 0))))
	b.WriteString((span{text: right, fg: cDim}).render())
	b.WriteString(border(gCornerTR))
	return b.String()
}

func frameLine(w int, spans []span, bg color.Color) string {
	spans = clipSpans(spans, w)
	if used := spansWidth(spans); used < w {
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

func frameHead(w int, title, note string) []string {
	spans := []span{
		{text: "  ", bg: cHeadBg},
		{text: strings.ToUpper(title), fg: cFg, bg: cHeadBg, bold: true},
	}
	if note != "" {
		spans = append(spans, span{text: "  " + note, fg: cFaint, bg: cHeadBg})
	}

	return []string{
		frameLine(w, nil, nil),
		frameLine(w, spans, cHeadBg),
		frameLine(w, nil, nil),
	}
}

func frameTabs(w int, titles []string, active int, note string) string {
	spans := make([]span, 0, 3+2*len(titles))
	spans = append(spans, span{text: " ", fg: cDim})
	used := 1
	for i, title := range titles {
		s := span{text: " " + title + " ", fg: cDim}
		if i == active {
			s = span{text: s.text, fg: cInvFg, bg: cAccent, bold: true}
		}
		spans = append(spans, s, span{text: " ", fg: cDim})
		used += s.width() + 1
	}

	if note != "" {
		spans = append(spans,
			span{text: strings.Repeat(" ", max(w-used-len([]rune(note))-1, 0)), fg: cDim},
			span{text: note + " ", fg: cFaint},
		)
	}

	return frameLine(w, spans, nil)
}

func frameRule(w int, left, right string) string {
	return border(left) + border(strings.Repeat(gLineH, w)) + border(right)
}

func frameBottom(w int) string { return frameRule(w, gCornerBL, gCornerBR) }

func frameBadge(w int, label string, labelBg color.Color, title string) string {
	return frameLine(w, []span{
		{text: "  ", fg: cDim},
		{text: " " + label + " ", fg: cInvFg, bg: labelBg, bold: true},
		{text: "  " + title, fg: cFg},
	}, nil)
}

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

func cursorPos(cur, n int) string {
	if n == 0 {
		return "0/0"
	}
	return fmt.Sprintf("%d/%d", cur+1, n)
}

func frameKeybar(w int, binds []keybind) []string {
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
	return frameTagged(w, " PONG ", cAccent, text)
}

func framePing(w int, text string) string {
	return frameTagged(w, " PING ", cAmber, text)
}

func frameTagged(w int, tag string, bg color.Color, text string) string {
	return frameLine(w, []span{
		{text: " ", fg: cDim},
		{text: tag, fg: cInvFg, bg: bg, bold: true},
		{text: " " + text, fg: cPong},
	}, nil)
}

func frameErrPanel(w int, e *connError) []string {
	tag := " " + strings.ToUpper(string(e.op)) + " FAILED "

	lines := []string{
		frameRule(w, gTeeL, gTeeR),
		frameLine(w, []span{
			{text: " ", fg: cDim},
			{text: tag, fg: cErrTagFg, bg: cErrTagBg, bold: true},
			{text: " " + e.code, fg: cErrCode, bold: true},
		}, nil),
	}
	if e.target != "" {
		lines = append(lines, frameErrKV(w, "target", e.target, cErrValue))
	}
	if e.during != "" {
		lines = append(lines, frameErrKV(w, "during", e.during, cErrValue))
	}
	lines = append(lines, frameErrKVWrapped(w, "error", e.detail, cErrCode)...)
	for i, row := range wrapText(e.hint, w-4) {
		if e.hint == "" {
			break
		}
		arrow := gHintArrow + " "
		if i > 0 {
			arrow = "  "
		}
		lines = append(lines, frameLine(w, []span{
			{text: "  ", fg: cDim},
			{text: arrow, fg: cAmber},
			{text: truncPad(row, w-4, false), fg: cHint},
		}, nil))
	}
	return append(lines, frameLine(w, []span{
		{text: "  ", fg: cDim},
		{text: keyMap.Connection.Retry.hint, fg: cFg, bold: true},
		{text: " retry · ", fg: cErrMuted},
		{text: keyMap.Cancel.hint, fg: cFg, bold: true},
		{text: " dismiss", fg: cErrMuted},
	}, nil))
}

func frameErrKVWrapped(w int, label, value string, valColor color.Color) []string {
	rows := wrapText(value, w-11)
	lines := make([]string, 0, len(rows))
	for i, row := range rows {
		if i > 0 {
			label = ""
		}
		lines = append(lines, frameErrKV(w, label, row, valColor))
	}
	return lines
}

func wrapText(s string, w int) []string {
	if w <= 0 {
		return []string{s}
	}

	var rows []string
	for paragraph := range strings.SplitSeq(s, "\n") {
		row := ""
		for word := range strings.FieldsSeq(paragraph) {
			for len([]rune(word)) > w {
				if row != "" {
					rows = append(rows, row)
					row = ""
				}
				runes := []rune(word)
				rows = append(rows, string(runes[:w]))
				word = string(runes[w:])
			}
			switch {
			case row == "":
				row = word
			case len([]rune(row))+1+len([]rune(word)) <= w:
				row += " " + word
			default:
				rows = append(rows, row)
				row = word
			}
		}
		if row != "" {
			rows = append(rows, row)
		}
	}
	if len(rows) == 0 {
		return []string{""}
	}
	return rows
}

func frameErrKV(w int, label, value string, valColor color.Color) string {
	return frameLine(w, []span{
		{text: " ", fg: cDim},
		{text: " " + truncPad(label, 8, false), fg: cErrMuted},
		{text: truncPad(value, w-11, false), fg: valColor},
	}, nil)
}

type selectBox struct {
	value  string
	at, n  int
	active bool
	bad    bool
}

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

const chipMaxW = formBoxW - 8

type chip struct {
	text    string
	focused bool
}

func fitChip(text string, typing bool) string {
	r := []rune(text)
	switch {
	case len(r) <= chipMaxW:
		return text
	case typing:
		return gEllipsis + string(r[len(r)-chipMaxW+1:])
	default:
		return truncPad(text, chipMaxW, false)
	}
}

func (c chip) label() string {
	if c.focused {
		return "[" + c.text + "]"
	}
	return " " + c.text + " "
}

func chipsWindow(chips []chip, w int) (from, to int) {
	focus := max(slices.IndexFunc(chips, func(c chip) bool { return c.focused }), 0)
	width := func(from, to int) int {
		n := 0
		for _, c := range chips[from:to] {
			n += len([]rune(c.label())) + 1
		}
		return n
	}
	for from < focus && width(from, focus+1) > w {
		from++
	}
	to = from
	for to < len(chips) && width(from, to+1) <= w {
		to++
	}
	return from, max(to, min(focus+1, len(chips)))
}

func frameChips(w int, chips []chip, active bool, bg color.Color) []span {
	if len(chips) == 0 {
		fg := cFaint
		if active {
			fg = cInvFg
		}
		return []span{{text: truncPad(" "+gEmptyList, w, false), fg: fg, bg: bg}}
	}

	fg := cFg
	if active {
		fg = cInvFg
	}

	from, to := chipsWindow(chips, w-5)
	spans := make([]span, 0, 2*(to-from)+4)
	used := 0
	add := func(s span) {
		spans = append(spans, s)
		used += s.width()
	}
	add(span{text: " ", bg: bg})
	if from > 0 {
		add(span{text: gEllipsis + " ", fg: fg, bg: bg})
	}
	for j := from; j < to; j++ {
		if j > from {
			add(span{text: " ", bg: bg})
		}
		add(span{text: chips[j].label(), fg: fg, bg: bg, bold: true})
	}
	if to < len(chips) {
		add(span{text: " " + gEllipsis, fg: fg, bg: bg})
	}
	return append(spans, span{text: strings.Repeat(" ", max(w-used, 0)), bg: bg})
}
