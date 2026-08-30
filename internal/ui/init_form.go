package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui/spec"
)

// initFormModel edits one database: the client conm drives it with, and
// whether it is enabled at all. Every field is a select, so the form has no
// insert mode — a value is cycled, never typed.
type initFormModel struct {
	spec spec.FormSpec[config.Database]
	kind config.ConnType

	vals map[spec.FormFieldKey]spec.FormFieldValue

	idx  int
	help bool

	attempted bool
	submitted bool

	status     string
	statusKind statusKind
}

func newInitFormModel(spc spec.FormSpec[config.Database], kind config.ConnType, initial map[spec.FormFieldKey]spec.FormFieldValue) initFormModel {
	m := initFormModel{
		spec:       spc,
		kind:       kind,
		vals:       make(map[spec.FormFieldKey]spec.FormFieldValue, len(spc.Fields)),
		status:     statusReady,
		statusKind: kindIdle,
	}

	for _, f := range spc.Fields {
		value := initial[f.Key]
		if value == "" {
			value = f.DefaultValue
		}
		if value == "" && len(f.Options) > 0 {
			value = f.Options[0]
		}
		m.vals[f.Key] = value
	}

	return m
}

func (m initFormModel) Init() tea.Cmd { return nil }

func (m initFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}

	n := len(m.spec.Fields)

	switch k := key.String(); {
	case keyMap.Interrupt.matches(k), keyMap.Cancel.matches(k):
		return m, tea.Quit
	case keyMap.Down.matches(k):
		m.idx = (m.idx + 1) % n
	case keyMap.Up.matches(k):
		m.idx = (m.idx - 1 + n) % n
	case keyMap.Right.matches(k):
		m.cycle(m.idx, 1)
	case keyMap.Left.matches(k):
		m.cycle(m.idx, -1)
	case keyMap.Confirm.matches(k):
		return m.submit()
	case keyMap.Help.matches(k):
		m.help = !m.help
		m.setFormStatus(keyhintStatus(m.help), kindIdle)
	}

	return m, nil
}

func (m *initFormModel) cycle(i, dir int) {
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

func (m *initFormModel) setFormStatus(s string, k statusKind) {
	m.status, m.statusKind = s, k
}

func (m initFormModel) fieldError(i int) string {
	f := m.spec.Fields[i]
	value := m.vals[f.Key]

	if f.Property == spec.RequiredFieldProperty && value == "" {
		return strings.ToLower(f.Label) + " is required"
	}
	if f.ValidateFunc != nil {
		if err := f.ValidateFunc(value); err != nil {
			return err.Error()
		}
	}
	return ""
}

func (m initFormModel) submit() (tea.Model, tea.Cmd) {
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
		m.idx = badIdx
		m.setFormStatus(fmt.Sprintf("%d %s need attention", bad, plural(bad, "field", "fields")), kindErr)
		return m, nil
	}

	m.submitted = true
	return m, tea.Quit
}

func (m initFormModel) result() (config.Database, error) {
	return m.spec.BuildFunc(m.vals)
}
