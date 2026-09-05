package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui/spec"
)

type secretFormModel struct {
	spec   spec.FormSpec[config.Secret]
	title  string
	isEdit bool

	vals map[string]string

	idx    int
	insert bool
	reveal bool
	help   bool

	attempted bool
	submitted bool

	status     string
	statusKind statusKind
}

func newSecretFormModel(spc spec.FormSpec[config.Secret], title string, initial formValues, isEdit bool) secretFormModel {
	m := secretFormModel{
		spec:       spc,
		title:      title,
		isEdit:     isEdit,
		vals:       make(map[string]string, len(spc.Fields)),
		status:     statusReady,
		statusKind: kindIdle,
	}

	for _, f := range spc.Fields {
		seeded, ok := initial[f.Key]
		switch {
		case ok:
			m.vals[f.Key] = seeded
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

func (m secretFormModel) Init() tea.Cmd { return nil }

func (m secretFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		if m.insert {
			return m.insertKey(msg)
		}
		return m.navKey(msg)
	}
	return m, nil
}

func (m secretFormModel) navKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	n := len(m.spec.Fields)
	cur := m.spec.Fields[m.idx]

	key := msg.String()

	switch {
	case keyMap.Interrupt.matches(key), keyMap.Cancel.matches(key):
		return m, tea.Quit
	case keyMap.Down.matches(key):
		m.idx = (m.idx + 1) % n
	case keyMap.Up.matches(key):
		m.idx = (m.idx - 1 + n) % n
	case keyMap.Right.matches(key):
		if cur.Kind == spec.SelectFieldKind {
			m.cycle(m.idx, 1)
		}
	case keyMap.Left.matches(key):
		if cur.Kind == spec.SelectFieldKind {
			m.cycle(m.idx, -1)
		}
	case keyMap.Edit.matches(key):
		if cur.Kind == spec.SelectFieldKind {
			m.setStatus(strings.ToLower(cur.Label)+" is a list · use "+keyMap.Cycle.hint, kindWarn)
		} else {
			m.insert = true
			m.setStatus("editing "+strings.ToLower(cur.Label)+" · "+keyMap.Cancel.hint+" when done", kindIdle)
		}
	case keyMap.Secret.matches(key):
		if cur.Kind == spec.HiddenFieldKind {
			m.reveal = !m.reveal
			if m.reveal {
				m.setStatus("password visible · "+keyMap.Secret.hint+" to hide", kindIdle)
			} else {
				m.setStatus("password hidden", kindIdle)
			}
		}
	case keyMap.Confirm.matches(key):
		return m.submit()
	case keyMap.Help.matches(key):
		m.help = !m.help
		m.setStatus(keyhintStatus(m.help), kindIdle)
	}
	return m, nil
}

func (m secretFormModel) insertKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	cur := m.spec.Fields[m.idx]

	key := msg.String()

	switch {
	case keyMap.Interrupt.matches(key):
		return m, tea.Quit
	case keyMap.Cancel.matches(key):
		m.insert = false
		m.setStatus("done editing · "+keyMap.MoveVertical.hint+" to move", kindIdle)
		return m, nil
	case keyMap.Confirm.matches(key):
		m.insert = false
		if m.idx < len(m.spec.Fields)-1 {
			m.idx++
		}
		return m, nil
	case keyMap.Backspace.matches(key):
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

func (m *secretFormModel) setStatus(s string, k statusKind) {
	m.status, m.statusKind = s, k
}

func (m *secretFormModel) cycle(i, dir int) {
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

func (m secretFormModel) fieldError(i int) string {
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

func (m secretFormModel) submit() (tea.Model, tea.Cmd) {
	m.attempted = true

	bad, badIdx := 0, -1
	for i := range m.spec.Fields {
		if m.fieldError(i) != "" {
			bad++
			if badIdx < 0 {
				badIdx = i
			}
		}
	}

	if bad > 0 {
		m.idx, m.insert = badIdx, false
		m.setStatus(fmt.Sprintf("%d %s need attention", bad, plural(bad, "field", "fields")), kindErr)
		return m, nil
	}

	m.submitted = true
	return m, tea.Quit
}

// result returns the record and, separately, the password typed for it. The two
// never travel together: config.Secret holds the reference, the material is
// handed to the repository alongside it.
func (m secretFormModel) result() (config.Secret, string, error) {
	sec, err := m.spec.BuildFunc(m.values())
	if err != nil {
		return nil, "", err
	}
	return sec, m.vals[spec.KeyringPasswordKey], nil
}

// values collects the current field values keyed by FormField.Key. Every value
// is trimmed except hidden fields (passwords), where surrounding whitespace may
// be meaningful.
func (m secretFormModel) values() map[spec.FormFieldKey]spec.FormFieldValue {
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
