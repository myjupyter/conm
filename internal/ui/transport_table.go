package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/repository"
	"github.com/myjupyter/conm/internal/ui/view"
)

type transportModel struct {
	transports *view.Transports
	conns      *view.Connections

	status     string
	statusKind statusKind
}

type transportChangedMsg struct {
	err error
}

func newTransportModel(transports *view.Transports, conns *view.Connections) transportModel {
	return transportModel{
		transports: transports,
		conns:      conns,
		status:     statusReady,
		statusKind: kindIdle,
	}
}

func (m transportModel) Init() tea.Cmd { return nil }

func (m transportModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m.handleTransportKeyMsg(msg)

	case transportChangedMsg:
		return m.applyTransportChange(msg.err), nil
	}

	return m, nil
}

func (m transportModel) handleTransportKeyMsg(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if m.transports.Searching() {
		return m.handleTransportSearchKey(msg)
	}
	if m.transports.Confirming() {
		return m.handleTransportConfirmKey(key)
	}
	return m.handleTransportKey(key)
}

func (m transportModel) handleTransportSearchKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch {
	case keyMap.Interrupt.matches(key):
		return m, tea.Quit
	case keyMap.Cancel.matches(key):
		m.transports.CancelSearch()
	case keyMap.Search.matches(key) && m.transports.Query() == "":
		m.transports.CancelSearch()
	case keyMap.Confirm.matches(key):
		m.transports.CommitSearch()
	case keyMap.Backspace.matches(key):
		m.transports.TrimSearch()
	default:
		m.transports.AppendSearch(msg.Text)
	}
	return m, nil
}

func (m transportModel) handleTransportKey(key string) (tea.Model, tea.Cmd) {
	switch {
	case keyMap.Search.matches(key):
		m.transports.StartSearch()
	case m.transports.Filtered() && keyMap.Cancel.matches(key):
		m.transports.ClearSearch()
	case keyMap.Quit.matches(key):
		return m, tea.Quit
	case keyMap.Up.matches(key):
		m.transports.MoveUp()
	case keyMap.Down.matches(key):
		m.transports.MoveDown()
	case keyMap.SwitchKind.matches(key):
		m.setTransportStatus(m.transports.Active()+" is the only transport kind", kindWarn)
	case keyMap.Confirm.matches(key):
		return m.describe(), nil
	case keyMap.Transport.Add.matches(key):
		return m, m.addTransportCmd()
	case keyMap.Transport.Edit.matches(key):
		if m.transports.Len() > 0 {
			return m, m.editTransportCmd()
		}
	case keyMap.Transport.Delete.matches(key):
		m.transports.AskConfirm()
	case keyMap.Help.matches(key):
		m.transports.ToggleHelp()
	}
	return m, nil
}

func (m transportModel) handleTransportConfirmKey(key string) (tea.Model, tea.Cmd) {
	switch {
	case keyMap.Interrupt.matches(key):
		return m, tea.Quit
	case keyMap.Prompt.Yes.matches(key):
		m.transports.ClearConfirm()
		return m, m.removeTransportCmd()
	case keyMap.Prompt.No.matches(key):
		m.transports.ClearConfirm()
	}
	return m, nil
}

func (m transportModel) describe() transportModel {
	t, ok := m.transports.Transport()
	if !ok {
		return m
	}

	used := m.transports.Usages()
	if len(used) == 0 {
		m.setTransportStatus(fmt.Sprintf("%s via %s · unused", t.Name, m.viaText(t)), kindOK)
		return m
	}

	m.setTransportStatus(fmt.Sprintf("%s via %s · used by %s", t.Name, m.viaText(t), usageNames(used)), kindOK)
	return m
}

func (m transportModel) viaText(t config.Transport) string {
	via, ok := m.conns.Named(config.SSHConnType, t.Via)
	if !ok {
		return t.Via + " (missing)"
	}
	return fmt.Sprintf("%s (%s:%d)", t.Via, via.Host(), via.Port())
}

func (m transportModel) applyTransportChange(err error) transportModel {
	m.transports.Sync()
	if err != nil {
		m.setTransportStatus(err.Error(), kindErr)
		return m
	}

	n := m.transports.Len()
	m.setTransportStatus(fmt.Sprintf("%d %s", n, plural(n, "transport", "transports")), kindOK)
	return m
}

func (m *transportModel) setTransportStatus(s string, k statusKind) {
	m.status, m.statusKind = s, k
}

func (m transportModel) cursorTransportLabel() string {
	t, ok := m.transports.Transport()
	if !ok {
		return ""
	}
	return t.Name
}

func usageNames(used []repository.Usage) string {
	names := make([]string, 0, len(used))
	for _, u := range used {
		names = append(names, u.String())
	}
	return strings.Join(names, ", ")
}
