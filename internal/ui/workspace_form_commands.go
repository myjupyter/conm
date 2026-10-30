package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/repository"
	"github.com/myjupyter/conm/internal/ui/spec"
)

func RunAddWorkspaceForm() (bool, error) {
	repo, err := repository.NewWorkspaceRepository()
	if err != nil {
		return false, err
	}

	return runAddWorkspace(repo.Add)
}

func workspaceForm(isEdit bool, save func(string) error) formLoop[string] {
	title := spec.WorkspaceFormSpec.AddTitle
	if isEdit {
		title = spec.WorkspaceFormSpec.EditTitle
	}

	return formLoop[string]{
		open: func(initial formValues, status formStatus) (formResult[string], error) {
			model := newWorkspaceFormModel(spec.WorkspaceFormSpec, title, initial, isEdit)
			if status.text != "" {
				model.setStatus(status.text, status.kind)
			}

			return runWorkspaceForm(model)
		},
		save: save,
	}
}

func runAddWorkspace(save func(string) error) (bool, error) {
	return workspaceForm(false, save).run(nil, formStatus{})
}

func runRenameWorkspace(existing string, save func(string) error) (bool, error) {
	return workspaceForm(true, save).run(spec.WorkspaceFormSpec.SeedFunc(existing), formStatus{})
}

func runWorkspaceForm(model workspaceFormModel) (formResult[string], error) {
	var none formResult[string]

	m, err := tea.NewProgram(model).Run()
	if err != nil {
		return none, err
	}

	fm, ok := m.(workspaceFormModel)
	if !ok {
		return none, fmt.Errorf("workspace form returned an unexpected model %T", m)
	}
	if !fm.submitted {
		return none, nil
	}

	values := fm.values()
	name, err := fm.spec.BuildFunc(values)
	if err != nil {
		return none, err
	}

	return formResult[string]{value: name, values: values, ok: true}, nil
}
