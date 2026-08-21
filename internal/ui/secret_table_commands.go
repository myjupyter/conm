package ui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/repository"
)

// The forms take over the terminal, so they run the same way the connection
// forms do: a nested program behind tea.Exec, with the result folded back in as
// a message.

func (m secretModel) addSecretCmd() tea.Cmd {
	return secretChangeExec(func() error {
		sec, password, ok, err := runAddSecretForm(m.repo.Kind())
		if err != nil || !ok {
			return err
		}
		return m.repo.Add(context.Background(), sec, password)
	})
}

func (m secretModel) editSecretCmd(i int) tea.Cmd {
	existing, ok := m.repo.Get(i)
	if !ok {
		return nil
	}
	return secretChangeExec(func() error {
		sec, password, ok, err := runEditSecretForm(m.repo.Kind(), existing)
		if err != nil || !ok {
			return err
		}
		return m.repo.Edit(context.Background(), i, sec, password)
	})
}

func (m secretModel) removeSecretCmd(i int) tea.Cmd {
	return func() tea.Msg {
		return secretChangedMsg{err: m.repo.Remove(context.Background(), i)}
	}
}

func secretChangeExec(fn func() error) tea.Cmd {
	return tea.Exec(
		runExec{run: fn},
		func(err error) tea.Msg { return secretChangedMsg{err: err} },
	)
}

// runSecretTable opens the secrets screen and blocks until the user leaves it.
func runSecretTable(repo repository.Secrets) error {
	_, err := tea.NewProgram(newSecretModel(repo)).Run()
	return err
}

// runSecretPicker opens the same screen as a chooser and reports the location
// the user attached, if any.
func runSecretPicker(repo repository.Secrets, current string) (string, bool, error) {
	m, err := tea.NewProgram(newSecretPicker(repo, current)).Run()
	if err != nil {
		return "", false, err
	}

	sm, ok := m.(secretModel)
	if !ok {
		return "", false, fmt.Errorf("secret picker returned an unexpected model %T", m)
	}
	return sm.picked, sm.picked != "", nil
}
