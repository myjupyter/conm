package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/ui/spec"
)

type workspaceFormModel struct {
	spec   spec.FormSpec[string]
	title  string
	isEdit bool

	vals map[string]string

	idx    int
	insert bool
	help   bool

	attempted bool
	submitted bool

	status     string
	statusKind statusKind
}

func newWorkspaceFormModel(spc spec.FormSpec[string], title string, initial formValues, isEdit bool) workspaceFormModel {
	m := workspaceFormModel{
		spec:       spc,
		title:      title,
		isEdit:     isEdit,
		vals:       make(map[string]string, len(spc.Fields)),
		status:     statusReady,
		statusKind: kindIdle,
	}

	for i := range spc.Fields {
		key := spc.Fields[i].Key
		m.vals[key] = initial[key]
	}

	return m
}

func (m workspaceFormModel) Init() tea.Cmd { return nil }

func (m workspaceFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m workspaceFormModel) current() spec.FormField {
	return m.spec.Fields[m.idx]
}

func (m workspaceFormModel) navKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	n := len(m.spec.Fields)
	key := msg.String()

	switch {
	case keyMap.Interrupt.matches(key), keyMap.Cancel.matches(key):
		return m, tea.Quit
	case keyMap.Down.matches(key):
		m.idx = (m.idx + 1) % n
	case keyMap.Up.matches(key):
		m.idx = (m.idx - 1 + n) % n
	case keyMap.Form.EditField.matches(key):
		m.insert = true
		m.setStatus("editing "+strings.ToLower(m.current().Label)+" · "+keyMap.Cancel.hint+" when done", kindIdle)
	case keyMap.Confirm.matches(key):
		return m.submit()
	case keyMap.Help.matches(key):
		m.help = !m.help
		m.setStatus(keyhintStatus(m.help), kindIdle)
	}
	return m, nil
}

func (m workspaceFormModel) insertKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	cur := m.current()
	key := msg.String()

	switch {
	case keyMap.Interrupt.matches(key):
		return m, tea.Quit
	case keyMap.Cancel.matches(key):
		m.insert = false
		m.setStatus("done editing · "+keyMap.Confirm.hint+" to save", kindIdle)
		return m, nil
	case keyMap.Confirm.matches(key):
		m.insert = false
		return m.submit()
	case keyMap.Backspace.matches(key):
		if r := []rune(m.vals[cur.Key]); len(r) > 0 {
			m.vals[cur.Key] = string(r[:len(r)-1])
		}
		return m, nil
	}

	m.vals[cur.Key] += msg.Text
	return m, nil
}

func (m workspaceFormModel) applyPaste(raw string, err error) workspaceFormModel {
	text, refusal := pasteReady(raw, err)
	if text == "" {
		m.setStatus(refusal.text, refusal.kind)
		return m
	}

	m.insert = true
	m.vals[m.current().Key] += text
	m.setStatus("pasted · "+keyMap.Cancel.hint+" when done", kindIdle)
	return m
}

func (m *workspaceFormModel) setStatus(s string, k statusKind) {
	m.status, m.statusKind = s, k
}

func (m workspaceFormModel) fieldError(i int) string {
	f := m.spec.Fields[i]
	v := strings.TrimSpace(m.vals[f.Key])

	if v == "" {
		return strings.ToLower(f.Label) + " is required"
	}
	if f.ValidateFunc != nil {
		if err := f.ValidateFunc(v); err != nil {
			return err.Error()
		}
	}
	return ""
}

func (m workspaceFormModel) submit() (tea.Model, tea.Cmd) {
	m.attempted = true

	bad, badAt := 0, -1
	for i := range m.spec.Fields {
		if m.fieldError(i) != "" {
			bad++
			if badAt < 0 {
				badAt = i
			}
		}
	}

	if bad > 0 {
		m.idx, m.insert = badAt, false
		m.setStatus(fmt.Sprintf("%d %s need attention", bad, plural(bad, "field", "fields")), kindErr)
		return m, nil
	}

	m.submitted = true
	return m, tea.Quit
}

func (m workspaceFormModel) values() formValues {
	out := make(formValues, len(m.spec.Fields))
	for i := range m.spec.Fields {
		key := m.spec.Fields[i].Key
		out[key] = strings.TrimSpace(m.vals[key])
	}
	return out
}
