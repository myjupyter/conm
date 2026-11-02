package ui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/ui/view"
)

func (m secretModel) addSecretCmd() tea.Cmd {
	return secretChangeExec(func() error {
		_, err := runAddSecret(m.secrets.Active(), func(r secretResult) error {
			return m.secrets.Add(context.Background(), r.record, r.password)
		})

		return err
	})
}

func (m secretModel) editSecretCmd() tea.Cmd {
	existing, ok := m.secrets.Secret()
	if !ok {
		return nil
	}
	return secretChangeExec(func() error {
		_, err := runEditSecret(m.secrets.Active(), existing, func(r secretResult) error {
			return m.secrets.Edit(context.Background(), r.record, r.password)
		})

		return err
	})
}

func (m secretModel) removeSecretCmd() tea.Cmd {
	return func() tea.Msg {
		return secretChangedMsg{err: m.secrets.Remove(context.Background())}
	}
}

func secretChangeExec(fn func() error) tea.Cmd {
	return screenCmd(
		fn,
		func(err error) tea.Msg { return secretChangedMsg{err: err} },
	)
}

// runSecretTable opens the secrets screen and blocks until the user leaves it.
func runSecretTable(secrets *view.Secrets) error {
	_, err := runScreen(newSecretModel(secrets))
	return err
}

// runSecretPicker opens the same screen as a chooser and reports the location
// the user attached, if any.
func runSecretPicker(secrets *view.Secrets, current string) (string, bool, error) {
	secrets.StartPicking(current)

	m, err := runScreen(newSecretPicker(secrets))
	picked, ok := secrets.StopPicking()
	if err != nil {
		return "", false, err
	}

	if _, isSecretModel := m.(secretModel); !isSecretModel {
		return "", false, fmt.Errorf("secret picker returned an unexpected model %T", m)
	}
	return picked, ok, nil
}
