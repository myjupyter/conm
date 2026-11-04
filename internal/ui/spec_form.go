package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/ui/spec"
	"github.com/myjupyter/conm/internal/ui/view"
)

type specFormModel[T any] struct {
	size screenSize
	body view.Scroll

	spec   spec.FormSpec[T]
	badge  string
	noun   string
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

func newSpecFormModel[T any](spc spec.FormSpec[T], badge, noun, title string, initial formValues, isEdit bool) specFormModel[T] {
	m := specFormModel[T]{
		spec:       spc,
		badge:      badge,
		noun:       noun,
		title:      title,
		isEdit:     isEdit,
		vals:       make(map[string]string, len(spc.Fields)),
		status:     statusReady,
		statusKind: kindIdle,
	}

	for i := range spc.Fields {
		f := &spc.Fields[i]
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

func (m specFormModel[T]) Init() tea.Cmd { return nil }

func (m specFormModel[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.size.observe(msg)
	return followed[specFormModel[T]](m.update(msg))
}

func (m specFormModel[T]) followCursor() specFormModel[T] {
	m.size.follow(&m.body, len(m.headLines(""))+len(m.tailLines()), focusSpan{from: 2 * m.idx, row: 2 * m.idx, to: 2*m.idx + 1}, len(m.bodyLines()))
	return m
}

func (m specFormModel[T]) update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if keyMap.Paste.matches(msg.String()) {
			return m, readClipboardCmd()
		}
		if m.insert {
			return m.insertKey(msg)
		}
		return m.navKey(msg)
	case clipboardMsg:
		return m.applyPaste(msg.text, msg.err), nil
	case tea.PasteMsg:
		return m.applyPaste(msg.Content, nil), nil
	}
	return m, nil
}

func (m specFormModel[T]) applyPaste(raw string, err error) specFormModel[T] {
	text, refusal := pasteReady(raw, err)
	if text == "" {
		m.setStatus(refusal.text, refusal.kind)
		return m
	}

	cur := m.spec.Fields[m.idx]
	if cur.Kind == spec.SelectFieldKind {
		m.setStatus(strings.ToLower(cur.Label)+" is a list · use "+keyMap.Cycle.hint, kindWarn)
		return m
	}

	m.insert = true
	m.vals[cur.Key] += text
	m.setStatus("pasted · "+keyMap.Cancel.hint+" when done", kindIdle)
	return m
}

func (m specFormModel[T]) navKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
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
	case keyMap.Form.EditField.matches(key):
		if cur.Kind == spec.SelectFieldKind {
			m.setStatus(strings.ToLower(cur.Label)+" is a list · use "+keyMap.Cycle.hint, kindWarn)
		} else {
			m.insert = true
			m.setStatus("editing "+strings.ToLower(cur.Label)+" · "+keyMap.Cancel.hint+" when done", kindIdle)
		}
	case keyMap.Screen.Secrets.matches(key):
		if cur.Kind == spec.HiddenFieldKind {
			m.reveal = !m.reveal
			if m.reveal {
				m.setStatus("password visible · "+keyMap.Screen.Secrets.hint+" to hide", kindIdle)
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

func (m specFormModel[T]) insertKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
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

func (m *specFormModel[T]) setStatus(s string, k statusKind) {
	m.status, m.statusKind = s, k
}

func (m *specFormModel[T]) cycle(i, dir int) {
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

func (m specFormModel[T]) fieldError(i int) string {
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

func (m specFormModel[T]) submit() (tea.Model, tea.Cmd) {
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

func (m specFormModel[T]) values() map[spec.FormFieldKey]spec.FormFieldValue {
	out := make(map[spec.FormFieldKey]spec.FormFieldValue, len(m.spec.Fields))
	for i := range m.spec.Fields {
		f := &m.spec.Fields[i]
		v := m.vals[f.Key]
		if f.Kind != spec.HiddenFieldKind {
			v = strings.TrimSpace(v)
		}
		out[f.Key] = v
	}
	return out
}
