package ui

import (
	"fmt"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/network"
	"github.com/myjupyter/conm/internal/ui/view"
)

type Model struct {
	conns   *view.Connections
	secrets *view.Secrets
	pings   map[view.ConnRef]*connState

	status     string
	statusKind statusKind
}

type connState struct {
	pingSpinner spinner.Model

	pingStatus pingStatus
	pingResult network.PingResult

	connErr *connError
}

type pingStatus int

const (
	pingUndefined pingStatus = iota
	pingPinging
	pingOK
	pingFailed
)

type pingResultMsg struct {
	ref    view.ConnRef
	result network.PingResult
	err    error
}

type runResultMsg struct {
	ref view.ConnRef
	err error
}

type connChangedMsg struct {
	err        error
	renumbered bool
}

func newConnState() *connState {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	return &connState{
		pingSpinner: s,
		pingStatus:  pingUndefined,
	}
}

func New(conns *view.Connections, secrets *view.Secrets) Model {
	m := Model{
		conns:      conns,
		secrets:    secrets,
		pings:      make(map[view.ConnRef]*connState, conns.Len()),
		status:     statusReady,
		statusKind: kindIdle,
	}

	return m.syncPings()
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m.handleKeyMsg(msg)

	case pingResultMsg:
		return m.applyPingResult(msg), nil

	case runResultMsg:
		return m.applyRunResult(msg), nil

	case connChangedMsg:
		return m.applyChange(msg), nil

	case spinner.TickMsg:
		return m.tickSpinners(msg)
	}

	return m, nil
}

func (m Model) handleKeyMsg(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if m.conns.Searching() {
		return m.handleSearchKey(msg)
	}
	if m.conns.Confirming() {
		return m.handleConfirmKey(key)
	}
	if m.currentErr() != nil {
		switch {
		case keyMap.Cancel.matches(key):
			m.currentState().connErr = nil
			m.status, m.statusKind = statusReady, kindIdle
			return m, nil
		case keyMap.Retry.matches(key):
			return m.retry()
		}
	}
	return m.handleKey(key)
}

func (m Model) handleSearchKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch {
	case keyMap.Interrupt.matches(key):
		return m, tea.Quit
	case keyMap.Cancel.matches(key):
		m.conns.CancelSearch()
	case keyMap.Search.matches(key) && m.conns.Query() == "":
		m.conns.CancelSearch()
	case keyMap.Confirm.matches(key):
		m.conns.CommitSearch()
	case keyMap.Backspace.matches(key):
		m.conns.TrimSearch()
	default:
		m.conns.AppendSearch(msg.Text)
	}
	return m.syncPings(), nil
}

func (m Model) handleKey(key string) (tea.Model, tea.Cmd) {
	switch {
	case keyMap.Search.matches(key):
		m.conns.StartSearch()
	case m.conns.Filtered() && keyMap.Cancel.matches(key):
		m.conns.ClearSearch()
		return m.syncPings(), nil
	case keyMap.Quit.matches(key):
		return m, tea.Quit
	case keyMap.Up.matches(key):
		m.conns.MoveUp()
	case keyMap.Down.matches(key):
		m.conns.MoveDown()
	case keyMap.Confirm.matches(key):
		return m, m.runCmd()
	case keyMap.Add.matches(key):
		return m, m.addCmd()
	case keyMap.Edit.matches(key):
		if m.conns.Len() > 0 {
			return m, m.editCmd()
		}
	case keyMap.Delete.matches(key):
		m.conns.AskConfirm()
	case keyMap.Ping.matches(key):
		if m.conns.Len() > 0 {
			return m.pingOne()
		}
	case keyMap.Secret.matches(key):
		return m, m.secretsCmd()
	case keyMap.Help.matches(key):
		m.conns.ToggleHelp()
	}
	return m, nil
}

func (m Model) handleConfirmKey(key string) (tea.Model, tea.Cmd) {
	switch {
	case keyMap.Interrupt.matches(key):
		return m, tea.Quit
	case keyMap.Yes.matches(key):
		m.conns.ClearConfirm()
		return m, m.removeCmd()
	case keyMap.No.matches(key):
		m.conns.ClearConfirm()
	}
	return m, nil
}

func (m Model) pingOne() (tea.Model, tea.Cmd) {
	ref, ok := m.conns.Ref()
	if !ok {
		return m, nil
	}
	cfg, ok := m.conns.Config()
	if !ok {
		return m, nil
	}
	st, ok := m.pings[ref]
	if !ok {
		return m, nil
	}
	st.pingStatus = pingPinging
	m.status = "pinging " + cfg.Host() + " …"
	m.statusKind = kindPending
	return m, tea.Batch(st.pingSpinner.Tick, m.pingCmd(ref))
}

func (m Model) pingAt(i int) *connState {
	ref, ok := m.conns.RefAt(i)
	if !ok {
		return nil
	}
	return m.pings[ref]
}

func (m Model) currentState() *connState {
	return m.pingAt(m.conns.Cursor())
}

func (m Model) currentErr() *connError {
	st := m.currentState()
	if st == nil {
		return nil
	}
	return st.connErr
}

func (m Model) currentPong() string {
	st := m.currentState()
	if st == nil || st.connErr != nil || st.pingStatus != pingOK {
		return ""
	}
	cfg, ok := m.conns.Config()
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s:%d answered in %dms", cfg.Host(), cfg.Port(), st.pingResult.PingTime.Milliseconds())
}

func (m Model) retry() (tea.Model, tea.Cmd) {
	e := m.currentErr()
	if e == nil {
		return m, nil
	}
	switch e.op {
	case network.PingOperation:
		return m.pingOne()
	case network.ConnectOperation:
		m.status, m.statusKind = "reconnecting …", kindPending
		return m, m.runCmd()
	}
	return m, nil
}

func (m Model) applyPingResult(msg pingResultMsg) Model {
	st, ok := m.pings[msg.ref]
	if !ok {
		return m
	}
	cfg, live := m.conns.ConfigFor(msg.ref)
	name := connLabel(cfg)

	if msg.err != nil {
		e := newConnError(msg.err, network.PingOperation, name)
		st.pingStatus = pingFailed
		if live {
			st.connErr = e
		}
		m.status, m.statusKind = "ping failed · "+name+" · "+e.code, kindErr
		return m
	}

	st.pingStatus = pingOK
	st.pingResult = msg.result
	st.connErr = nil
	m.status = fmt.Sprintf("pong · %s responded in %dms", cfg.Host(), msg.result.PingTime.Milliseconds())
	m.statusKind = kindOK
	return m
}

func (m Model) applyRunResult(msg runResultMsg) Model {
	cfg, live := m.conns.ConfigFor(msg.ref)
	name := connLabel(cfg)
	st := m.pings[msg.ref]

	if msg.err != nil {
		e := newConnError(msg.err, network.ConnectOperation, name)
		if live && st != nil {
			st.pingStatus = pingFailed
			st.connErr = e
		}
		m.status, m.statusKind = "connect failed · "+name+" · "+e.code, kindErr
		return m
	}

	if st != nil {
		st.connErr = nil
	}
	m.status, m.statusKind = "session closed · "+name, kindOK
	return m
}

func (m Model) tickSpinners(msg spinner.TickMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	for _, st := range m.pings {
		if st.pingStatus != pingPinging {
			continue
		}
		var cmd tea.Cmd
		st.pingSpinner, cmd = st.pingSpinner.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m Model) applyChange(msg connChangedMsg) Model {
	if msg.renumbered {
		m.pings = make(map[view.ConnRef]*connState, m.conns.Len())
	}
	m = m.syncPings()
	if msg.err != nil {
		m.status, m.statusKind = msg.err.Error(), kindErr
	}
	return m
}

func (m Model) syncPings() Model {
	m.conns.Sync()
	for i := range m.conns.Len() {
		ref, ok := m.conns.RefAt(i)
		if !ok {
			continue
		}
		if _, ok := m.pings[ref]; !ok {
			m.pings[ref] = newConnState()
		}
	}
	return m
}

func (m Model) cursorLabel() string {
	cfg, ok := m.conns.Config()
	if !ok {
		return ""
	}
	return connLabel(cfg)
}

func connLabel(cfg config.Connection) string {
	if cfg == nil {
		return ""
	}
	if name := cfg.Name(); name != "" {
		return name
	}
	return cfg.Host()
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
