package ui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
)

func RunAddForm(t config.ConnType) (config.ConnectionConfig, bool, error) {
	spec, ok := formSpecs[t]
	if !ok {
		return nil, false, fmt.Errorf("add form is not implemented for connection type %q", t)
	}

	m, err := tea.NewProgram(newAddModel(spec)).Run()
	if err != nil {
		return nil, false, err
	}

	am := m.(addModel)
	if !am.submitted {
		return nil, false, nil
	}

	conn, err := am.result()
	if err != nil {
		return nil, false, err
	}

	return conn, true, nil
}

type addModel struct {
	spec   FormSpec
	inputs []textinput.Model

	// selects holds the chosen option index for each SelectFieldKind field,
	// keyed by its position in spec.Fields.
	selects map[int]int

	focus     int
	submitted bool
	err       string
}

func newAddModel(spec FormSpec) addModel {
	m := addModel{
		spec:    spec,
		inputs:  make([]textinput.Model, len(spec.Fields)),
		selects: make(map[int]int),
	}

	const inputWidth = 40

	for i, f := range spec.Fields {
		if f.Kind == SelectFieldKind {
			m.selects[i] = defaultSelectIndex(f)
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

		if f.Kind == HiddenFieldKind {
			in.EchoMode = textinput.EchoPassword
		}

		m.inputs[i] = in
	}

	m.setFocus(0)

	return m
}

func defaultSelectIndex(f FormField) int {
	for i, opt := range f.Options {
		if opt == f.DefaultValue {
			return i
		}
	}
	return 0
}

func (m addModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m addModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m addModel) onSelect() bool {
	return m.spec.Fields[m.focus].Kind == SelectFieldKind
}

func (m addModel) isLast() bool {
	return m.focus == len(m.spec.Fields)-1
}

func (m addModel) focusNext() addModel {
	m.setFocus((m.focus + 1) % len(m.spec.Fields))
	return m
}

func (m addModel) focusPrev() addModel {
	m.setFocus((m.focus - 1 + len(m.spec.Fields)) % len(m.spec.Fields))
	return m
}

func (m *addModel) setFocus(next int) {
	if m.spec.Fields[m.focus].Kind != SelectFieldKind {
		m.inputs[m.focus].Blur()
	}
	m.focus = next
	if m.spec.Fields[next].Kind != SelectFieldKind {
		m.inputs[next].Focus()
	}
}

func (m addModel) submit() (tea.Model, tea.Cmd) {
	values := m.values()
	for _, f := range m.spec.Fields {
		v := values[f.Key]

		if f.Property == RequiredFieldProperty && strings.TrimSpace(v) == "" {
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

func (m addModel) result() (config.ConnectionConfig, error) {
	return m.spec.BuildFunc(m.values())
}

// values collects the current field values keyed by FormField.Key. Every value
// is trimmed except hidden fields (passwords), where surrounding whitespace may
// be meaningful.
func (m addModel) values() map[FormFieldKey]FormFieldValue {
	out := make(map[FormFieldKey]FormFieldValue, len(m.spec.Fields))
	for i, f := range m.spec.Fields {
		v := m.fieldValue(i)
		if f.Kind != HiddenFieldKind {
			v = strings.TrimSpace(v)
		}
		out[f.Key] = v
	}
	return out
}

func (m addModel) fieldValue(i int) string {
	f := m.spec.Fields[i]
	if f.Kind == SelectFieldKind {
		return f.Options[m.selects[i]]
	}
	return m.inputs[i].Value()
}

func parseTags(raw string) []string {
	var tags []string
	for t := range strings.SplitSeq(raw, ",") {
		if t = strings.TrimSpace(t); t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}

func (m addModel) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m addModel) render() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(m.spec.Title))
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
		if f.Kind == SelectFieldKind {
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

func requiredMark(p FieldProperty) string {
	if p == RequiredFieldProperty {
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
