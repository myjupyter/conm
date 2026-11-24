package ui

import (
	"slices"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui/view"
)

const tableSettingsWindow = 15

type tableTab struct {
	cursor  int
	window  view.Scroll
	editing bool
	draft   string
}

type tableRow struct {
	connType config.ConnType
	column   string
}

func (r tableRow) isColor() bool { return r.column == "" }

const tabColorLabel = "tab colour"

var tabColorChoices = slices.Concat(config.ConnTypeNames(), config.ColorNames)

func (m settingsModel) tableRows() []tableRow {
	var rows []tableRow
	for _, s := range m.conm.Connections {
		rows = append(rows, tableRow{connType: s.Type})
		for _, c := range orderedColumns(s) {
			rows = append(rows, tableRow{connType: s.Type, column: c})
		}
	}
	return rows
}

func orderedColumns(s config.ConnectionSettings) []string {
	visible := s.VisibleColumns()
	ordered := slices.Clone(visible)
	for _, c := range s.Type.Columns() {
		if !slices.Contains(visible, c) {
			ordered = append(ordered, c)
		}
	}
	return ordered
}

func (m settingsModel) tableCursorFor(s config.ConnectionSettings, column string) int {
	for i, r := range m.tableRows() {
		if r.connType == s.Type && r.isColor() {
			return i + 1 + slices.Index(orderedColumns(s), column)
		}
	}
	return m.table.cursor
}

func (m settingsModel) tableRow() tableRow {
	rows := m.tableRows()
	return rows[min(m.table.cursor, len(rows)-1)]
}

func (m settingsModel) typeSettings(t config.ConnType) config.ConnectionSettings {
	s, _ := m.conm.Connection(t)
	s.Columns = slices.Clone(s.Columns)
	return s
}

func (m settingsModel) handleTableKey(key string) (tea.Model, tea.Cmd, bool) {
	n := len(m.tableRows())
	row := m.tableRow()
	switch {
	case keyMap.Down.matches(key):
		m.table.cursor = (m.table.cursor + 1) % n
	case keyMap.Up.matches(key):
		m.table.cursor = (m.table.cursor - 1 + n) % n
	case keyMap.Right.matches(key):
		model, cmd := m.stepTableRow(1)
		return model, cmd, true
	case keyMap.Left.matches(key):
		model, cmd := m.stepTableRow(-1)
		return model, cmd, true
	case keyMap.Confirm.matches(key) && !row.isColor():
		model, cmd := m.toggleColumn()
		return model, cmd, true
	case keyMap.Settings.EditValue.matches(key) && row.isColor():
		m.table.editing, m.table.draft = true, ""
		m.setSettingsStatus("type a type, tab, a colour name or #rrggbb", kindIdle)
		return m, nil, true
	case keyMap.Settings.Reset.matches(key):
		model, cmd := m.resetTableRow()
		return model, cmd, true
	default:
		return m, nil, false
	}

	m.setSettingsStatus(statusReady, kindIdle)
	return m, nil, true
}

func (m settingsModel) stepTableRow(step int) (tea.Model, tea.Cmd) {
	row := m.tableRow()
	s := m.typeSettings(row.connType)
	if row.isColor() {
		n := len(tabColorChoices)
		at := slices.Index(tabColorChoices, s.ResolvedTabColor())
		if at < 0 && step < 0 {
			at = 0
		}
		next := tabColorChoices[((at+step)%n+n)%n]
		s.TabColor = ownColorCleared(row.connType, next)
		return m, m.saveTableCmd(s, row.connType.String()+" "+tabColorLabel+" → "+next)
	}

	columns := slices.Clone(s.VisibleColumns())
	at := slices.Index(columns, row.column)
	if at < 0 {
		m.setSettingsStatus(row.column+" is hidden · show it first", kindWarn)
		return m, nil
	}
	to := at + step
	if to < 0 || to >= len(columns) {
		return m, nil
	}
	columns[at], columns[to] = columns[to], columns[at]
	s.Columns = columns
	m.table.cursor += step
	return m, m.saveTableCmd(s, row.connType.String()+" "+row.column+" is column "+strconv.Itoa(to+1))
}

func (m settingsModel) toggleColumn() (tea.Model, tea.Cmd) {
	row := m.tableRow()
	if row.column == config.NameColumn {
		m.setSettingsStatus("the name column is always shown", kindWarn)
		return m, nil
	}

	s := m.typeSettings(row.connType)
	columns := slices.Clone(s.VisibleColumns())
	status := row.connType.String() + " " + row.column
	if at := slices.Index(columns, row.column); at >= 0 {
		columns = slices.Delete(columns, at, at+1)
		status += " hidden"
	} else {
		columns = append(columns, row.column)
		status += " shown as column " + strconv.Itoa(len(columns))
	}
	s.Columns = columns
	m.table.cursor = m.tableCursorFor(s, row.column)
	return m, m.saveTableCmd(s, status)
}

func (m settingsModel) resetTableRow() (tea.Model, tea.Cmd) {
	row := m.tableRow()
	s := m.typeSettings(row.connType)
	if row.isColor() {
		s.TabColor = ""
		return m, m.saveTableCmd(s, row.connType.String()+" "+tabColorLabel+" → "+row.connType.String())
	}
	s.Columns = nil
	m.table.cursor = m.tableCursorFor(s, row.column)
	return m, m.saveTableCmd(s, row.connType.String()+" columns back to the default order")
}

func (m settingsModel) handleTableInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch {
	case keyMap.Interrupt.matches(key):
		return m, tea.Quit
	case keyMap.Cancel.matches(key):
		m.table.editing = false
		m.setSettingsStatus(statusReady, kindIdle)
	case keyMap.Confirm.matches(key):
		return m.saveTableDraft()
	case keyMap.Backspace.matches(key):
		r := []rune(m.table.draft)
		m.table.draft = string(r[:max(len(r)-1, 0)])
	default:
		m = m.typeTableColor(msg.Text)
	}
	return m, nil
}

func (m settingsModel) typeTableColor(text string) settingsModel {
	draft := m.table.draft + strings.Map(colorRune, strings.ToLower(text))
	m.table.draft = draft[:min(len(draft), colorDraftMax)]
	return m
}

func (m settingsModel) saveTableDraft() (tea.Model, tea.Cmd) {
	if !config.IsTabColor(m.table.draft) {
		m.setSettingsStatus(strconv.Quote(m.table.draft)+" is neither a type, a colour name nor #rrggbb", kindWarn)
		return m, nil
	}
	row := m.tableRow()
	s := m.typeSettings(row.connType)
	s.TabColor = ownColorCleared(row.connType, m.table.draft)
	m.table.editing = false
	return m, m.saveTableCmd(s, row.connType.String()+" "+tabColorLabel+" → "+m.table.draft)
}

func ownColorCleared(t config.ConnType, value string) string {
	if value == t.String() {
		return ""
	}
	return value
}
