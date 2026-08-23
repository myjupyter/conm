package ui

import (
	"slices"
	"strings"

	"github.com/myjupyter/conm/internal/ui/spec"
)

type keyhintGroup struct {
	title string
	binds []keybind
}

const (
	// Group titles shared by every screen's keyhint table.
	keyhintNavigation = "navigation"
	keyhintAction     = "action"
	keyhintSearch     = "search"
	keyhintScreen     = "screen"

	keyhintRows   = 2
	keyhintIndent = 2
	keyhintGap    = 1
)

type keyhintFooter struct {
	groups []keyhintGroup
	open   bool
	search bool
	typing bool
	query  string
	pos    string
}

func keyhintLines(w int, f keyhintFooter) []string {
	var lines []string
	if f.open {
		lines = append(keyhintTable(w, f.groups), frameRule(w, gTeeL, gTeeR))
	}
	if f.search {
		lines = append(lines, frameSearch(w, f.query, f.typing), frameRule(w, gTeeL, gTeeR))
	}
	return append(lines, keyhintPrompt(w, f))
}

func keyhintPrompt(w int, f keyhintFooter) string {
	spans := keyhintPromptBinds(f)
	if f.pos != "" {
		used := 0
		for _, s := range spans {
			used += s.width()
		}
		spans = append(spans,
			span{text: strings.Repeat(" ", max(w-used-len([]rune(f.pos))-1, 0)), fg: cDim},
			span{text: f.pos + " ", fg: cMuted},
		)
	}
	return frameLine(w, spans, nil)
}

func keyhintPromptBinds(f keyhintFooter) []span {
	spans := make([]span, 0, 6)
	if f.typing {
		spans = append(spans,
			span{text: " ", fg: cDim},
			span{text: keyMap.Confirm.hint, fg: cFg, bold: true},
			span{text: " keep · ", fg: cDim},
			span{text: keyMap.Cancel.hint, fg: cFg, bold: true},
			span{text: " clear", fg: cDim},
		)
	} else {
		label := " help"
		if f.open {
			label = " exit from help"
		}
		spans = append(spans,
			span{text: " ", fg: cDim},
			span{text: keyMap.Help.hint, fg: cAccent, bold: true},
			span{text: label, fg: cDim},
		)
	}

	if !hasKeyhintGroup(f.groups, keyhintSearch) {
		return spans
	}

	label := " search"
	if f.typing {
		label = " exit from search"
	}
	return append(spans,
		span{text: " · ", fg: cFaint},
		span{text: keyMap.Search.hint, fg: cAccent, bold: true},
		span{text: label, fg: cDim},
	)
}

func hasKeyhintGroup(groups []keyhintGroup, title string) bool {
	return slices.ContainsFunc(groups, func(g keyhintGroup) bool { return g.title == title })
}

func searchKeyhints(label string, filtered bool) []keybind {
	binds := []keybind{{keyMap.Search.hint, label}}
	if filtered {
		binds = append(binds, keybind{keyMap.Cancel.hint, "clear the filter"})
	}
	return binds
}

// initKeyhintLine spells the same hints for the init screens, which are not
// framed and so get one plain line instead of a key bar.
func initKeyhintLine() string {
	binds := []keybind{
		{keyMap.Up.hint, "up"},
		{keyMap.Down.hint, "down"},
		{keyMap.Confirm.hint, "select"},
		{keyMap.Quit.hint, "quit"},
	}

	parts := make([]string, 0, len(binds))
	for _, b := range binds {
		parts = append(parts, b.key+" "+b.label)
	}
	return strings.Join(parts, " · ")
}

func keyhintBar(w int, groups []keyhintGroup) []string {
	return frameKeybar(w, flattenKeyhints(groups))
}

func keyhintStatus(open bool) string {
	if open {
		return "all keys for this screen · " + keyMap.Help.hint + " to close"
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
				{keyMap.Up.hint, "up"},
				{keyMap.Down.hint, "down"},
			},
		},
		{
			title: keyhintAction,
			binds: []keybind{
				{keyMap.Confirm.hint, "connect"},
				{keyMap.Ping.hint, "ping"},
				{keyMap.Add.hint, "add new"},
				{keyMap.Edit.hint, "edit"},
				{keyMap.Delete.hint, "delete"},
			},
		},
		{
			title: keyhintSearch,
			binds: searchKeyhints("filter connections", m.conns.Filtered()),
		},
		{
			title: keyhintScreen,
			binds: []keybind{
				{keyMap.Secret.hint, "secrets"},
				{keyMap.Quit.hint, "quit"},
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
					{keyMap.Cancel.hint, "stop editing"},
					{keyMap.Confirm.hint, "confirm and next"},
					{keyMap.NextSection.hint, "switch section"},
				},
			},
		}
	}

	cur := m.currentField()

	nav := []keybind{{keyMap.MoveVertical.hint, "move between fields"}, {keyMap.NextSection.hint, "switch section"}}
	switch {
	case m.isSelector(cur):
		nav = append(nav, keybind{keyMap.Cycle.hint, "change a selector"})
	case m.isSecretField(cur) && m.isRef():
		nav = append(nav, keybind{keyMap.Cycle.hint, "pick an entry"})
	}

	action := []keybind{{keyMap.Edit.hint, "edit field"}}
	switch {
	case m.isProviderField(cur) && m.isRef():
		action = append(action, keybind{keyMap.Secret.hint, "open " + m.storeLabel()})
	case m.spec.Fields[cur].Kind == spec.HiddenFieldKind && !m.isRef():
		label := "show password"
		if m.reveal {
			label = "hide password"
		}
		action = append(action, keybind{keyMap.Secret.hint, label})
	}
	action = append(action,
		keybind{keyMap.Ping.hint, "test the connection"},
		keybind{keyMap.Confirm.hint, "validate and save"},
	)

	return []keyhintGroup{
		{title: keyhintNavigation, binds: nav},
		{title: keyhintAction, binds: action},
		{title: keyhintScreen, binds: []keybind{{keyMap.Cancel.hint, "cancel"}}},
	}
}

func (m secretModel) keyhints() []keyhintGroup {
	use, back := "show usage", "back to connections"
	if m.secrets.Picking() {
		use, back = "attach to connection", "back to form"
	}
	return []keyhintGroup{
		{title: keyhintNavigation, binds: []keybind{
			{keyMap.MoveVertical.hint, "move"},
			{keyMap.SwitchStore.hint, "switch store"},
		}},
		{title: keyhintAction, binds: []keybind{
			{keyMap.Confirm.hint, use},
			{keyMap.Add.hint, "add new"},
			{keyMap.Edit.hint, "edit entry"},
			{keyMap.Delete.hint, "delete entry"},
		}},
		{title: keyhintSearch, binds: searchKeyhints("filter entries", m.secrets.Filtered())},
		{title: keyhintScreen, binds: []keybind{{keyMap.Cancel.hint, back}}},
	}
}

func (m secretFormModel) keyhints() []keyhintGroup {
	if m.insert {
		return []keyhintGroup{
			{title: "editing", binds: []keybind{
				{keyMap.Cancel.hint, "stop editing"},
				{keyMap.Confirm.hint, "confirm and next"},
			}},
		}
	}

	action := []keybind{{keyMap.Edit.hint, "edit field"}}
	if m.spec.Fields[m.idx].Kind == spec.HiddenFieldKind {
		label := "show password"
		if m.reveal {
			label = "hide password"
		}
		action = append(action, keybind{keyMap.Secret.hint, label})
	}
	action = append(action, keybind{keyMap.Confirm.hint, "store entry"})

	return []keyhintGroup{
		{title: keyhintNavigation, binds: []keybind{{keyMap.MoveVertical.hint, "move between fields"}}},
		{title: keyhintAction, binds: action},
		{title: keyhintScreen, binds: []keybind{{keyMap.Cancel.hint, "cancel"}}},
	}
}
