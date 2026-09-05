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

type secretResult struct {
	record   config.Secret
	password string
}

func secretForm(
	scheme secret.Scheme,
	isEdit bool,
	save func(secretResult) error,
) (formLoop[secretResult], error) {
	formSpec, ok := secretFormSpecs[scheme]
	if !ok {
		return formLoop[secretResult]{}, fmt.Errorf("form is not implemented for secret store %q", scheme)
	}

	title := formSpec.AddTitle
	if isEdit {
		title = formSpec.EditTitle
	}

	return formLoop[secretResult]{
		open: func(initial formValues, status formStatus) (formResult[secretResult], error) {
			model := newSecretFormModel(formSpec, title, initial, isEdit)
			if status.text != "" {
				model.setStatus(status.text, status.kind)
			}

			return runSecretForm(model)
		},
		save: save,
	}, nil
}

func runAddSecret(scheme secret.Scheme, save func(secretResult) error) (bool, error) {
	loop, err := secretForm(scheme, false, save)
	if err != nil {
		return false, err
	}

	return loop.run(nil, formStatus{})
}

func runEditSecret(
	scheme secret.Scheme,
	existing config.Secret,
	save func(secretResult) error,
) (bool, error) {
	initial, err := seedSecret(scheme, existing)
	if err != nil {
		return false, err
	}

	loop, err := secretForm(scheme, true, save)
	if err != nil {
		return false, err
	}

	return loop.run(initial, formStatus{})
}

func seedSecret(scheme secret.Scheme, existing config.Secret) (formValues, error) {
	formSpec, ok := secretFormSpecs[scheme]
	if !ok {
		return nil, fmt.Errorf("edit form is not implemented for secret store %q", scheme)
	}
	if formSpec.SeedFunc == nil {
		return nil, fmt.Errorf("edit form is not seedable for secret store %q", scheme)
	}

	initial := formSpec.SeedFunc(existing)
	if initial == nil {
		return nil, fmt.Errorf("edit form cannot seed a %T as secret store %q", existing, scheme)
	}

	return initial, nil
}

func runSecretForm(model secretFormModel) (formResult[secretResult], error) {
	var none formResult[secretResult]

	m, err := tea.NewProgram(model).Run()
	if err != nil {
		return none, err
	}

	fm, ok := m.(secretFormModel)
	if !ok {
		return none, fmt.Errorf("secret form returned an unexpected model %T", m)
	}
	if !fm.submitted {
		return none, nil
	}

	sec, password, err := fm.result()
	if err != nil {
		return none, err
	}

	return formResult[secretResult]{
		value:  secretResult{record: sec, password: password},
		values: fm.values(),
		ok:     true,
	}, nil
}
