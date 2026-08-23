package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/repository"
)

// secretModel lists the records of one secret store. It owns no material: the
// repository is the only thing that talks to the provider, and a password only
// ever passes through on its way from the form into repository.Secrets.
type secretModel struct {
	repo repository.Secrets

	// picking turns the screen into a chooser: enter returns the highlighted
	// location to the caller instead of describing it.
	picking bool
	picked  string

	cursor     int
	confirming bool
	help       bool // the key hints table, expanded over the key bar

	status     string
	statusKind statusKind
}

type secretChangedMsg struct {
	err error
}

func newSecretModel(repo repository.Secrets) secretModel {
	n := repo.Len()
	return secretModel{
		repo:       repo,
		status:     fmt.Sprintf("%s · %d %s", repo.Kind(), n, plural(n, "entry", "entries")),
		statusKind: kindIdle,
	}
}

// newSecretPicker opens the same screen as a chooser, positioned on the entry
// the caller already holds.
func newSecretPicker(repo repository.Secrets, current string) secretModel {
	m := newSecretModel(repo)
	m.picking = true
	for i := range repo.Len() {
		if s, ok := repo.Get(i); ok && s.Location() == current {
			m.cursor = i
			break
		}
	}
	if repo.Len() == 0 {
		m.setSecretStatus(repo.Kind()+" is empty · press "+keyMap.Add.hint+" to add an entry", kindWarn)
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
	if m.confirming {
		return m.handleSecretConfirmKey(key)
	}
	return m.handleSecretKey(key)
}

func (m secretModel) handleSecretKey(key string) (tea.Model, tea.Cmd) {
	switch {
	case keyMap.Quit.matches(key):
		return m, tea.Quit
	case keyMap.Up.matches(key):
		if m.cursor > 0 {
			m.cursor--
		}
	case keyMap.Down.matches(key):
		if m.cursor < m.repo.Len()-1 {
			m.cursor++
		}
	case keyMap.SwitchStore.matches(key):
		// One store is configured, so switching is a no-op worth saying out loud.
		m.setSecretStatus(m.repo.Kind()+" is the only store configured", kindIdle)
	case keyMap.Confirm.matches(key):
		if m.picking {
			if s, ok := m.repo.Get(m.cursor); ok {
				m.picked = s.Location()
				return m, tea.Quit
			}
			return m, nil
		}
		return m.describe(), nil
	case keyMap.Add.matches(key):
		return m, m.addSecretCmd()
	case keyMap.Edit.matches(key):
		if m.repo.Len() > 0 {
			return m, m.editSecretCmd(m.cursor)
		}
	case keyMap.Delete.matches(key):
		if m.repo.Len() > 0 {
			m.confirming = true
		}
	case keyMap.Help.matches(key):
		m.help = !m.help
		m.setSecretStatus(keyhintStatus(m.help), kindIdle)
	}
	return m, nil
}

func (m secretModel) handleSecretConfirmKey(key string) (tea.Model, tea.Cmd) {
	switch {
	case keyMap.Interrupt.matches(key):
		return m, tea.Quit
	case keyMap.Yes.matches(key):
		m.confirming = false
		return m, m.removeSecretCmd(m.cursor)
	case keyMap.No.matches(key):
		m.confirming = false
	}
	return m, nil
}

// describe reports where the selected entry points and who depends on it — the
// two things the table has no room to spell out in full.
func (m secretModel) describe() secretModel {
	s, ok := m.repo.Get(m.cursor)
	if !ok {
		return m
	}

	used := m.repo.UsagesAt(m.cursor)
	if len(used) == 0 {
		m.setSecretStatus(fmt.Sprintf("%s:%s · unused", s.Provider(), s.Location()), kindIdle)
		return m
	}

	m.setSecretStatus(fmt.Sprintf("%s:%s · used by %d %s",
		s.Provider(), s.Location(), len(used), plural(len(used), "connection", "connections")), kindIdle)
	return m
}

func (m secretModel) applySecretChange(err error) secretModel {
	if m.cursor > m.repo.Len()-1 {
		m.cursor = max(m.repo.Len()-1, 0)
	}
	if err != nil {
		m.setSecretStatus(err.Error(), kindErr)
		return m
	}

	n := m.repo.Len()
	m.setSecretStatus(fmt.Sprintf("%s · %d %s", m.repo.Kind(), n, plural(n, "entry", "entries")), kindOK)
	return m
}

func (m *secretModel) setSecretStatus(s string, k statusKind) {
	m.status, m.statusKind = s, k
}

func (m secretModel) cursorSecretLabel() string {
	s, ok := m.repo.Get(m.cursor)
	if !ok {
		return ""
	}
	return s.ID()
}
