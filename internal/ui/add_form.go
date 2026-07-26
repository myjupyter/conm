package ui

import (
	"strconv"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
)

const (
	defaultPGPort = 5432
	inputWidth    = 40
)

func RunAddForm() (config.Postgres, bool, error) {
	m, err := tea.NewProgram(newAddModel()).Run()
	if err != nil {
		return config.Postgres{}, false, err
	}

	am := m.(addModel)
	if !am.submitted {
		return config.Postgres{}, false, nil
	}

	return am.result(), true, nil
}

type addModel struct {
	fields []FormField
	inputs []textinput.Model

	focus     int
	sslCursor int
	submitted bool
	err       string
}

func newAddModel() addModel {
	m := addModel{
		fields: PostgresFormFields,
		inputs: make([]textinput.Model, len(PostgresFormFields)),
	}

	for i, f := range m.fields {
		in := textinput.New()
		in.Prompt = ""
		in.Placeholder = f.Example
		in.SetWidth(inputWidth)

		// Highlight the whole field (text + padding) while it is focused.
		st := in.Styles()
		st.Focused.Text = st.Focused.Text.Foreground(highlightFg).Background(highlightBg)
		st.Focused.Placeholder = st.Focused.Placeholder.Foreground(highlightPlaceholderFg).Background(highlightBg)
		in.SetStyles(st)

		if f.Name == config.PostgresFormFieldPassword {
			in.EchoMode = textinput.EchoPassword
		}

		m.inputs[i] = in
	}

	m.setFocus(0)

	return m
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
			if m.onSSL() {
				m.sslCursor = (m.sslCursor - 1 + len(SSLModesOrder)) % len(SSLModesOrder)
				return m, nil
			}
		case "right":
			if m.onSSL() {
				m.sslCursor = (m.sslCursor + 1) % len(SSLModesOrder)
				return m, nil
			}
		}
	}

	if !m.onSSL() {
		var cmd tea.Cmd
		m.inputs[m.focus], cmd = m.inputs[m.focus].Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m addModel) onSSL() bool {
	return m.fields[m.focus].Name == config.PostgresFormFieldSSLMode
}

func (m addModel) isLast() bool {
	return m.focus == len(m.fields)-1
}

func (m addModel) focusNext() addModel {
	m.setFocus((m.focus + 1) % len(m.fields))
	return m
}

func (m addModel) focusPrev() addModel {
	m.setFocus((m.focus - 1 + len(m.fields)) % len(m.fields))
	return m
}

func (m *addModel) setFocus(next int) {
	if !m.onSSL() {
		m.inputs[m.focus].Blur()
	}
	m.focus = next
	if !m.onSSL() {
		m.inputs[next].Focus()
	}
}

func (m addModel) submit() (tea.Model, tea.Cmd) {
	for i, f := range m.fields {
		if f.Name == config.PostgresFormFieldSSLMode {
			continue // always has a valid value
		}
		if f.Name == config.PostgresFormFieldPort {
			if port := strings.TrimSpace(m.inputs[i].Value()); port != "" {
				if _, err := strconv.Atoi(port); err != nil {
					m.err = "port must be a number"
					return m, nil
				}
			}
			continue // empty port falls back to the default
		}
		if f.Property == RequiredFieldProperty && strings.TrimSpace(m.inputs[i].Value()) == "" {
			m.err = strings.ToLower(f.Name) + " is required"
			return m, nil
		}
	}

	m.submitted = true
	return m, tea.Quit
}

func (m addModel) result() config.Postgres {
	port := defaultPGPort
	if raw := m.value(config.PostgresFormFieldPort); raw != "" {
		if p, err := strconv.Atoi(raw); err == nil {
			port = p
		}
	}

	return config.Postgres{
		Meta: config.ConnMeta{
			Name:        m.value(config.PostgresFormFieldName),
			Description: m.value(config.PostgresFormFieldDescription),
			Tags:        parseTags(m.value(config.PostgresFormFieldTags)),
		},
		Hostname:   m.value(config.PostgresFormFieldHost),
		PortNumber: port,
		User:       m.value(config.PostgresFormFieldUsername),
		Password:   m.rawValue(config.PostgresFormFieldPassword),
		DBName:     m.value(config.PostgresFormFieldDatabase),
		SSLMode:    string(SSLModesOrder[m.sslCursor]),
	}
}

func (m addModel) value(name string) string {
	return strings.TrimSpace(m.rawValue(name))
}

func (m addModel) rawValue(name string) string {
	for i, f := range m.fields {
		if f.Name == name {
			return m.inputs[i].Value()
		}
	}
	return ""
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

	b.WriteString(titleStyle.Render("Add a Postgres connection"))
	b.WriteByte('\n')

	for i, f := range m.fields {
		focused := m.focus == i

		b.WriteString(" " + labelStyle.Render(f.Name) + requiredMark(f.Property))
		b.WriteByte('\n')

		if focused {
			b.WriteString(markerStyle.Render(">"))
		} else {
			b.WriteString(" ")
		}
		if f.Name == config.PostgresFormFieldSSLMode {
			b.WriteString(sslModeField(m.sslCursor, focused))
		} else {
			b.WriteString(m.inputs[i].View())
		}
		b.WriteString("\n\n")
	}

	if m.err != "" {
		b.WriteString(errorStyle.Render(m.err))
		b.WriteByte('\n')
	}

	b.WriteString(helpStyle.Render("tab/↑↓ move · ←/→ ssl mode · enter/ctrl+s submit · esc cancel"))
	return b.String()
}

func requiredMark(p FieldProperty) string {
	if p == RequiredFieldProperty {
		return requiredStyle.Render(" *")
	}
	return ""
}

func sslModeField(cursor int, focused bool) string {
	value := "‹ " + string(SSLModesOrder[cursor]) + " ›"
	if focused {
		return markerStyle.Render(value)
	}
	return value
}
