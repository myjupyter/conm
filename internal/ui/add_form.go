package ui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/registry"
	"github.com/myjupyter/conm/internal/ui/spec"
)

func RunAddForm(cfg config.Conm, t config.ConnType) (bool, error) {
	conn, ok, err := runAddForm(t)
	if err != nil || !ok {
		return false, err
	}

	pg, ok := conn.(config.Postgres)
	if !ok {
		return false, fmt.Errorf("unexpected connection type %T for postgres add form", conn)
	}

	reg, err := registry.NewPostgresRegistry(cfg)
	if err != nil {
		return false, err
	}
	defer reg.Close()

	if err := reg.Add(pg); err != nil {
		return false, err
	}

	return true, nil
}

func runAddForm(t config.ConnType) (config.Connection, bool, error) {
	formSpec, ok := spec.FormSpecs[t]
	if !ok {
		return nil, false, fmt.Errorf("add form is not implemented for connection type %q", t)
	}

	return runForm(newFormModel(formSpec, formSpec.AddTitle, nil))
}

func runForm(model formModel) (config.Connection, bool, error) {
	m, err := tea.NewProgram(model).Run()
	if err != nil {
		return nil, false, err
	}

	fm := m.(formModel)
	if !fm.submitted {
		return nil, false, nil
	}

	conn, err := fm.result()
	if err != nil {
		return nil, false, err
	}

	return conn, true, nil
}

type formModel struct {
	spec   spec.FormSpec
	title  string
	inputs []textinput.Model

	// selects holds the chosen option index for each SelectFieldKind field,
	// keyed by its position in spec.Fields.
	selects map[int]int

	focus     int
	submitted bool
	err       string
}

func newFormModel(spc spec.FormSpec, title string, initial map[spec.FormFieldKey]spec.FormFieldValue) formModel {
	m := formModel{
		spec:    spc,
		title:   title,
		inputs:  make([]textinput.Model, len(spc.Fields)),
		selects: make(map[int]int),
	}

	const inputWidth = 40

	for i, f := range spc.Fields {
		if f.Kind == spec.SelectFieldKind {
			m.selects[i] = seedSelectIndex(f, initial)
			continue
		}

		in := textinput.New()
		in.Prompt = ""
		in.Placeholder = f.Example
		in.SetWidth(inputWidth)

		// Highlight the whole field (text + padding) while it is focused
		st := in.Styles()
		st.Focused.Text = st.Focused.Text.Foreground(highlightFg).Background(highlightBg)
		st.Focused.Placeholder = st.Focused.Placeholder.Foreground(highlightPlaceholderFg).Background(highlightBg)
		in.SetStyles(st)

		if f.Kind == spec.HiddenFieldKind {
			in.EchoMode = textinput.EchoPassword
		}

		if v, ok := initial[f.Key]; ok {
			in.SetValue(v)
		}

		m.inputs[i] = in
	}

	m.setFocus(0)

	return m
}

func defaultSelectIndex(f spec.FormField) int {
	for i, opt := range f.Options {
		if opt == f.DefaultValue {
			return i
		}
	}
	return 0
}

func seedSelectIndex(f spec.FormField, initial map[spec.FormFieldKey]spec.FormFieldValue) int {
	if v, ok := initial[f.Key]; ok {
		for i, opt := range f.Options {
			if opt == v {
				return i
			}
		}
	}
	return defaultSelectIndex(f)
}

func (m formModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "ctrl+s":
			return m.submit()
		case "tab", "down":
			return m.focusNext(), nil
		case "shift+tab", "up":
			return m.focusPrev(), nil
		case "enter":
			if m.isLast() {
				return m.submit()
			}
			return m.focusNext(), nil
		case "left":
			if m.onSelect() {
				n := len(m.spec.Fields[m.focus].Options)
				m.selects[m.focus] = (m.selects[m.focus] - 1 + n) % n
				return m, nil
			}
		case "right":
			if m.onSelect() {
				n := len(m.spec.Fields[m.focus].Options)
				m.selects[m.focus] = (m.selects[m.focus] + 1) % n
				return m, nil
			}
		}
	}

	if !m.onSelect() {
		var cmd tea.Cmd
		m.inputs[m.focus], cmd = m.inputs[m.focus].Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m formModel) onSelect() bool {
	return m.spec.Fields[m.focus].Kind == spec.SelectFieldKind
}

func (m formModel) isLast() bool {
	return m.focus == len(m.spec.Fields)-1
}

func (m formModel) focusNext() formModel {
	m.setFocus((m.focus + 1) % len(m.spec.Fields))
	return m
}

func (m formModel) focusPrev() formModel {
	m.setFocus((m.focus - 1 + len(m.spec.Fields)) % len(m.spec.Fields))
	return m
}

func (m *formModel) setFocus(next int) {
	if m.spec.Fields[m.focus].Kind != spec.SelectFieldKind {
		m.inputs[m.focus].Blur()
	}
	m.focus = next
	if m.spec.Fields[next].Kind != spec.SelectFieldKind {
		m.inputs[next].Focus()
	}
}

func (m formModel) submit() (tea.Model, tea.Cmd) {
	values := m.values()
	for _, f := range m.spec.Fields {
		v := values[f.Key]

		if f.Property == spec.RequiredFieldProperty && strings.TrimSpace(v) == "" {
			m.err = strings.ToLower(f.Label) + " is required"
			return m, nil
		}

		if f.ValidateFunc != nil {
			if err := f.ValidateFunc(v); err != nil {
				m.err = err.Error()
				return m, nil
			}
		}
	}

	m.submitted = true
	return m, tea.Quit
}

func (m formModel) result() (config.Connection, error) {
	return m.spec.BuildFunc(m.values())
}

// values collects the current field values keyed by FormField.Key. Every value
// is trimmed except hidden fields (passwords), where surrounding whitespace may
// be meaningful.
func (m formModel) values() map[spec.FormFieldKey]spec.FormFieldValue {
	out := make(map[spec.FormFieldKey]spec.FormFieldValue, len(m.spec.Fields))
	for i, f := range m.spec.Fields {
		v := m.fieldValue(i)
		if f.Kind != spec.HiddenFieldKind {
			v = strings.TrimSpace(v)
		}
		out[f.Key] = v
	}
	return out
}

func (m formModel) fieldValue(i int) string {
	f := m.spec.Fields[i]
	if f.Kind == spec.SelectFieldKind {
		return f.Options[m.selects[i]]
	}
	return m.inputs[i].Value()
}

func (m formModel) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m formModel) render() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(m.title))
	b.WriteByte('\n')

	for i, f := range m.spec.Fields {
		focused := m.focus == i

		b.WriteString(" " + labelStyle.Render(f.Label) + requiredMark(f.Property))
		b.WriteByte('\n')

		if focused {
			b.WriteString(markerStyle.Render(">"))
		} else {
			b.WriteString(" ")
		}
		if f.Kind == spec.SelectFieldKind {
			b.WriteString(selectField(f.Options, m.selects[i], focused))
		} else {
			b.WriteString(m.inputs[i].View())
		}
		b.WriteString("\n\n")
	}

	if m.err != "" {
		b.WriteString(errorStyle.Render(m.err))
		b.WriteByte('\n')
	}

	b.WriteString(helpStyle.Render("tab/↑↓ move · ←/→ select · enter submit · esc cancel"))
	return b.String()
}

func requiredMark(p spec.FieldProperty) string {
	if p == spec.RequiredFieldProperty {
		return requiredStyle.Render(" *")
	}
	return ""
}

func selectField(options []string, cursor int, focused bool) string {
	value := "‹ " + options[cursor] + " ›"
	if focused {
		return markerStyle.Render(value)
	}
	return value
}
