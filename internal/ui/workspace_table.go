package ui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/ui/view"
)

type workspaceModel struct {
	workspaces *view.Workspaces
	entered    bool

	status     string
	statusKind statusKind
}

type workspaceChangedMsg struct {
	err error
}

func newWorkspaceModel(workspaces *view.Workspaces) workspaceModel {
	return workspaceModel{
		workspaces: workspaces,
		status:     statusReady,
		statusKind: kindIdle,
	}
}

func (m workspaceModel) Init() tea.Cmd { return nil }

func (m workspaceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m.handleWorkspaceKeyMsg(msg)

	case workspaceChangedMsg:
		return m.applyWorkspaceChange(msg.err), nil
	}

	return m, nil
}

func (m workspaceModel) handleWorkspaceKeyMsg(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch {
	case m.workspaces.Searching():
		return m.handleWorkspaceSearchKey(msg)
	case m.workspaces.Confirming():
		return m.handleWorkspaceConfirmKey(key)
	}
	return m.handleWorkspaceKey(key)
}

func (m workspaceModel) handleWorkspaceSearchKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch {
	case keyMap.Interrupt.matches(key):
		return m, tea.Quit
	case keyMap.Cancel.matches(key):
		m.workspaces.CancelSearch()
	case keyMap.Search.matches(key) && m.workspaces.Query() == "":
		m.workspaces.CancelSearch()
	case keyMap.Confirm.matches(key):
		m.workspaces.CommitSearch()
	case keyMap.Backspace.matches(key):
		m.workspaces.TrimSearch()
	default:
		m.workspaces.AppendSearch(msg.Text)
	}
	return m, nil
}

func (m workspaceModel) handleWorkspaceKey(key string) (tea.Model, tea.Cmd) {
	switch {
	case keyMap.Search.matches(key):
		m.workspaces.StartSearch()
	case m.workspaces.Filtered() && keyMap.Cancel.matches(key):
		m.workspaces.ClearSearch()
	case keyMap.Quit.matches(key):
		return m, tea.Quit
	case keyMap.Up.matches(key):
		m.workspaces.MoveUp()
	case keyMap.Down.matches(key):
		m.workspaces.MoveDown()
	case keyMap.Confirm.matches(key):
		return m.enter()
	case keyMap.Workspace.Add.matches(key):
		return m, m.addWorkspaceCmd()
	case keyMap.Workspace.Edit.matches(key):
		if m.workspaces.Len() > 0 {
			return m, m.renameWorkspaceCmd()
		}
	case keyMap.Workspace.Delete.matches(key):
		if m.workspaces.Len() > 0 {
			m.workspaces.AskConfirm()
		}
	case keyMap.Help.matches(key):
		m.workspaces.ToggleHelp()
	}
	return m, nil
}

func (m workspaceModel) handleWorkspaceConfirmKey(key string) (tea.Model, tea.Cmd) {
	switch {
	case keyMap.Interrupt.matches(key):
		return m, tea.Quit
	case keyMap.Prompt.Yes.matches(key):
		m.workspaces.ClearConfirm()
		return m, m.removeWorkspaceCmd()
	case keyMap.Prompt.No.matches(key):
		m.workspaces.ClearConfirm()
	}
	return m, nil
}

func (m workspaceModel) enter() (tea.Model, tea.Cmd) {
	if err := m.workspaces.Use(); err != nil {
		m.setWorkspaceStatus(err.Error(), kindErr)
		return m, nil
	}
	m.entered = true
	return m, tea.Quit
}

func (m workspaceModel) applyWorkspaceChange(err error) workspaceModel {
	m.workspaces.Sync()
	if err != nil {
		m.setWorkspaceStatus(err.Error(), kindErr)
		return m
	}

	n := m.workspaces.Len()
	m.setWorkspaceStatus(fmt.Sprintf("%d %s", n, plural(n, "workspace", "workspaces")), kindOK)
	return m
}

func (m *workspaceModel) setWorkspaceStatus(s string, k statusKind) {
	m.status, m.statusKind = s, k
}

func (m workspaceModel) cursorWorkspaceLabel() string {
	name, _ := m.workspaces.Workspace()
	return name
}
