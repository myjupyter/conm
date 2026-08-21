package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
	"github.com/myjupyter/conm/internal/ui/spec"
)

// secretFormSpecs maps a secret store to the form that edits its records.
// Keyring is the only store conm supports today; a new one registers its spec
// here and nothing else in the UI changes.
var secretFormSpecs = map[secret.Scheme]spec.FormSpec[config.Secret]{
	secret.Keyring: spec.KeyringFormSpec,
}

func runAddSecretForm(scheme secret.Scheme) (config.Secret, string, bool, error) {
	formSpec, ok := secretFormSpecs[scheme]
	if !ok {
		return nil, "", false, fmt.Errorf("add form is not implemented for secret store %q", scheme)
	}

	return runSecretForm(newSecretFormModel(formSpec, formSpec.AddTitle, nil))
}

func runEditSecretForm(scheme secret.Scheme, existing config.Secret) (config.Secret, string, bool, error) {
	formSpec, ok := secretFormSpecs[scheme]
	if !ok {
		return nil, "", false, fmt.Errorf("edit form is not implemented for secret store %q", scheme)
	}
	if formSpec.SeedFunc == nil {
		return nil, "", false, fmt.Errorf("edit form is not seedable for secret store %q", scheme)
	}

	initial := formSpec.SeedFunc(existing)
	if initial == nil {
		return nil, "", false, fmt.Errorf("edit form cannot seed a %T as secret store %q", existing, scheme)
	}

	return runSecretForm(newSecretFormModel(formSpec, formSpec.EditTitle, initial))
}

func runSecretForm(model secretFormModel) (config.Secret, string, bool, error) {
	m, err := tea.NewProgram(model).Run()
	if err != nil {
		return nil, "", false, err
	}

	fm, ok := m.(secretFormModel)
	if !ok {
		return nil, "", false, fmt.Errorf("secret form returned an unexpected model %T", m)
	}
	if !fm.submitted {
		return nil, "", false, nil
	}

	sec, password, err := fm.result()
	if err != nil {
		return nil, "", false, err
	}

	return sec, password, true, nil
}
