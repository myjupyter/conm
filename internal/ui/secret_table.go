package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/ui/view"
)

// secretModel lists the records of one secret store. It owns no material: the
// repository is the only thing that talks to the provider, and a password only
// ever passes through on its way from the form into repository.Secrets.
type secretModel struct {
	secrets *view.Secrets

	status     string
	statusKind statusKind
}

type secretChangedMsg struct {
	err error
}

func newSecretModel(secrets *view.Secrets) secretModel {
	n := secrets.Len()
	return secretModel{
		secrets:    secrets,
		status:     fmt.Sprintf("%s · %d %s", secrets.Active(), n, plural(n, "entry", "entries")),
		statusKind: kindIdle,
	}
}

// newSecretPicker opens the same screen as a chooser, positioned on the entry
// the caller already holds.
func newSecretPicker(secrets *view.Secrets) secretModel {
	m := newSecretModel(secrets)
	if m.secrets.Len() == 0 {
		m.setSecretStatus(m.secrets.Active()+" is empty · press "+keyMap.Add.hint+" to add an entry", kindWarn)
	} else {
		m.setSecretStatus("pick an entry · "+keyMap.Confirm.hint+" to attach it", kindIdle)
	}
	return m
}

func (m secretModel) Init() tea.Cmd { return nil }

func (m secretModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m.handleSecretKeyMsg(msg.String())

	case secretChangedMsg:
		return m.applySecretChange(msg.err), nil
	}

	return m, nil
}

func (m secretModel) handleSecretKeyMsg(key string) (tea.Model, tea.Cmd) {
	if m.secrets.Confirming() {
		return m.handleSecretConfirmKey(key)
	}
	return m.handleSecretKey(key)
}

func (m secretModel) handleSecretKey(key string) (tea.Model, tea.Cmd) {
	switch {
	case keyMap.Quit.matches(key):
		return m, tea.Quit
	case keyMap.Up.matches(key):
		m.secrets.MoveUp()
	case keyMap.Down.matches(key):
		m.secrets.MoveDown()
	case keyMap.SwitchStore.matches(key):
		// One store is configured, so switching is a no-op worth saying out loud.
		m.setSecretStatus(m.secrets.Active()+" is the only store configured", kindIdle)
	case keyMap.Confirm.matches(key):
		if m.secrets.Picking() {
			if _, ok := m.secrets.Pick(); ok {
				return m, tea.Quit
			}
			return m, nil
		}
		return m.describe(), nil
	case keyMap.Add.matches(key):
		return m, m.addSecretCmd()
	case keyMap.Edit.matches(key):
		if m.secrets.Len() > 0 {
			return m, m.editSecretCmd()
		}
	case keyMap.Delete.matches(key):
		m.secrets.AskConfirm()
	case keyMap.Help.matches(key):
		m.setSecretStatus(keyhintStatus(m.secrets.ToggleHelp()), kindIdle)
	}
	return m, nil
}

func (m secretModel) handleSecretConfirmKey(key string) (tea.Model, tea.Cmd) {
	switch {
	case keyMap.Interrupt.matches(key):
		return m, tea.Quit
	case keyMap.Yes.matches(key):
		m.secrets.ClearConfirm()
		return m, m.removeSecretCmd()
	case keyMap.No.matches(key):
		m.secrets.ClearConfirm()
	}
	return m, nil
}

// describe reports where the selected entry points and who depends on it — the
// two things the table has no room to spell out in full.
func (m secretModel) describe() secretModel {
	s, ok := m.secrets.Secret()
	if !ok {
		return m
	}

	used := m.secrets.Usages()
	if len(used) == 0 {
		m.setSecretStatus(fmt.Sprintf("%s:%s · unused", s.Provider(), s.Location()), kindIdle)
		return m
	}

	m.setSecretStatus(fmt.Sprintf("%s:%s · used by %d %s",
		s.Provider(), s.Location(), len(used), plural(len(used), "connection", "connections")), kindIdle)
	return m
}

func (m secretModel) applySecretChange(err error) secretModel {
	m.secrets.Sync()
	if err != nil {
		m.setSecretStatus(err.Error(), kindErr)
		return m
	}

	n := m.secrets.Len()
	m.setSecretStatus(fmt.Sprintf("%s · %d %s", m.secrets.Active(), n, plural(n, "entry", "entries")), kindOK)
	return m
}

func (m *secretModel) setSecretStatus(s string, k statusKind) {
	m.status, m.statusKind = s, k
}

func (m secretModel) cursorSecretLabel() string {
	s, ok := m.secrets.Secret()
	if !ok {
		return ""
	}
	return s.ID()
}
