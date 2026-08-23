package ui

import (
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/ui/spec"
)

func (m secretFormModel) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m secretFormModel) render() string {
	lines := []string{
		m.secretFormTopLine(),
		m.secretBadgeLine(),
		frameLine(formInner, nil, nil),
	}

	for i := range m.spec.Fields {
		lines = append(lines, m.secretFieldRow(i), m.secretMessageRow(i))
	}

	lines = append(lines,
		frameLine(formInner, nil, nil),
		frameRule(formInner, gTeeL, gTeeR),
		m.secretFormStatusLine(),
	)
	if m.insert {
		lines = append(lines, keyhintBar(formInner, m.keyhints())...)
	} else {
		lines = append(lines, keyhintLines(formInner, keyhintFooter{groups: m.keyhints(), open: m.help})...)
	}
	lines = append(lines, frameBottom(formInner))
	return strings.Join(lines, "\n")
}

func (m secretFormModel) secretFormTopLine() string {
	crumb := "new keyring entry"
	if m.isEdit {
		crumb = "edit keyring entry"
	}
	return frameTop(formInner, crumb, "")
}

func (m secretFormModel) secretBadgeLine() string {
	title := m.title
	if m.isEdit {
		if id := strings.TrimSpace(m.vals[spec.KeyringIDKey]); id != "" {
			title = "editing " + id
		}
	}
	return frameBadge(formInner, "keyring", cAccent, title)
}

func (m secretFormModel) secretFieldRow(i int) string {
	f := m.spec.Fields[i]
	active := i == m.idx

	var bg color.Color
	if active {
		bg = cAccent
	}

	caret := "   "
	if active {
		caret = " " + gCaret + " "
	}
	gutter, gc := m.secretGutter(i, active)

	label := strings.ToLower(f.Label)
	if f.Property == spec.RequiredFieldProperty {
		label += " *"
	}
	labelFg := cSoft
	if active {
		labelFg = cInvFg
	}

	box := m.secretInputBox(i, active, bg)
	spans := make([]span, 0, 3+len(box))
	spans = append(spans,
		span{text: caret, fg: caretColor(active), bg: bg, bold: true},
		span{text: gutter + " ", fg: gc, bg: bg},
		span{text: truncPad(label, formLabelW, false), fg: labelFg, bg: bg, bold: active},
	)
	spans = append(spans, box...)
	return frameLine(formInner, spans, bg)
}

func (m secretFormModel) secretInputBox(i int, active bool, bg color.Color) []span {
	borderFg := cBorder
	if active {
		borderFg = cInvFg
	}
	if m.attempted && m.fieldError(i) != "" {
		borderFg = cRed
	}

	disp, ghost := m.secretDisplay(i, active)
	if active && m.insert {
		disp += gInputCursor
	}
	valFg := cValue
	if ghost {
		valFg = cFaint
	}
	if active {
		valFg = cInvFg
	}
	return []span{
		{text: gLineV, fg: borderFg, bg: bg},
		{text: truncPad(disp, formBoxW-2, false), fg: valFg, bg: bg},
		{text: gLineV, fg: borderFg, bg: bg},
	}
}

func (m secretFormModel) secretDisplay(i int, active bool) (string, bool) {
	f := m.spec.Fields[i]
	raw := m.vals[f.Key]
	if raw != "" {
		if f.Kind == spec.HiddenFieldKind && !m.reveal {
			return strings.Repeat(gInputMask, len([]rune(raw))), false
		}
		return raw, false
	}
	switch {
	// An edit leaves the password blank on purpose: the material is write-only,
	// and an untouched field must not look like an empty password.
	case f.Kind == spec.HiddenFieldKind && m.isEdit:
		return "unchanged", true
	case f.Kind == spec.HiddenFieldKind:
		return "already in the keyring", true
	case active:
		return f.Example, true
	case f.Property == spec.RequiredFieldProperty:
		return "", true
	default:
		return gEmpty, true
	}
}

func (m secretFormModel) secretGutter(i int, active bool) (string, color.Color) {
	f := m.spec.Fields[i]

	var g string
	var c color.Color
	switch {
	case m.attempted && m.fieldError(i) != "":
		g, c = gFieldBad, cRed
	case m.vals[f.Key] != "":
		g, c = gDotOn, cAccent
	default:
		g, c = gDotOff, cFaint
	}
	if active {
		c = cInvFg
	}
	return g, c
}

func (m secretFormModel) secretMessageRow(i int) string {
	text := ""
	if m.attempted {
		if e := m.fieldError(i); e != "" {
			text = gFieldBad + " " + e
		}
	}
	prefix := strings.Repeat(" ", 5+formLabelW+1)
	return frameLine(formInner, []span{
		{text: prefix, fg: cDim},
		{text: truncPad(text, formInner-len([]rune(prefix))-2, false), fg: cRed},
	}, nil)
}

func (m secretFormModel) secretFormStatusLine() string {
	icon, col := statusGlyph(m.statusKind)
	return frameStatus(formInner, icon, col, m.status, cursorPos(m.idx, len(m.spec.Fields)))
}
