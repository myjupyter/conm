package ui

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/network"
	"github.com/myjupyter/conm/internal/secret"
	"github.com/myjupyter/conm/internal/ui/spec"
)

type formModel struct {
	spec     spec.FormSpec
	title    string
	isEdit   bool
	sections []spec.FormSection

	vals map[string]string

	section int
	idx     int
	insert  bool
	reveal  bool

	attempted bool
	submitted bool

	status     string
	statusKind statusKind

	pong string     // successful in-form ping banner
	ping *connError // failed in-form ping panel
}

type formPingMsg struct {
	conn   config.Connection
	result network.PingResult
	err    error
}

func newFormModel(spc spec.FormSpec, title string, initial map[spec.FormFieldKey]spec.FormFieldValue) formModel {
	m := formModel{
		spec:       spc,
		title:      title,
		isEdit:     initial != nil,
		sections:   spc.Sections,
		vals:       make(map[string]string, len(spc.Fields)),
		status:     "ready",
		statusKind: kindIdle,
	}

	if len(m.sections) == 0 {
		keys := make([]spec.FormFieldKey, len(spc.Fields))
		for i, f := range spc.Fields {
			keys[i] = f.Key
		}
		m.sections = []spec.FormSection{{Fields: keys}}
	}

	for _, f := range spc.Fields {
		switch {
		case initial != nil:
			m.vals[f.Key] = initial[f.Key]
		case f.Kind == spec.SelectFieldKind:
			m.vals[f.Key] = f.DefaultValue
			if m.vals[f.Key] == "" && len(f.Options) > 0 {
				m.vals[f.Key] = f.Options[0]
			}
		default:
			m.vals[f.Key] = ""
		}
	}

	return m
}

func (m formModel) Init() tea.Cmd { return nil }

func (m formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.insert {
			return m.insertKey(msg)
		}
		return m.navKey(msg)
	case formPingMsg:
		return m.applyPing(msg), nil
	}
	return m, nil
}

func (m formModel) navKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	fields := m.sectionFields(m.section)
	n := len(fields)
	cur := m.spec.Fields[fields[m.idx]]

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		if m.pong != "" || m.ping != nil {
			m.pong, m.ping = "", nil
			m.setStatus("ready", kindIdle)
			return m, nil
		}
		return m, tea.Quit
	case "down", "j":
		m.idx = (m.idx + 1) % n
	case "up", "k":
		m.idx = (m.idx - 1 + n) % n
	case "tab":
		m.switchSection(1)
	case "shift+tab":
		m.switchSection(-1)
	case "right", "l":
		if cur.Kind == spec.SelectFieldKind {
			m.cycle(fields[m.idx], 1)
		}
	case "left", "h":
		if cur.Kind == spec.SelectFieldKind {
			m.cycle(fields[m.idx], -1)
		}
	case "e":
		if cur.Kind == spec.SelectFieldKind {
			m.setStatus(strings.ToLower(cur.Label)+" is a list · use ←/→", kindWarn)
		} else {
			m.insert = true
			m.setStatus("editing "+strings.ToLower(cur.Label)+" · esc when done", kindIdle)
		}
	case "s":
		if cur.Kind == spec.HiddenFieldKind {
			m.reveal = !m.reveal
			if m.reveal {
				m.setStatus("password visible · s to hide", kindIdle)
			} else {
				m.setStatus("password hidden", kindIdle)
			}
		}
	case "p":
		return m.pingForm()
	case "r":
		if m.ping != nil {
			return m.pingForm()
		}
	case "enter":
		return m.submit()
	}
	return m, nil
}

func (m formModel) insertKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	fields := m.sectionFields(m.section)
	cur := m.spec.Fields[fields[m.idx]]

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.insert = false
		m.setStatus("done editing · hjkl to move", kindIdle)
		return m, nil
	case "enter":
		m.insert = false
		if m.idx < len(fields)-1 {
			m.idx++
		}
		return m, nil
	case "tab":
		m.insert = false
		m.switchSection(1)
		return m, nil
	case "backspace":
		if r := []rune(m.vals[cur.Key]); len(r) > 0 {
			m.vals[cur.Key] = string(r[:len(r)-1])
		}
		return m, nil
	}

	if msg.Text != "" {
		m.vals[cur.Key] += msg.Text
	}
	return m, nil
}

func (m *formModel) setStatus(s string, k statusKind) {
	m.status, m.statusKind = s, k
}

func (m *formModel) switchSection(dir int) {
	m.section = (m.section + dir + len(m.sections)) % len(m.sections)
	m.idx = 0
	s := m.sections[m.section]
	m.setStatus(s.Title+" · "+s.Note, kindIdle)
}

func (m *formModel) cycle(i, dir int) {
	f := m.spec.Fields[i]
	if len(f.Options) == 0 {
		return
	}
	at := 0
	for j, opt := range f.Options {
		if opt == m.vals[f.Key] {
			at = j
			break
		}
	}
	m.vals[f.Key] = f.Options[(at+dir+len(f.Options))%len(f.Options)]
}

// sectionFields returns the indices into spec.Fields for the fields in section si,
// in the section's declared order.
func (m formModel) sectionFields(si int) []int {
	keys := m.sections[si].Fields
	out := make([]int, 0, len(keys))
	for _, key := range keys {
		for i, f := range m.spec.Fields {
			if f.Key == key {
				out = append(out, i)
				break
			}
		}
	}
	return out
}

func (m formModel) currentField() int {
	return m.sectionFields(m.section)[m.idx]
}

func (m formModel) fieldError(i int) string {
	f := m.spec.Fields[i]
	v := m.vals[f.Key]
	if f.Kind != spec.HiddenFieldKind {
		v = strings.TrimSpace(v)
	}

	if f.Property == spec.RequiredFieldProperty && strings.TrimSpace(v) == "" {
		return strings.ToLower(f.Label) + " is required"
	}
	if f.ValidateFunc != nil {
		if err := f.ValidateFunc(v); err != nil {
			return err.Error()
		}
	}
	return ""
}

func (m formModel) sectionErrCount(si int) int {
	n := 0
	for _, i := range m.sectionFields(si) {
		if m.fieldError(i) != "" {
			n++
		}
	}
	return n
}

func (m formModel) submit() (tea.Model, tea.Cmd) {
	m.attempted = true

	bad, badSection, badIdx := 0, -1, -1
	for si := range m.sections {
		for j, i := range m.sectionFields(si) {
			if m.fieldError(i) != "" {
				bad++
				if badSection < 0 {
					badSection, badIdx = si, j
				}
			}
		}
	}

	if bad > 0 {
		m.section, m.idx, m.insert = badSection, badIdx, false
		m.setStatus(fmt.Sprintf("%d %s need attention", bad, plural(bad, "field", "fields")), kindErr)
		return m, nil
	}

	m.submitted = true
	return m, tea.Quit
}

func (m formModel) pingForm() (tea.Model, tea.Cmd) {
	conn, err := m.spec.BuildFunc(m.values())
	if err != nil || conn.Host() == "" {
		m.attempted = true
		m.setStatus("fill host before pinging", kindErr)
		return m, nil
	}

	m.pong, m.ping = "", nil
	m.setStatus("pinging "+conn.Host()+" …", kindPending)
	return m, pingFormCmd(conn)
}

func pingFormCmd(conn config.Connection) tea.Cmd {
	return func() tea.Msg {
		client, err := network.NewConnection(config.Conm{}, conn, secret.Default())
		if err != nil {
			return formPingMsg{conn: conn, err: err}
		}
		defer client.Close()

		result, err := client.Ping(context.Background())
		return formPingMsg{conn: conn, result: result, err: err}
	}
}

func (m formModel) applyPing(msg formPingMsg) formModel {
	c := msg.conn
	label := c.Name()
	if label == "" {
		label = c.Host()
	}

	if msg.err != nil {
		code := errCode(msg.err)
		m.ping = &connError{
			action: "ping",
			conn:   label,
			code:   code,
			target: fmt.Sprintf("postgres://%s@%s:%d/%s", c.Username(), c.Host(), c.Port(), c.Database()),
			op:     "dial tcp " + net.JoinHostPort(c.Host(), strconv.Itoa(c.Port())),
			detail: msg.err.Error(),
			hint:   hintFor(code),
		}
		m.pong = ""
		m.setStatus("ping failed · "+label+" · "+code, kindErr)
		return m
	}

	ms := msg.result.PingTime.Milliseconds()
	m.ping = nil
	m.pong = fmt.Sprintf("%s:%d answered in %dms", c.Host(), c.Port(), ms)
	m.setStatus(fmt.Sprintf("pong · %dms", ms), kindOK)
	return m
}

func (m formModel) result() (config.Connection, error) {
	return m.spec.BuildFunc(m.values())
}

// values collects the current field values keyed by FormField.Key. Every value
// is trimmed except hidden fields (passwords), where surrounding whitespace may
// be meaningful.
func (m formModel) values() map[spec.FormFieldKey]spec.FormFieldValue {
	out := make(map[spec.FormFieldKey]spec.FormFieldValue, len(m.spec.Fields))
	for _, f := range m.spec.Fields {
		v := m.vals[f.Key]
		if f.Kind != spec.HiddenFieldKind {
			v = strings.TrimSpace(v)
		}
		out[f.Key] = v
	}
	return out
}
