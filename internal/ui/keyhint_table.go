package ui

import (
	"slices"
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/myjupyter/conm/internal/i18n"

	"github.com/myjupyter/conm/internal/ui/spec"
)

type keyhintGroup struct {
	title string
	binds []keybind
}

var (
	keyhintNavigation = i18n.Sprintf("navigation")
	keyhintAction     = i18n.Sprintf("action")
	keyhintSearch     = i18n.Sprintf("search")
	keyhintScreen     = i18n.Sprintf("screen")

	keyhintMoveFields = i18n.Sprintf("move between fields")
	keyhintBackToList = i18n.Sprintf("back to connections")
)

const (
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
			span{text: strings.Repeat(" ", max(w-used-ansi.StringWidth(f.pos)-1, 0)), fg: cDim},
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
			span{text: " " + i18n.Sprintf("keep") + " · ", fg: cDim},
			span{text: keyMap.Cancel.hint, fg: cFg, bold: true},
			span{text: " " + i18n.Sprintf("clear"), fg: cDim},
		)
	} else {
		label := " " + i18n.Sprintf("help")
		if f.open {
			label = " " + i18n.Sprintf("exit from help")
		}
		spans = append(spans,
			span{text: " ", fg: cDim},
			span{text: keyMap.Help.hint, fg: cAccent, bold: true},
			span{text: label, fg: cDim},
		)
	}

	if f.typing || !hasKeyhintGroup(f.groups, keyhintSearch) {
		return spans
	}

	spans = append(spans,
		span{text: " · ", fg: cFaint},
		span{text: keyMap.Search.hint, fg: cAccent, bold: true},
		span{text: " " + i18n.Sprintf("search"), fg: cDim},
	)

	if !f.search {
		return spans
	}

	return append(spans,
		span{text: " · ", fg: cFaint},
		span{text: keyMap.Cancel.hint, fg: cAccent, bold: true},
		span{text: " " + i18n.Sprintf("clear filter"), fg: cDim},
	)
}

func hasKeyhintGroup(groups []keyhintGroup, title string) bool {
	return slices.ContainsFunc(groups, func(g keyhintGroup) bool { return g.title == title })
}

func searchKeyhints(label string, filtered bool) []keybind {
	binds := []keybind{{keyMap.Search.hint, label}}
	if filtered {
		binds = append(binds, keybind{keyMap.Cancel.hint, i18n.Sprintf("clear the filter")})
	}
	return binds
}

func keyhintBar(w int, groups []keyhintGroup) []string {
	return frameKeybar(w, flattenKeyhints(groups))
}

func keyhintStatus(open bool) string {
	if open {
		return i18n.Sprintf("all keys for this screen · %s to close", keyMap.Help.hint)
	}
	return statusReady
}

func keyhintTable(w int, groups []keyhintGroup) []string {
	keyW := 0
	for _, g := range groups {
		for _, b := range g.binds {
			keyW = max(keyW, ansi.StringWidth(b.key))
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
				{keyMap.Up.hint, i18n.Sprintf("up")},
				{keyMap.Down.hint, i18n.Sprintf("down")},
				{keyMap.SwitchKind.hint, i18n.Sprintf("switch database")},
			},
		},
		{
			title: keyhintAction,
			binds: []keybind{
				{keyMap.Confirm.hint, i18n.Sprintf("connect")},
				{keyMap.Ping.hint, i18n.Sprintf("ping")},
				{keyMap.Add.hint, i18n.Sprintf("add new")},
				{keyMap.AddDB.hint, i18n.Sprintf("add new database")},
				{keyMap.Edit.hint, i18n.Sprintf("edit")},
				{keyMap.Delete.hint, i18n.Sprintf("delete")},
			},
		},
		{
			title: keyhintSearch,
			binds: searchKeyhints(i18n.Sprintf("filter connections"), m.conns.Filtered()),
		},
		{
			title: keyhintScreen,
			binds: []keybind{
				{keyMap.Info.hint, i18n.Sprintf("info")},
				{keyMap.Secret.hint, i18n.Sprintf("secrets")},
				{keyMap.Transport.hint, i18n.Sprintf("transports")},
				{keyMap.Quit.hint, i18n.Sprintf("quit")},
			},
		},
	}
}

func (m infoModel) keyhints() []keyhintGroup {
	return []keyhintGroup{
		{title: keyhintNavigation, binds: []keybind{
			{keyMap.MoveVertical.hint, keyhintMoveFields},
			{keyMap.SwitchKind.hint, strings.Join(m.tabTitles(), " · ")},
		}},
		{title: keyhintAction, binds: []keybind{
			{keyMap.Yank.hint, i18n.Sprintf("yank the value")},
			{keyMap.Confirm.hint, i18n.Sprintf("open the link · connect")},
			{keyMap.Edit.hint, i18n.Sprintf("edit")},
		}},
		{title: keyhintScreen, binds: []keybind{{keyMap.Back.hint, keyhintBackToList}}},
	}
}

func (m formModel) keyhints() []keyhintGroup {
	if m.insert {
		return []keyhintGroup{
			{
				title: i18n.Sprintf("editing"),
				binds: []keybind{
					{keyMap.Cancel.hint, i18n.Sprintf("stop editing")},
					{keyMap.Confirm.hint, i18n.Sprintf("confirm and next")},
					{keyMap.Paste.hint, i18n.Sprintf("paste")},
					{keyMap.NextSection.hint, i18n.Sprintf("switch section")},
				},
			},
		}
	}

	cur := m.currentField()

	nav := []keybind{{keyMap.MoveVertical.hint, keyhintMoveFields}, {keyMap.NextSection.hint, i18n.Sprintf("switch section")}}
	switch {
	case m.isSelector(cur):
		nav = append(nav, keybind{keyMap.Cycle.hint, i18n.Sprintf("change a selector")})
	case m.isSecretField(cur) && m.isRef():
		nav = append(nav, keybind{keyMap.Cycle.hint, i18n.Sprintf("pick an entry")})
	}

	action := []keybind{{keyMap.Edit.hint, i18n.Sprintf("edit field")}}
	switch {
	case m.isProviderField(cur) && m.isRef():
		action = append(action, keybind{keyMap.Secret.hint, i18n.Sprintf("open %s", m.storeLabel())})
	case m.fields[cur].Kind == spec.HiddenFieldKind && m.isLiteral():
		label := i18n.Sprintf("show password")
		if m.reveal {
			label = i18n.Sprintf("hide password")
		}
		action = append(action, keybind{keyMap.Secret.hint, label})
	}
	if spec.IsLinkKey(m.fields[cur].Key) {
		action = append(action,
			keybind{keyMap.Confirm.hint, i18n.Sprintf("open the link")},
			keybind{keyMap.Delete.hint, i18n.Sprintf("remove the link")},
		)
	}
	if m.hasLinks() && m.section == m.linkSect {
		action = append(action, keybind{keyMap.Add.hint, i18n.Sprintf("add a link")})
	}

	action = append(action,
		keybind{keyMap.Ping.hint, i18n.Sprintf("test the connection")},
	)
	if !spec.IsLinkKey(m.fields[cur].Key) {
		action = append(action, keybind{keyMap.Confirm.hint, i18n.Sprintf("validate and save")})
	}

	return []keyhintGroup{
		{title: keyhintNavigation, binds: nav},
		{title: keyhintAction, binds: action},
		{title: keyhintScreen, binds: []keybind{{keyMap.Cancel.hint, i18n.Sprintf("cancel")}}},
	}
}

func (m secretModel) keyhints() []keyhintGroup {
	use, back := i18n.Sprintf("show usage"), keyhintBackToList
	if m.secrets.Picking() {
		use, back = i18n.Sprintf("attach to connection"), i18n.Sprintf("back to form")
	}
	return []keyhintGroup{
		{title: keyhintNavigation, binds: []keybind{
			{keyMap.MoveVertical.hint, i18n.Sprintf("move")},
			{keyMap.SwitchKind.hint, i18n.Sprintf("switch store")},
		}},
		{title: keyhintAction, binds: []keybind{
			{keyMap.Confirm.hint, use},
			{keyMap.Add.hint, i18n.Sprintf("add new")},
			{keyMap.Edit.hint, i18n.Sprintf("edit entry")},
			{keyMap.Delete.hint, i18n.Sprintf("delete entry")},
		}},
		{title: keyhintSearch, binds: searchKeyhints(i18n.Sprintf("filter entries"), m.secrets.Filtered())},
		{title: keyhintScreen, binds: []keybind{{keyMap.Cancel.hint, back}}},
	}
}

func (m secretFormModel) keyhints() []keyhintGroup {
	if m.insert {
		return []keyhintGroup{
			{title: i18n.Sprintf("editing"), binds: []keybind{
				{keyMap.Cancel.hint, i18n.Sprintf("stop editing")},
				{keyMap.Confirm.hint, i18n.Sprintf("confirm and next")},
				{keyMap.Paste.hint, i18n.Sprintf("paste")},
			}},
		}
	}

	action := []keybind{{keyMap.Edit.hint, i18n.Sprintf("edit field")}}
	if m.spec.Fields[m.idx].Kind == spec.HiddenFieldKind {
		label := i18n.Sprintf("show password")
		if m.reveal {
			label = i18n.Sprintf("hide password")
		}
		action = append(action, keybind{keyMap.Secret.hint, label})
	}
	action = append(action, keybind{keyMap.Confirm.hint, i18n.Sprintf("store entry")})

	return []keyhintGroup{
		{title: keyhintNavigation, binds: []keybind{{keyMap.MoveVertical.hint, keyhintMoveFields}}},
		{title: keyhintAction, binds: action},
		{title: keyhintScreen, binds: []keybind{{keyMap.Cancel.hint, i18n.Sprintf("cancel")}}},
	}
}

func (m initModel) keyhints() []keyhintGroup {
	confirm, leave := i18n.Sprintf("open the connections"), keybind{keyMap.Quit.hint, i18n.Sprintf("quit")}
	if m.adding {
		confirm, leave = keyhintBackToList, keybind{keyMap.Cancel.hint, keyhintBackToList}
	}

	return []keyhintGroup{
		{title: keyhintNavigation, binds: []keybind{{keyMap.MoveVertical.hint, i18n.Sprintf("move")}}},
		{title: keyhintAction, binds: []keybind{
			{keyMap.Toggle.hint, i18n.Sprintf("enable / disable")},
			{keyMap.Confirm.hint, confirm},
			{keyMap.Edit.hint, i18n.Sprintf("edit")},
		}},
		{title: keyhintScreen, binds: []keybind{leave}},
	}
}

func (m initFormModel) keyhints() []keyhintGroup {
	return []keyhintGroup{
		{title: keyhintNavigation, binds: []keybind{
			{keyMap.MoveVertical.hint, keyhintMoveFields},
			{keyMap.Cycle.hint, i18n.Sprintf("change the value")},
		}},
		{title: keyhintAction, binds: []keybind{{keyMap.Confirm.hint, i18n.Sprintf("save and go back")}}},
		{title: keyhintScreen, binds: []keybind{{keyMap.Cancel.hint, i18n.Sprintf("back without saving")}}},
	}
}

func (m transportModel) keyhints() []keyhintGroup {
	return []keyhintGroup{
		{title: keyhintNavigation, binds: []keybind{
			{keyMap.MoveVertical.hint, i18n.Sprintf("move")},
		}},
		{title: keyhintAction, binds: []keybind{
			{keyMap.Confirm.hint, i18n.Sprintf("show usage")},
			{keyMap.Add.hint, i18n.Sprintf("add new")},
			{keyMap.Edit.hint, i18n.Sprintf("edit transport")},
			{keyMap.Delete.hint, i18n.Sprintf("delete transport")},
		}},
		{title: keyhintSearch, binds: searchKeyhints(i18n.Sprintf("filter transports"), m.transports.Filtered())},
		{title: keyhintScreen, binds: []keybind{{keyMap.Cancel.hint, keyhintBackToList}}},
	}
}

func (m transportFormModel) keyhints() []keyhintGroup {
	if m.insert {
		return []keyhintGroup{
			{
				title: i18n.Sprintf("editing"),
				binds: []keybind{
					{keyMap.Cancel.hint, i18n.Sprintf("stop editing")},
					{keyMap.Confirm.hint, i18n.Sprintf("confirm and next")},
					{keyMap.Paste.hint, i18n.Sprintf("paste")},
				},
			},
		}
	}

	nav := []keybind{{keyMap.MoveVertical.hint, keyhintMoveFields}}
	if cur := m.current(); isSelect(cur) && len(m.options(cur)) > 0 {
		nav = append(nav, keybind{keyMap.Cycle.hint, i18n.Sprintf("change a selector")})
	}

	return []keyhintGroup{
		{title: keyhintNavigation, binds: nav},
		{title: keyhintAction, binds: []keybind{
			{keyMap.Edit.hint, i18n.Sprintf("edit field")},
			{keyMap.Ping.hint, i18n.Sprintf("ping the connection it goes through")},
			{keyMap.Confirm.hint, i18n.Sprintf("validate and save")},
		}},
		{title: keyhintScreen, binds: []keybind{{keyMap.Cancel.hint, i18n.Sprintf("cancel")}}},
	}
}
