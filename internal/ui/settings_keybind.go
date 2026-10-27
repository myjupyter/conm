package ui

import (
	"slices"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui/view"
)

const (
	bindsWindow = 15
	bindLabelW  = 20
)

type bindsTab struct {
	cursor int
	window view.Scroll
	chip   int

	capture   bool
	captureAt int
	captureID int
	first     string
	confirm   bool
}

type bindCaptureExpiredMsg struct{ id int }

func bindGroup(b config.Keybinding) string {
	group, _, _ := strings.Cut(b.Path, ".")
	return group
}

func bindLabel(b config.Keybinding) string {
	_, name, _ := strings.Cut(b.Path, ".")
	return strings.ReplaceAll(name, "_", " ")
}

func defaultKeys(path string) []string {
	for _, b := range config.DefaultKeybindings().Bindings() {
		if b.Path == path {
			return b.Keys
		}
	}
	return nil
}

func (m settingsModel) bindings() []config.Keybinding {
	return slices.DeleteFunc(m.conm.Keybindings.Bindings(), func(b config.Keybinding) bool { return b.Locked })
}

func (m settingsModel) binding() config.Keybinding {
	return m.bindings()[m.binds.cursor]
}

func (m settingsModel) bindChip() int {
	return min(m.binds.chip, max(len(m.binding().Keys)-1, 0))
}

func (m settingsModel) handleBindsKey(key string) (tea.Model, tea.Cmd, bool) {
	switch {
	case m.binds.capture:
		model, cmd := m.captureKey(key)
		return model, cmd, true
	case m.binds.confirm:
		model, cmd := m.confirmReset(key)
		return model, cmd, true
	}

	b := m.binding()
	replaces := keyMap.Settings.ChangeKey.matches(key) || keyMap.Settings.RemoveKey.matches(key)

	switch {
	case replaces && m.isPinned(b, m.bindChip()):
		m.setSettingsStatus(hintOf([]string{b.Keys[m.bindChip()]})+" is a default of "+bindLabel(b)+
			" · "+keyMap.Settings.AddKey.hint+" adds another key", kindWarn)
	case keyMap.Down.matches(key):
		m = m.moveBind(1)
	case keyMap.Up.matches(key):
		m = m.moveBind(-1)
	case keyMap.Right.matches(key):
		m.binds.chip = min(m.bindChip()+1, len(b.Keys)-1)
	case keyMap.Left.matches(key):
		m.binds.chip = max(m.bindChip()-1, 0)
	case keyMap.Settings.AddKey.matches(key):
		m = m.startCapture(len(b.Keys))
	case keyMap.Settings.ChangeKey.matches(key):
		m = m.startCapture(m.bindChip())
	case keyMap.Settings.RemoveKey.matches(key):
		return m.removeKey()
	case keyMap.Settings.Reset.matches(key):
		m = m.askReset()
	default:
		return m, nil, false
	}
	return m, nil, true
}

func (m settingsModel) isPinned(b config.Keybinding, at int) bool {
	if !b.OnlyAppend || at >= len(b.Keys) {
		return false
	}
	stroke := sequence(b.Keys[at])
	return slices.ContainsFunc(defaultKeys(b.Path), func(key string) bool { return sequence(key) == stroke })
}

func (m settingsModel) moveBind(step int) settingsModel {
	n := len(m.bindings())
	m.binds.cursor = (m.binds.cursor + step + n) % n
	m.binds.window.Follow(m.binds.cursor, n)
	m.binds.chip = 0
	m.setSettingsStatus(statusReady, kindIdle)
	return m
}

func (m settingsModel) startCapture(at int) settingsModel {
	m.binds.capture, m.binds.captureAt = true, at
	b := m.binding()
	if at < len(b.Keys) {
		m.setSettingsStatus("press a key to replace "+hintOf([]string{b.Keys[at]})+" · "+keyMap.Cancel.hint+" cancels", kindPending)
	} else {
		m.setSettingsStatus("press a key to add to "+bindLabel(b)+" · "+keyMap.Cancel.hint+" cancels", kindPending)
	}
	return m
}

func (m settingsModel) captureKey(key string) (tea.Model, tea.Cmd) {
	switch {
	case keyMap.Interrupt.matches(key):
		return m, tea.Quit
	case keyMap.Cancel.matches(key):
		m = m.stopCapture()
		m.setSettingsStatus(statusReady, kindIdle)
		return m, nil
	}

	if _, err := config.ParseKey(key); err != nil {
		m = m.stopCapture()
		m.setSettingsStatus(err.Error(), kindWarn)
		return m, nil
	}

	if m.binds.first != "" {
		return m.commitCapture(m.binds.first + " " + keystroke(key))
	}

	m.binds.first = keystroke(key)
	m.binds.captureID++
	id := m.binds.captureID
	m.setSettingsStatus(hintOf([]string{m.binds.first})+" · press a second key for a sequence", kindPending)
	return m, tea.Tick(chordTimeout, func(time.Time) tea.Msg { return bindCaptureExpiredMsg{id: id} })
}

func (m settingsModel) captureExpired(msg bindCaptureExpiredMsg) (tea.Model, tea.Cmd) {
	if !m.binds.capture || msg.id != m.binds.captureID || m.binds.first == "" {
		return m, nil
	}
	return m.commitCapture(m.binds.first)
}

func (m settingsModel) stopCapture() settingsModel {
	m.binds.capture, m.binds.first = false, ""
	return m
}

func (m settingsModel) commitCapture(stroke string) (tea.Model, tea.Cmd) {
	m = m.stopCapture()
	b := m.binding()

	if slices.ContainsFunc(b.Keys, func(key string) bool { return sequence(key) == stroke }) {
		m.setSettingsStatus(hintOf([]string{stroke})+" is already on "+bindLabel(b), kindWarn)
		return m, nil
	}

	keys := slices.Clone(b.Keys)
	status := bindLabel(b) + " + " + hintOf([]string{stroke})
	if at := m.binds.captureAt; at < len(keys) {
		status = bindLabel(b) + " · " + hintOf([]string{keys[at]}) + " → " + hintOf([]string{stroke})
		keys[at] = stroke
	} else {
		keys = append(keys, stroke)
	}
	m.binds.chip = min(m.binds.captureAt, len(keys)-1)

	return m, m.saveKeysCmd(b.Path, keys, status)
}

func (m settingsModel) removeKey() (tea.Model, tea.Cmd, bool) {
	b := m.binding()
	if len(b.Keys) <= 1 {
		m.setSettingsStatus(bindLabel(b)+" needs at least one key · "+keyMap.Settings.ChangeKey.hint+" replaces it", kindWarn)
		return m, nil, true
	}

	at := m.bindChip()
	keys := slices.Delete(slices.Clone(b.Keys), at, at+1)
	m.binds.chip = min(at, len(keys)-1)

	return m, m.saveKeysCmd(b.Path, keys, "removed "+hintOf([]string{b.Keys[at]})+" from "+bindLabel(b)), true
}

func (m settingsModel) askReset() settingsModel {
	b := m.binding()
	if slices.Equal(b.Keys, defaultKeys(b.Path)) {
		m.setSettingsStatus(bindLabel(b)+" is already the default", kindOK)
		return m
	}

	m.binds.confirm = true
	m.setSettingsStatus("reset "+bindLabel(b)+" to "+hintOf(defaultKeys(b.Path))+" ?  "+keyMap.Prompt.Yes.hint+" / "+keyMap.Prompt.No.hint, kindWarn)
	return m
}

func (m settingsModel) confirmReset(key string) (tea.Model, tea.Cmd) {
	m.binds.confirm = false
	b := m.binding()

	if !keyMap.Prompt.Yes.matches(key) {
		m.setSettingsStatus("kept "+bindLabel(b)+" as it is", kindIdle)
		return m, nil
	}

	m.binds.chip = 0
	defaults := defaultKeys(b.Path)
	return m, m.saveKeysCmd(b.Path, defaults, bindLabel(b)+" reset to "+hintOf(defaults))
}
