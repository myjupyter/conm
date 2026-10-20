package ui

import (
	"context"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/i18n"
	"github.com/myjupyter/conm/internal/network"
	"github.com/myjupyter/conm/internal/ui/spec"
)

type transportFormModel struct {
	spec    spec.FormSpec[config.Transport]
	title   string
	isEdit  bool
	sources spec.Sources
	pingVia func(ctx context.Context, name string) (network.PingResult, error)

	order []int
	vals  map[string]string

	idx    int
	insert bool
	help   bool

	attempted bool
	submitted bool

	status     string
	statusKind statusKind
}

func newTransportFormModel(
	spc spec.FormSpec[config.Transport],
	title string,
	initial formValues,
	isEdit bool,
	ctx transportFormContext,
) transportFormModel {
	m := transportFormModel{
		spec:       spc,
		title:      title,
		isEdit:     isEdit,
		sources:    ctx.sources,
		pingVia:    ctx.pingVia,
		vals:       make(map[string]string, len(spc.Fields)),
		status:     statusReady,
		statusKind: kindIdle,
	}

	for _, s := range spc.Sections {
		for _, key := range s.Fields {
			if i := m.fieldIndex(key); i >= 0 {
				m.order = append(m.order, i)
			}
		}
	}

	for i := range spc.Fields {
		f := &spc.Fields[i]
		seeded, ok := initial[f.Key]
		switch {
		case ok:
			m.vals[f.Key] = seeded
		case isSelect(*f):
			m.vals[f.Key] = f.DefaultValue
			if options := m.options(*f); m.vals[f.Key] == "" && len(options) > 0 {
				m.vals[f.Key] = options[0]
			}
		default:
			m.vals[f.Key] = ""
		}
	}

	return m
}

func (m transportFormModel) Init() tea.Cmd { return nil }

func (m transportFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if keyMap.Paste.matches(msg.String()) {
			return m, readClipboardCmd()
		}
		if m.insert {
			return m.insertKey(msg)
		}
		return m.navKey(msg)
	case transportPingMsg:
		return m.applyPing(msg), nil
	case clipboardMsg:
		return m.applyPaste(msg.text, msg.err), nil
	case tea.PasteMsg:
		return m.applyPaste(msg.Content, nil), nil
	}
	return m, nil
}

func (m transportFormModel) fieldIndex(key spec.FormFieldKey) int {
	for i := range m.spec.Fields {
		if m.spec.Fields[i].Key == key {
			return i
		}
	}
	return -1
}

func (m transportFormModel) current() spec.FormField {
	return m.spec.Fields[m.order[m.idx]]
}

func (m transportFormModel) options(f spec.FormField) []string {
	switch {
	case f.Kind == spec.SourceFieldKind && f.Property == spec.OptionalFieldProperty:
		return append([]string{""}, m.sources[f.Source.Name]...)
	case f.Kind == spec.SourceFieldKind:
		return m.sources[f.Source.Name]
	default:
		return f.Options
	}
}

func (m transportFormModel) isMissing(f spec.FormField) bool {
	value := m.vals[f.Key]
	return f.Kind == spec.SourceFieldKind && value != "" && !slices.Contains(m.options(f), value)
}

func (m transportFormModel) navKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	n := len(m.order)
	cur := m.current()

	key := msg.String()

	switch {
	case keyMap.Interrupt.matches(key), keyMap.Cancel.matches(key):
		return m, tea.Quit
	case keyMap.Down.matches(key):
		m.idx = (m.idx + 1) % n
	case keyMap.Up.matches(key):
		m.idx = (m.idx - 1 + n) % n
	case keyMap.Right.matches(key):
		m.cycle(cur, 1)
	case keyMap.Left.matches(key):
		m.cycle(cur, -1)
	case keyMap.Edit.matches(key):
		m.edit(cur)
	case keyMap.Ping.matches(key):
		return m.pingViaCmd()
	case keyMap.Confirm.matches(key):
		return m.submit()
	case keyMap.Help.matches(key):
		m.help = !m.help
		m.setStatus(keyhintStatus(m.help), kindIdle)
	}
	return m, nil
}

func (m *transportFormModel) edit(cur spec.FormField) {
	switch {
	case isSelect(cur) && len(m.options(cur)) > 0:
		m.setStatus(i18n.Sprintf("%s is a list · use %s", strings.ToLower(cur.Label), keyMap.Cycle.hint), kindWarn)
	case isSelect(cur):
		m.setStatus(i18n.Sprintf("%s has no options to pick from", strings.ToLower(cur.Label)), kindWarn)
	default:
		m.insert = true
		m.setStatus(i18n.Sprintf("editing %s · %s when done", strings.ToLower(cur.Label), keyMap.Cancel.hint), kindIdle)
	}
}

func (m transportFormModel) applyPaste(raw string, err error) transportFormModel {
	text, refusal := pasteReady(raw, err)
	if text == "" {
		m.setStatus(refusal.text, refusal.kind)
		return m
	}

	cur := m.current()
	if isSelect(cur) {
		m.setStatus(i18n.Sprintf("%s is a list · use %s", strings.ToLower(cur.Label), keyMap.Cycle.hint), kindWarn)
		return m
	}

	m.insert = true
	m.vals[cur.Key] += text
	m.setStatus(i18n.Sprintf("pasted · %s when done", keyMap.Cancel.hint), kindIdle)
	return m
}

func (m transportFormModel) insertKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	cur := m.current()

	key := msg.String()

	switch {
	case keyMap.Interrupt.matches(key):
		return m, tea.Quit
	case keyMap.Cancel.matches(key):
		m.insert = false
		m.setStatus(i18n.Sprintf("done editing · %s to move", keyMap.MoveVertical.hint), kindIdle)
		return m, nil
	case keyMap.Confirm.matches(key):
		m.insert = false
		if m.idx < len(m.order)-1 {
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

func (m *transportFormModel) cycle(f spec.FormField, dir int) {
	options := m.options(f)
	if !isSelect(f) || len(options) == 0 {
		return
	}

	at := max(slices.Index(options, m.vals[f.Key]), 0)
	m.vals[f.Key] = options[(at+dir+len(options))%len(options)]
}

func (m *transportFormModel) setStatus(s string, k statusKind) {
	m.status, m.statusKind = s, k
}

func (m transportFormModel) fieldError(i int) string {
	f := m.spec.Fields[i]
	v := strings.TrimSpace(m.vals[f.Key])

	if f.Property == spec.RequiredFieldProperty && v == "" {
		return i18n.Sprintf("%s is required", strings.ToLower(f.Label))
	}
	if m.isMissing(f) {
		return i18n.Sprintf("%s is no longer available · pick another with %s", v, keyMap.Cycle.hint)
	}
	if f.ValidateFunc != nil {
		if err := f.ValidateFunc(v); err != nil {
			return err.Error()
		}
	}
	return ""
}

func (m transportFormModel) submit() (tea.Model, tea.Cmd) {
	m.attempted = true

	bad, badAt := 0, -1
	for at, i := range m.order {
		if m.fieldError(i) != "" {
			bad++
			if badAt < 0 {
				badAt = at
			}
		}
	}

	if bad > 0 {
		m.idx, m.insert = badAt, false
		m.setStatus(i18n.Sprintf("%d fields need attention", bad), kindErr)
		return m, nil
	}

	m.submitted = true
	return m, tea.Quit
}

func (m transportFormModel) applyPing(msg transportPingMsg) transportFormModel {
	if msg.err != nil {
		m.setStatus(i18n.Sprintf("ping failed · %s · %s", msg.via, msg.err), kindErr)
		return m
	}

	m.setStatus(i18n.Sprintf("pong · %s answered in %dms", msg.via, msg.result.PingTime.Milliseconds()), kindOK)
	return m
}

func (m transportFormModel) values() formValues {
	out := make(formValues, len(m.spec.Fields))
	for i := range m.spec.Fields {
		key := m.spec.Fields[i].Key
		out[key] = strings.TrimSpace(m.vals[key])
	}
	return out
}
