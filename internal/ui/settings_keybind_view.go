package ui

import (
	"image/color"
	"slices"

	"github.com/myjupyter/conm/internal/config"
)

func (m settingsModel) bindsLines() []string {
	bindings := m.bindings()
	start, end := m.binds.window.Range(len(bindings))

	var lines []string
	for i := start; i < end; i++ {
		if i == start || bindGroup(bindings[i]) != bindGroup(bindings[i-1]) {
			lines = append(lines, frameHead(infoInner, bindGroup(bindings[i]), "")...)
		}
		lines = append(lines, m.bindLine(bindings[i], i == m.binds.cursor))
	}
	return lines
}

func (m settingsModel) bindLine(b config.Keybinding, on bool) string {
	var bg color.Color
	noteFg := cMuted
	if on {
		bg, noteFg = cAccent, cInvFg
	}

	value := frameChips(m.bindChips(b, on), on, bg)
	value = append(value, span{text: "  " + m.bindNote(b, on), fg: noteFg, bg: bg})
	return settingsLabelLine(bindLabel(b), bindLabelW, value, on)
}

func (m settingsModel) bindChips(b config.Keybinding, on bool) []chip {
	bound := slices.Clone(b.Keys)
	capturing := on && m.binds.capture
	if capturing && m.binds.captureAt >= len(bound) {
		bound = append(bound, "")
	}

	focus := -1
	switch {
	case capturing:
		focus = m.binds.captureAt
	case on:
		focus = m.bindChip()
	}

	chips := make([]chip, len(bound))
	for j, key := range bound {
		text := hintOf([]string{key})
		switch {
		case capturing && j == focus && m.binds.first != "":
			text = hintOf([]string{m.binds.first}) + " " + gEllipsis
		case capturing && j == focus:
			text = "press a key" + gEllipsis
		}
		chips[j] = chip{text: text, focused: j == focus}
	}
	return chips
}

func (m settingsModel) bindNote(b config.Keybinding, on bool) string {
	switch {
	case on && m.binds.capture:
		return keyMap.Cancel.hint + " cancels"
	case on && m.binds.confirm:
		return "reset to default?  " + keyMap.Prompt.Yes.hint + " / " + keyMap.Prompt.No.hint
	}
	if defaults := defaultKeys(b.Path); !slices.Equal(b.Keys, defaults) {
		return "default " + hintOf(defaults)
	}
	return ""
}
