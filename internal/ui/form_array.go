package ui

import (
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/ui/spec"
)

func isArray(f spec.FormField) bool {
	return f.Kind == spec.StringArrayFieldKind
}

func (m formModel) items(f spec.FormField) []string {
	return spec.SplitStringArray(m.vals[f.Key])
}

func (m *formModel) setItems(f spec.FormField, items []string) {
	m.vals[f.Key] = spec.JoinStringArray(items)
}

func (m formModel) chipAtCursor(f spec.FormField) int {
	return min(m.chip, max(len(m.items(f))-1, 0))
}

func (m *formModel) navArray(key string, cur spec.FormField) bool {
	if !isArray(cur) {
		return false
	}

	items := m.items(cur)
	at := m.chipAtCursor(cur)

	switch {
	case keyMap.Form.AddItem.matches(key):
		m.startChip(cur, len(items))
	case keyMap.Form.RemoveItem.matches(key):
		m.removeChip(cur, at)
	case keyMap.Left.matches(key):
		m.chip = max(at-1, 0)
	case keyMap.Right.matches(key):
		m.chip = min(at+1, max(len(items)-1, 0))
	default:
		return false
	}
	return true
}

func (m *formModel) startChip(f spec.FormField, at int) {
	items := m.items(f)
	m.chipAt, m.chipDraft, m.insert = at, "", true
	if at < len(items) {
		m.chipDraft = items[at]
		m.setStatus("editing "+items[at]+" · "+keyMap.Confirm.hint+" keeps it, "+keyMap.Cancel.hint+" discards", kindIdle)
		return
	}
	m.setStatus("adding to "+m.label(f)+" · "+keyMap.Confirm.hint+" keeps it, "+keyMap.Cancel.hint+" discards", kindIdle)
}

func (m *formModel) editChip(f spec.FormField) {
	if len(m.items(f)) == 0 {
		m.startChip(f, 0)
		return
	}
	m.startChip(f, m.chipAtCursor(f))
}

func (m *formModel) commitChip(f spec.FormField) {
	m.insert = false
	items := m.items(f)
	draft := strings.TrimSpace(m.chipDraft)
	m.chipDraft = ""

	switch {
	case draft == "" && m.chipAt < len(items):
		m.removeChip(f, m.chipAt)
		return
	case draft == "":
		m.setStatus("nothing added to "+m.label(f), kindIdle)
		return
	case m.isDuplicate(items, draft):
		m.insert, m.chipDraft = true, draft
		m.setStatus(draft+" is already in "+m.label(f), kindWarn)
		return
	case m.chipAt < len(items):
		items[m.chipAt] = draft
	default:
		items = append(items, draft)
	}

	m.setItems(f, items)
	m.chip = min(m.chipAt, len(m.items(f))-1)
	m.setStatus(m.label(f)+" · "+draft, kindOK)
}

func (m *formModel) discardChip() {
	m.insert, m.chipDraft = false, ""
	m.setStatus("discarded · "+keyMap.MoveVertical.hint+" to move", kindIdle)
}

func (m *formModel) removeChip(f spec.FormField, at int) {
	items := m.items(f)
	if at >= len(items) {
		m.setStatus(m.label(f)+" is empty · "+keyMap.Form.AddItem.hint+" adds one", kindWarn)
		return
	}

	removed := items[at]
	items = slices.Delete(items, at, at+1)
	m.setItems(f, items)
	m.chip = min(at, max(len(items)-1, 0))
	m.setStatus("removed "+removed+" from "+m.label(f), kindWarn)
}

func (m formModel) insertChipKey(msg tea.KeyPressMsg, cur spec.FormField) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch {
	case keyMap.Interrupt.matches(key):
		return m, tea.Quit
	case keyMap.Cancel.matches(key):
		m.discardChip()
	case keyMap.Confirm.matches(key):
		m.commitChip(cur)
	case keyMap.NextSection.matches(key):
		m.commitChip(cur)
		m.switchSection(1)
	case keyMap.Backspace.matches(key):
		if r := []rune(m.chipDraft); len(r) > 0 {
			m.chipDraft = string(r[:len(r)-1])
		}
	default:
		m.chipDraft += msg.Text
	}
	return m, nil
}

func (m formModel) isDuplicate(items []string, draft string) bool {
	for j, item := range items {
		if j != m.chipAt && item == draft {
			return true
		}
	}
	return false
}
