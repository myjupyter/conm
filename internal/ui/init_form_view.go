package ui

import (
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/ui/spec"
)

func (m initFormModel) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m initFormModel) render() string {
	lines := []string{
		frameTop(formInner, "database availability", ""),
		frameBadge(formInner, m.kind.String(), typeColor(m.kind), m.spec.EditTitle),
		m.initFormNoteLine(),
		frameLine(formInner, nil, nil),
	}

	for i := range m.spec.Fields {
		lines = append(lines, m.initFormFieldRow(i), m.initFormMessageRow(i))
	}

	lines = append(lines,
		frameLine(formInner, nil, nil),
		frameRule(formInner, gTeeL, gTeeR),
		m.initFormStatusLine(),
	)
	lines = append(lines, keyhintLines(formInner, keyhintFooter{groups: m.keyhints(), open: m.help})...)
	return strings.Join(append(lines, frameBottom(formInner)), "\n")
}

func (m initFormModel) initFormNoteLine() string {
	note := ""
	if len(m.spec.Sections) > 0 {
		note = m.spec.Sections[0].Note
	}
	return frameLine(formInner, []span{
		{text: "  ", fg: cDim},
		{text: note, fg: cFaint},
	}, nil)
}

func (m initFormModel) initFormFieldRow(i int) string {
	f := m.spec.Fields[i]
	active := i == m.idx
	bad := m.attempted && m.fieldError(i) != ""

	var bg, labelFg, gutterFg color.Color = nil, cSoft, cAccent
	if active {
		bg, labelFg, gutterFg = cAccent, cInvFg, cInvFg
	}

	caret, gutter := "   ", gDotOn
	if active {
		caret = " " + gCaret + " "
	}
	if bad {
		gutter = gFieldBad
		if !active {
			gutterFg = cRed
		}
	}

	label := strings.ToLower(f.Label)
	if f.Property == spec.RequiredFieldProperty {
		label += " *"
	}

	box := m.initFormBox(i, active, bad, bg)
	spans := make([]span, 0, 3+len(box))
	spans = append(spans,
		span{text: caret, fg: caretColor(active), bg: bg, bold: true},
		span{text: gutter + " ", fg: gutterFg, bg: bg},
		span{text: truncPad(label, formLabelW, false), fg: labelFg, bg: bg, bold: active},
	)
	return frameLine(formInner, append(spans, box...), bg)
}

func (m initFormModel) initFormBox(i int, active, bad bool, bg color.Color) []span {
	f := m.spec.Fields[i]

	at := 0
	for j, opt := range f.Options {
		if opt == m.vals[f.Key] {
			at = j + 1
			break
		}
	}

	return frameSelectBox(formBoxW, selectBox{
		value:  m.vals[f.Key],
		at:     at,
		n:      len(f.Options),
		active: active,
		bad:    bad,
	}, bg)
}

func (m initFormModel) initFormMessageRow(i int) string {
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

func (m initFormModel) initFormStatusLine() string {
	icon, col := statusGlyph(m.statusKind)
	return frameStatus(formInner, icon, col, m.status, cursorPos(m.idx, len(m.spec.Fields)))
}
