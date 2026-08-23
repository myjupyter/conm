package ui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/ui/view"
)

// The forms take over the terminal, so they run the same way the connection
// forms do: a nested program behind tea.Exec, with the result folded back in as
// a message.

func (m secretModel) addSecretCmd() tea.Cmd {
	return secretChangeExec(func() error {
		sec, password, ok, err := runAddSecretForm(m.secrets.Active())
		if err != nil || !ok {
			return err
		}
		return m.secrets.Add(context.Background(), sec, password)
	})
}

func (m secretModel) editSecretCmd() tea.Cmd {
	existing, ok := m.secrets.Secret()
	if !ok {
		return nil
	}
	return secretChangeExec(func() error {
		sec, password, ok, err := runEditSecretForm(m.secrets.Active(), existing)
		if err != nil || !ok {
			return err
		}
		return m.secrets.Edit(context.Background(), sec, password)
	})
}

func (m secretModel) removeSecretCmd() tea.Cmd {
	return func() tea.Msg {
		return secretChangedMsg{err: m.secrets.Remove(context.Background())}
	}
}

func secretChangeExec(fn func() error) tea.Cmd {
	return tea.Exec(
		runExec{run: fn},
		func(err error) tea.Msg { return secretChangedMsg{err: err} },
	)
}

// runSecretTable opens the secrets screen and blocks until the user leaves it.
func runSecretTable(secrets *view.Secrets) error {
	_, err := tea.NewProgram(newSecretModel(secrets)).Run()
	return err
}

// runSecretPicker opens the same screen as a chooser and reports the location
// the user attached, if any.
func runSecretPicker(secrets *view.Secrets, current string) (string, bool, error) {
	secrets.StartPicking(current)

	m, err := tea.NewProgram(newSecretPicker(secrets)).Run()
	picked, ok := secrets.StopPicking()
	if err != nil {
		return "", false, err
	}

	if _, isSecretModel := m.(secretModel); !isSecretModel {
		return "", false, fmt.Errorf("secret picker returned an unexpected model %T", m)
	}
	return picked, ok, nil
}
