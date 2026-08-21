package ui

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/myjupyter/conm/internal/network"
	registry "github.com/myjupyter/conm/internal/repository"
)

type Model struct {
	reg     registry.Connections
	secrets registry.Secrets
	states  []ConnState

	cursor     int
	confirming bool

	status     string
	statusKind statusKind
}

type statusKind int

const (
	kindIdle statusKind = iota
	kindOK
	kindPending
	kindWarn
	kindErr
)

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

type connError struct {
	action string
	conn   string
	code   string
	target string
	op     string
	detail string
	hint   string
}

type pingResultMsg struct {
	index  int
	result network.PingResult
	err    error
}

type runResultMsg struct {
	index int
	err   error
}

type connChangedMsg struct {
	err error
}

func newConnState() ConnState {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	return ConnState{
		pingSpinner: s,
		pingStatus:  pingUndefined,
	}
}

func New(reg registry.Connections, secrets registry.Secrets) Model {
	states := make([]ConnState, reg.Len())
	for i := range states {
		states[i] = newConnState()
	}
	return Model{
		reg:        reg,
		secrets:    secrets,
		states:     states,
		status:     fmt.Sprintf("ready · %d %s", reg.Len(), plural(reg.Len(), "connection", "connections")),
		statusKind: kindIdle,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m.handleKeyMsg(msg.String())

	case pingResultMsg:
		return m.applyPingResult(msg), nil

	case runResultMsg:
		return m.applyRunResult(msg), nil

	case connChangedMsg:
		return m.applyChange(msg.err), nil

	case spinner.TickMsg:
		return m.tickSpinners(msg)
	}

	return m, nil
}

func (m Model) handleKeyMsg(key string) (tea.Model, tea.Cmd) {
	if m.confirming {
		return m.handleConfirmKey(key)
	}
	if m.currentErr() != nil {
		switch key {
		case "esc":
			m.states[m.cursor].connErr = nil
			m.status, m.statusKind = "ready", kindIdle
			return m, nil
		case "r", "R":
			return m.retry()
		}
	}
	return m.handleKey(key)
}

func (m Model) handleKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "ctrl+c", "q", "esc":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < m.reg.Len()-1 {
			m.cursor++
		}
	case "enter":
		return m, m.runCmd(m.cursor)
	case "a":
		return m, m.addCmd()
	case "e":
		if m.reg.Len() > 0 {
			return m, m.editCmd(m.cursor)
		}
	case "d":
		if m.reg.Len() > 0 {
			m.confirming = true
		}
	case "p":
		if m.reg.Len() > 0 {
			return m.pingOne(m.cursor)
		}
	case "s":
		return m, m.secretsCmd()
	}
	return m, nil
}

func (m Model) handleConfirmKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "ctrl+c":
		return m, tea.Quit
	case "y":
		m.confirming = false
		return m, m.removeCmd(m.cursor)
	case "n", "esc":
		m.confirming = false
	}
	return m, nil
}

func (m Model) pingOne(i int) (tea.Model, tea.Cmd) {
	c, ok := m.reg.ConnectionAt(i)
	if !ok {
		return m, nil
	}
	m.states[i].pingStatus = pingPinging
	m.status = "pinging " + c.Host() + " …"
	m.statusKind = kindPending
	return m, tea.Batch(m.states[i].pingSpinner.Tick, m.pingCmd(i))
}

func (m Model) currentErr() *connError {
	if m.cursor < 0 || m.cursor >= len(m.states) {
		return nil
	}
	return m.states[m.cursor].connErr
}

func (m Model) currentPong() string {
	if m.cursor < 0 || m.cursor >= len(m.states) {
		return ""
	}
	st := m.states[m.cursor]
	if st.connErr != nil || st.pingStatus != pingOK {
		return ""
	}
	c, ok := m.reg.ConnectionAt(m.cursor)
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
		return m.pingOne(m.cursor)
	case "connect":
		m.status, m.statusKind = "reconnecting …", kindPending
		return m, m.runCmd(m.cursor)
	}
	return m, nil
}

func (m Model) applyPingResult(msg pingResultMsg) Model {
	if msg.index >= len(m.states) {
		return m
	}
	c, ok := m.reg.ConnectionAt(msg.index)
	name := connLabel(c)

	if msg.err != nil {
		code := errCode(msg.err)
		m.states[msg.index].pingStatus = pingFailed
		if ok {
			m.states[msg.index].connErr = connErrorFor("ping", "dial tcp", code, c, msg.err)
		}
		m.status, m.statusKind = "ping failed · "+name+" · "+code, kindErr
		return m
	}

	m.states[msg.index].pingStatus = pingOK
	m.states[msg.index].pingResult = msg.result
	m.states[msg.index].connErr = nil
	m.status = fmt.Sprintf("pong · %s responded in %dms", c.Host(), msg.result.PingTime.Milliseconds())
	m.statusKind = kindOK
	return m
}

func (m Model) applyRunResult(msg runResultMsg) Model {
	c, ok := m.reg.ConnectionAt(msg.index)
	name := connLabel(c)

	if msg.err != nil {
		code := errCode(msg.err)
		if ok {
			m.states[msg.index].pingStatus = pingFailed
			m.states[msg.index].connErr = connErrorFor("connect", "open session on", code, c, msg.err)
		}
		m.status, m.statusKind = "connect failed · "+name+" · "+code, kindErr
		return m
	}

	if msg.index < len(m.states) {
		m.states[msg.index].connErr = nil
	}
	m.status, m.statusKind = "session closed · "+name, kindIdle
	return m
}

func (m Model) tickSpinners(msg spinner.TickMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	for i := range m.states {
		if m.states[i].pingStatus != pingPinging {
			continue
		}
		var cmd tea.Cmd
		m.states[i].pingSpinner, cmd = m.states[i].pingSpinner.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m Model) applyChange(err error) Model {
	m = m.syncStates()
	if m.cursor > m.reg.Len()-1 {
		m.cursor = max(m.reg.Len()-1, 0)
	}
	if err != nil {
		m.status, m.statusKind = err.Error(), kindErr
	}
	return m
}

func (m Model) syncStates() Model {
	n := m.reg.Len()
	for len(m.states) < n {
		m.states = append(m.states, newConnState())
	}
	if len(m.states) > n {
		m.states = m.states[:n]
	}
	return m
}

func (m Model) cursorLabel() string {
	c, ok := m.reg.ConnectionAt(m.cursor)
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
