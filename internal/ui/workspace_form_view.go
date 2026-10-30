package ui

import (
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/ui/spec"
)

func (m workspaceFormModel) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m workspaceFormModel) render() string {
	crumb := "new workspace"
	if m.isEdit {
		crumb = "rename workspace"
	}

	lines := []string{
		frameTop(formInner, crumb, ""),
		frameBadge(formInner, "workspace", cAccent, m.badgeTitle()),
	}

	for _, s := range m.spec.Sections {
		lines = append(lines, frameHead(formInner, s.Title, s.Note)...)
		for i := range s.Fields {
			lines = append(lines, m.workspaceFieldRow(i), m.workspaceMessageRow(i))
		}
	}

	icon, col := statusGlyph(m.statusKind)
	lines = append(lines,
		frameLine(formInner, nil, nil),
		frameRule(formInner, gTeeL, gTeeR),
		frameStatus(formInner, icon, col, m.status, cursorPos(m.idx, len(m.spec.Fields))),
	)
	if m.insert {
		lines = append(lines, keyhintBar(formInner, m.keyhints())...)
	} else {
		lines = append(lines, keyhintLines(formInner, keyhintFooter{groups: m.keyhints(), open: m.help})...)
	}
	lines = append(lines, frameBottom(formInner))
	return strings.Join(lines, "\n")
}

func (m workspaceFormModel) badgeTitle() string {
	if name := strings.TrimSpace(m.vals[spec.WorkspaceNameKey]); m.isEdit && name != "" {
		return "renaming " + name
	}
	return m.title
}

func (m workspaceFormModel) workspaceFieldRow(i int) string {
	f := m.spec.Fields[i]
	active := i == m.idx

	var bg color.Color
	caret, labelFg := caretBlank, cSoft
	if active {
		bg, caret, labelFg = cAccent, " "+gCaret+" ", cInvFg
	}
	gutter, gc := m.workspaceGutter(i, active)

	box := m.workspaceInputBox(i, active, bg)
	spans := make([]span, 0, 3+len(box))
	spans = append(spans,
		span{text: caret, fg: caretColor(active), bg: bg, bold: true},
		span{text: gutter + " ", fg: gc, bg: bg},
		span{text: truncPad(strings.ToLower(f.Label)+" *", formLabelW, false), fg: labelFg, bg: bg, bold: active},
	)
	spans = append(spans, box...)
	return frameLine(formInner, spans, bg)
}

func (m workspaceFormModel) workspaceInputBox(i int, active bool, bg color.Color) []span {
	f := m.spec.Fields[i]

	borderFg := cBorder
	if active {
		borderFg = cInvFg
	}
	if m.attempted && m.fieldError(i) != "" {
		borderFg = cRed
	}

	disp, valFg := m.vals[f.Key], cValue
	if disp == "" && active {
		disp, valFg = f.Example, cFaint
	}
	if active && m.insert {
		disp += gInputCursor
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

func (m workspaceFormModel) workspaceGutter(i int, active bool) (string, color.Color) {
	g, c := gDotOff, cFaint
	switch {
	case m.attempted && m.fieldError(i) != "":
		g, c = gFieldBad, cRed
	case m.vals[m.spec.Fields[i].Key] != "":
		g, c = gDotOn, cAccent
	}
	if active {
		c = cInvFg
	}
	return g, c
}

func (m workspaceFormModel) workspaceMessageRow(i int) string {
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
