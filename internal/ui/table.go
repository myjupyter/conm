package ui

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/network"
	"github.com/myjupyter/conm/internal/ui/view"
)

type Model struct {
	conns   *view.Connections
	secrets *view.Secrets
	pings   map[view.ConnRef]*ConnState

	status     string
	statusKind statusKind
}

type ConnState struct {
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

func newConnState() *ConnState {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	return &ConnState{
		pingSpinner: s,
		pingStatus:  pingUndefined,
	}
}

func New(conns *view.Connections, secrets *view.Secrets) Model {
	m := Model{
		conns:      conns,
		secrets:    secrets,
		pings:      make(map[view.ConnRef]*ConnState, conns.Len()),
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
	c, ok := m.conns.ConnectionFor(ref)
	if !ok {
		return m, nil
	}
	st, ok := m.pings[ref]
	if !ok {
		return m, nil
	}
	st.pingStatus = pingPinging
	m.status = "pinging " + c.Host() + " …"
	m.statusKind = kindPending
	return m, tea.Batch(st.pingSpinner.Tick, m.pingCmd(ref))
}

func (m Model) pingAt(i int) *ConnState {
	ref, ok := m.conns.RefAt(i)
	if !ok {
		return nil
	}
	return m.pings[ref]
}

func (m Model) currentState() *ConnState {
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
	c, ok := m.conns.Connection()
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s:%d answered in %dms", c.Host(), c.Port(), st.pingResult.PingTime.Milliseconds())
}

func (m Model) retry() (tea.Model, tea.Cmd) {
	e := m.currentErr()
	if e == nil {
		return m, nil
	}
	switch e.action {
	case "ping":
		return m.pingOne()
	case "connect":
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
	c, live := m.conns.ConnectionFor(msg.ref)
	name := connLabel(c)

	if msg.err != nil {
		code := errCode(msg.err)
		st.pingStatus = pingFailed
		if live {
			st.connErr = connErrorFor("ping", "dial tcp", code, c, msg.err)
		}
		m.status, m.statusKind = "ping failed · "+name+" · "+code, kindErr
		return m
	}

	st.pingStatus = pingOK
	st.pingResult = msg.result
	st.connErr = nil
	m.status = fmt.Sprintf("pong · %s responded in %dms", c.Host(), msg.result.PingTime.Milliseconds())
	m.statusKind = kindOK
	return m
}

func (m Model) applyRunResult(msg runResultMsg) Model {
	c, live := m.conns.ConnectionFor(msg.ref)
	name := connLabel(c)
	st := m.pings[msg.ref]

	if msg.err != nil {
		code := errCode(msg.err)
		if live && st != nil {
			st.pingStatus = pingFailed
			st.connErr = connErrorFor("connect", "open session on", code, c, msg.err)
		}
		m.status, m.statusKind = "connect failed · "+name+" · "+code, kindErr
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
		m.pings = make(map[view.ConnRef]*ConnState, m.conns.Len())
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
	c, ok := m.conns.Connection()
	if !ok {
		return ""
	}
	return connLabel(c)
}

func connLabel(c network.Connection) string {
	if c == nil {
		return ""
	}
	if name := c.Name(); name != "" {
		return name
	}
	return c.Host()
}

func connErrorFor(action, op, code string, c network.Connection, err error) *connError {
	return &connError{
		action: action,
		conn:   connLabel(c),
		code:   code,
		target: target(c),
		op:     op + " " + net.JoinHostPort(c.Host(), strconv.Itoa(c.Port())),
		detail: err.Error(),
		hint:   hintFor(code),
	}
}

func target(c network.Connection) string {
	if c == nil {
		return ""
	}
	return fmt.Sprintf("postgres://%s@%s:%d/%s", c.Username(), c.Host(), c.Port(), c.Database())
}

func errCode(err error) string {
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "connection refused"):
		return "ECONNREFUSED"
	case strings.Contains(msg, "timeout"), strings.Contains(msg, "deadline exceeded"):
		return "ETIMEDOUT"
	case strings.Contains(msg, "no such host"), strings.Contains(msg, "no route to host"):
		return "EHOSTUNREACH"
	case strings.Contains(msg, "password authentication"), strings.Contains(msg, "authentication failed"):
		return "EAUTH"
	case strings.Contains(msg, "certificate"), strings.Contains(msg, "x509"), strings.Contains(msg, "tls"):
		return "ETLS"
	case strings.Contains(msg, "invalid connection config"):
		return "EINVALID"
	}
	return "ECONN"
}

func hintFor(code string) string {
	switch code {
	case "ECONNREFUSED":
		return "is the server running and accepting connections on that host and port?"
	case "ETIMEDOUT", "EHOSTUNREACH":
		return "host unreachable from this network · check the address, VPN, or firewall"
	case "EAUTH":
		return "credentials were rejected · press e to update the username or password"
	case "ETLS":
		return "TLS handshake failed · check sslmode and the CA certificate"
	case "EINVALID":
		return "this connection has validation errors · press e to fix the config"
	}
	return ""
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
