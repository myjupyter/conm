package ui

import (
	"fmt"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/myjupyter/conm/internal/network"
)

const tableHelp = "↑/k up · ↓/j down · enter · a add · e edit · d del · p ping · q/esc quit"

func (m Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m Model) render() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Postgres connections"))
	b.WriteByte('\n')

	if m.reg.Len() == 0 {
		b.WriteString(itemStyle.Render("No connections found."))
		b.WriteByte('\n')
		m.writeErrors(&b)
		b.WriteString(helpStyle.Render("a add · q/esc quit"))
		return b.String()
	}

	b.WriteString(m.renderTable())
	b.WriteByte('\n')
	m.writeErrors(&b)

	if m.confirming {
		b.WriteString(helpStyle.Render(fmt.Sprintf("delete %q? y/n", m.cursorLabel())))
	} else {
		b.WriteString(helpStyle.Render(tableHelp))
	}

	return b.String()
}

func (m Model) renderTable() string {
	headers := []string{"NAME", "USERNAME", "HOST", "PORT", "DBNAME", "TAGS"}
	if m.pinged {
		headers = append([]string{"PING"}, headers...)
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(borderStyle).
		Headers(headers...).
		StyleFunc(func(row, _ int) lipgloss.Style {
			switch row {
			case table.HeaderRow:
				return headerStyle
			case m.cursor:
				return selectedStyle
			default:
				return itemStyle
			}
		})

	for i := 0; i < m.reg.Len(); i++ {
		c, ok := m.reg.Get(i)
		if !ok {
			continue
		}
		row := []string{
			c.Name(),
			c.Username(),
			c.Host(),
			strconv.Itoa(c.Port()),
			c.Database(),
			strings.Join(c.Tags(), ", "),
		}
		if m.pinged {
			row = append([]string{m.pingCell(i)}, row...)
		}
		t.Row(row...)
	}

	return t.String()
}

func (m Model) pingCell(i int) string {
	st := m.states[i]
	switch st.pingStatus {
	case pingPinging:
		return st.pingSpinner.View()
	case pingOK:
		return formatPing(st.pingResult)
	case pingFailed:
		return pingFailedMark
	default:
		return pingUndefinedMark
	}
}

const (
	pingUndefinedMark = "•"
	pingFailedMark    = "✗"
	pingBarsTotal     = 4
)

func formatPing(r network.PingResult) string {
	return fmt.Sprintf("%dms %s", r.PingTime.Milliseconds(), pingBars(r.Bars()))
}

var pingBarLevels = [pingBarsTotal]rune{'▁', '▃', '▅', '▇'}

func pingBars(n int) string {
	if n < 0 {
		n = 0
	}
	if n > pingBarsTotal {
		n = pingBarsTotal
	}
	var b strings.Builder
	for i, r := range pingBarLevels {
		if i < n {
			b.WriteRune(r)
		} else {
			b.WriteByte(' ')
		}
	}
	return b.String()
}

func (m Model) writeErrors(b *strings.Builder) {
	if m.runErr != nil {
		b.WriteString(errorStyle.Render("run failed: " + m.runErr.Error()))
		b.WriteByte('\n')
	}
	if m.changeErr != nil {
		b.WriteString(errorStyle.Render(m.changeErr.Error()))
		b.WriteByte('\n')
	}
}
