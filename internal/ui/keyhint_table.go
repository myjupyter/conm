package ui

import (
	"strings"

	"github.com/myjupyter/conm/internal/ui/spec"
)

type keyhintGroup struct {
	title string
	binds []keybind
}

const (
	// keyhintKey is matched, keyhintKeyLabel is shown: a terminal reports the
	// character it produced, "?", never a "shift+…" keystroke.
	keyhintKey      = "?"
	keyhintKeyLabel = "shift + ?"

	// Group titles shared by every screen's keyhint table.
	keyhintNavigation = "navigation"
	keyhintAction     = "action"
	keyhintScreen     = "screen"

	keyhintRows   = 2
	keyhintIndent = 2
	keyhintGap    = 1
)

func keyhintLines(w int, groups []keyhintGroup, open bool) []string {
	if !open {
		return []string{keyhintPrompt(w)}
	}
	return keyhintTable(w, groups)
}

func keyhintPrompt(w int) string {
	return frameLine(w, []span{
		{text: " ", fg: cDim},
		{text: keyhintKeyLabel, fg: cAccent, bold: true},
		{text: " help", fg: cDim},
	}, nil)
}

func keyhintBar(w int, groups []keyhintGroup) []string {
	return frameKeybar(w, flattenKeyhints(groups))
}

func keyhintStatus(open bool) string {
	if open {
		return "all keys for this screen · " + keyhintKeyLabel + " to close"
	}
	return statusReady
}

func keyhintTable(w int, groups []keyhintGroup) []string {
	keyW := 0
	for _, g := range groups {
		for _, b := range g.binds {
			keyW = max(keyW, len([]rune(b.key)))
		}
	}

	var lines []string
	for i, g := range groups {
		if i > 0 {
			lines = append(lines, frameLine(w, nil, nil))
		}
		lines = append(lines, frameLine(w, []span{
			{text: "  ", fg: cDim},
			{text: strings.ToUpper(g.title), fg: cSoft, bold: true},
		}, nil))
		lines = append(lines, keyhintRowLines(w, g.binds, keyW)...)
	}
	return lines
}

func keyhintRowLines(w int, binds []keybind, keyW int) []string {
	if len(binds) == 0 {
		return nil
	}
	rows := min(len(binds), keyhintRows)
	cols := (len(binds) + keyhintRows - 1) / keyhintRows
	colW := (w - keyhintIndent) / cols
	labelW := max(colW-keyW-keyhintGap, 0)

	lines := make([]string, 0, rows)
	for r := range rows {
		spans := []span{{text: strings.Repeat(" ", keyhintIndent), fg: cDim}}
		for c := range cols {
			i := c*keyhintRows + r
			if i >= len(binds) {
				spans = append(spans, span{text: strings.Repeat(" ", colW), fg: cDim})
				continue
			}
			spans = append(spans,
				span{text: truncPad(binds[i].key, keyW, true), fg: cFg, bold: true},
				span{text: strings.Repeat(" ", keyhintGap) + truncPad(binds[i].label, labelW, false), fg: cDim},
			)
		}
		lines = append(lines, frameLine(w, spans, nil))
	}
	return lines
}

func flattenKeyhints(groups []keyhintGroup) []keybind {
	var binds []keybind
	for _, g := range groups {
		binds = append(binds, g.binds...)
	}
	return binds
}

func (m Model) keyhints() []keyhintGroup {
	return []keyhintGroup{
		{
			title: keyhintNavigation,
			binds: []keybind{
				{"↑/k", "up"},
				{"↓/j", "down"},
			},
		},
		{
			title: keyhintAction,
			binds: []keybind{
				{"enter", "connect"},
				{"p", "ping"},
				{"a", "add new"},
				{"e", "edit"},
				{"d", "delete"},
			},
		},
		{
			title: keyhintScreen,
			binds: []keybind{
				{"s", "secrets"},
				{"q/esc", "quit"},
			},
		},
	}
}

func (m formModel) keyhints() []keyhintGroup {
	if m.insert {
		return []keyhintGroup{
			{
				title: "editing",
				binds: []keybind{
					{"esc", "stop editing"},
					{"enter", "confirm and next"},
					{"tab", "switch section"},
				},
			},
		}
	}

	cur := m.currentField()

	nav := []keybind{{"↑↓/jk", "move between fields"}, {"tab", "switch section"}}
	switch {
	case m.isSelector(cur):
		nav = append(nav, keybind{"←/→", "change a selector"})
	case m.isSecretField(cur) && m.isRef():
		nav = append(nav, keybind{"←/→", "pick an entry"})
	}

	action := []keybind{{"e", "edit field"}}
	switch {
	case m.isProviderField(cur) && m.isRef():
		action = append(action, keybind{"s", "open " + m.storeLabel()})
	case m.spec.Fields[cur].Kind == spec.HiddenFieldKind && !m.isRef():
		label := "show password"
		if m.reveal {
			label = "hide password"
		}
		action = append(action, keybind{"s", label})
	}
	action = append(action,
		keybind{"p", "test the connection"},
		keybind{"enter", "validate and save"},
	)

	return []keyhintGroup{
		{title: keyhintNavigation, binds: nav},
		{title: keyhintAction, binds: action},
		{title: keyhintScreen, binds: []keybind{{"esc", "cancel"}}},
	}
}

func (m secretModel) keyhints() []keyhintGroup {
	use, back := "show usage", "back to connections"
	if m.picking {
		use, back = "attach to connection", "back to form"
	}
	return []keyhintGroup{
		{title: keyhintNavigation, binds: []keybind{
			{"↑↓/jk", "move"},
			{"tab/←→", "switch store"},
		}},
		{title: keyhintAction, binds: []keybind{
			{"enter", use},
			{"a", "add new"},
			{"e", "edit entry"},
			{"d", "delete entry"},
		}},
		{title: keyhintScreen, binds: []keybind{{"esc", back}}},
	}
}

func (m secretFormModel) keyhints() []keyhintGroup {
	if m.insert {
		return []keyhintGroup{
			{title: "editing", binds: []keybind{
				{"esc", "stop editing"},
				{"enter", "confirm and next"},
			}},
		}
	}

	action := []keybind{{"e", "edit field"}}
	if m.spec.Fields[m.idx].Kind == spec.HiddenFieldKind {
		label := "show password"
		if m.reveal {
			label = "hide password"
		}
		action = append(action, keybind{"s", label})
	}
	action = append(action, keybind{"enter", "store entry"})

	return []keyhintGroup{
		{title: keyhintNavigation, binds: []keybind{{"↑↓/jk", "move between fields"}}},
		{title: keyhintAction, binds: action},
		{title: keyhintScreen, binds: []keybind{{"esc", "cancel"}}},
	}
}
