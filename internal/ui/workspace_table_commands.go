package ui

import (
	"context"
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/repository"
	"github.com/myjupyter/conm/internal/secret"
	"github.com/myjupyter/conm/internal/ui/view"
)

func workspaceSecrets() (*secret.Resolver, error) {
	workspace, err := config.SecretWorkspace()
	if err != nil {
		return nil, err
	}

	return secret.DefaultFor(workspace), nil
}

func (m workspaceModel) addWorkspaceCmd() tea.Cmd {
	return workspaceChangeExec(func() error {
		_, err := runAddWorkspace(m.workspaces.Add)
		return err
	})
}

func (m workspaceModel) renameWorkspaceCmd() tea.Cmd {
	existing, ok := m.workspaces.Workspace()
	if !ok {
		return nil
	}
	return workspaceChangeExec(func() error {
		_, err := runRenameWorkspace(existing, m.workspaces.Rename)
		return err
	})
}

func workspaceChangeExec(fn func() error) tea.Cmd {
	return tea.Exec(
		runExec{run: fn},
		func(err error) tea.Msg { return workspaceChangedMsg{err: err} },
	)
}

func (m workspaceModel) removeWorkspaceCmd() tea.Cmd {
	workspaces := m.workspaces
	return func() tea.Msg {
		return workspaceChangedMsg{err: workspaces.Remove(context.Background())}
	}
}

func RunWorkspaces() error {
	chosen, entered, err := runWorkspaceTable()
	if err != nil || !entered {
		return err
	}

	if err := config.UseWorkspace(chosen); err != nil {
		return err
	}
	conm, err := config.ReadConm(config.ConmPath())
	if err != nil {
		return err
	}

	return Run(conm)
}

func runWorkspaceTable() (string, bool, error) {
	repo, err := repository.NewWorkspaceRepository()
	if err != nil {
		return "", false, err
	}

	m, err := tea.NewProgram(newWorkspaceModel(view.NewWorkspaces(repo))).Run()
	if err != nil {
		return "", false, err
	}

	wm, ok := m.(workspaceModel)
	if !ok {
		return "", false, fmt.Errorf("workspaces screen returned an unexpected model %T", m)
	}

	return wm.workspaces.Current(), wm.entered, nil
}
