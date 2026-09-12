package ui

import (
	"image/color"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui/spec"
	"github.com/myjupyter/conm/internal/ui/view"
)

const infoLinkWindow = 8

type infoTab int

const (
	infoGeneral infoTab = iota
	infoEndpoint
	infoLinks
	infoTabCount
)

var infoTabTitles = [infoTabCount]string{"general", "endpoint", "links"}

const (
	infoDSNLabel      = "dsn"
	infoEndpointLabel = "endpoint"
	infoCLILabel      = "cli"
	infoVersionLabel  = "version"
	infoPathLabel     = "path"
	infoLinksLabel    = "links"
)

type infoRow struct {
	label  string
	value  string
	tone   color.Color
	head   bool
	note   string
	link   bool
	hidden bool
}

type infoModel struct {
	conns *view.Connections

	client cli.Info
	probed bool

	tab    infoTab
	cursor int
	links  view.Scroll
	yank   bool

	edit bool

	status     string
	statusKind statusKind
	help       bool
}

type clientInfoMsg cli.Info

type infoRanMsg struct{ err error }

func newInfoModel(conns *view.Connections) infoModel {
	m := infoModel{conns: conns, cursor: -1, status: statusReady}
	m.links.Resize(infoLinkWindow)

	return m
}

func (m infoModel) Init() tea.Cmd { return m.clientInfoCmd() }

func (m infoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m.handleInfoKey(msg.String())

	case clientInfoMsg:
		m.client, m.probed = cli.Info(msg), true

	case clipboardWrittenMsg:
		if msg.err != nil {
			m.setInfoStatus("couldn't yank "+msg.label+" · "+msg.err.Error(), kindErr)
			break
		}
		m.setInfoStatus("yanked "+msg.label, kindOK)

	case linkOpenedMsg:
		if msg.err != nil {
			m.setInfoStatus("couldn't open "+msg.name+" · "+msg.err.Error(), kindErr)
			break
		}
		m.setInfoStatus("opened "+msg.name+" in the browser", kindOK)

	case infoRanMsg:
		if msg.err != nil {
			m.setInfoStatus(msg.err.Error(), kindErr)
			break
		}
		m.setInfoStatus("session closed · "+m.title(), kindOK)
	}

	return m, nil
}

func (m infoModel) handleInfoKey(key string) (tea.Model, tea.Cmd) {
	switch {
	case keyMap.Help.matches(key):
		m.help = !m.help
	case keyMap.Cancel.matches(key) && m.cursor >= 0:
		m = m.clearCursor()
	case keyMap.Back.matches(key):
		return m, tea.Quit
	case keyMap.Down.matches(key):
		m = m.move(1)
	case keyMap.Up.matches(key):
		m = m.move(-1)
	case keyMap.NextSection.matches(key), keyMap.Right.matches(key):
		m = m.switchTab(1)
	case keyMap.PrevSection.matches(key), keyMap.Left.matches(key):
		m = m.switchTab(-1)
	case keyMap.Yank.matches(key):
		return m.yankRow()
	case keyMap.Confirm.matches(key):
		return m.openRow()
	case keyMap.Edit.matches(key):
		m.edit = true
		return m, tea.Quit
	}

	return m, nil
}

func (m infoModel) move(step int) infoModel {
	n := m.pickCount()
	if n == 0 {
		return m
	}

	switch {
	case m.cursor >= 0:
		m.cursor = (m.cursor + step + n) % n
	case step > 0:
		m.cursor = 0
	default:
		m.cursor = n - 1
	}

	m.yank = false
	m.setInfoStatus(statusReady, kindIdle)
	if m.tab == infoLinks {
		m.links.Follow(m.cursor, n)
	}

	return m
}

func (m infoModel) tabs() []infoTab {
	tabs := []infoTab{infoGeneral, infoEndpoint}
	if cfg, ok := m.conns.Config(); ok && len(cfg.Meta().Links) > 0 {
		tabs = append(tabs, infoLinks)
	}

	return tabs
}

func (m infoModel) tabTitles() []string {
	tabs := m.tabs()
	titles := make([]string, 0, len(tabs))
	for _, tab := range tabs {
		titles = append(titles, infoTabTitles[tab])
	}

	return titles
}

func (m infoModel) switchTab(step int) infoModel {
	tabs := m.tabs()
	i := slices.Index(tabs, m.tab)
	m.tab = tabs[(i+step+len(tabs))%len(tabs)]
	m.links.Top()

	return m.clearCursor()
}

func (m infoModel) clearCursor() infoModel {
	m.cursor, m.yank = -1, false
	m.setInfoStatus(statusReady, kindIdle)

	return m
}

func (m infoModel) yankRow() (tea.Model, tea.Cmd) {
	row, ok := m.cursorRow()
	switch {
	case !ok:
		m.setInfoStatus("move the cursor onto a field first · "+keyMap.MoveVertical.hint, kindWarn)
	case m.yank:
		m.yank = false
		return m, writeClipboardCmd(row.label, row.value)
	default:
		m.yank = true
		m.setInfoStatus("press again to yank "+row.label, kindIdle)
	}

	return m, nil
}

func (m infoModel) openRow() (tea.Model, tea.Cmd) {
	row, ok := m.cursorRow()
	if !ok || !row.link {
		m.setInfoStatus("connecting to "+m.title()+" …", kindPending)
		return m, m.connectCmd()
	}

	if err := config.ValidateWebLinkURL(row.value); err != nil {
		m.setInfoStatus(row.label+" · "+err.Error(), kindErr)
		return m, nil
	}

	m.setInfoStatus("opening "+row.label+" · "+row.value, kindPending)

	return m, openURLCmd(row.label, row.value)
}

func (m *infoModel) setInfoStatus(s string, k statusKind) {
	m.status, m.statusKind = s, k
}

func (m infoModel) title() string {
	cfg, ok := m.conns.Config()
	if !ok {
		return ""
	}

	return connLabel(cfg)
}

func (m infoModel) rows() []infoRow {
	cfg, ok := m.conns.Config()
	if !ok {
		return nil
	}

	switch m.tab {
	case infoEndpoint:
		head := infoRow{label: infoEndpointLabel, head: true}
		rows := append([]infoRow{head}, specRows(cfg, false)...)

		return append(rows, infoValueRow(infoDSNLabel, infoDSN(cfg)))
	case infoLinks:
		return m.linkRows(cfg.Meta().Links)
	default:
		head := infoRow{label: spec.MetadataSectionTitle, head: true}
		return append(append([]infoRow{head}, specRows(cfg, true)...), m.clientRows()...)
	}
}

func specRows(cfg config.Connection, meta bool) []infoRow {
	sp, ok := spec.FormSpecs[cfg.ConnType()]
	if !ok {
		return nil
	}

	values := sp.SeedFunc(cfg)
	labels := make(map[spec.FormFieldKey]string, len(sp.Fields))
	for _, f := range sp.Fields {
		labels[f.Key] = strings.ToLower(f.Label)
	}

	var rows []infoRow
	for _, section := range sp.Sections {
		if (section.Title == spec.MetadataSectionTitle) != meta {
			continue
		}
		for _, key := range section.Fields {
			if key == spec.SecretProviderKey || key == spec.SecretValueKey {
				continue
			}
			rows = append(rows, infoValueRow(labels[key], values[key]))
		}
	}

	return rows
}

func (m infoModel) clientRows() []infoRow {
	name, version, path, tone := m.client.Name, m.client.Version, m.client.Path, cAccent
	switch {
	case !m.probed:
		name, version, path, tone = gEllipsis, gEllipsis, gEllipsis, cFaint
	case name == "":
		name, version, path, tone = "no client configured", gEmpty, gEmpty, cAmber
	case !m.client.Installed:
		version, path, tone = "not installed", "not on PATH", cAmber
	case version == "":
		version = "unknown"
	}

	return []infoRow{
		{label: infoCLILabel, head: true},
		{label: infoCLILabel, value: name, tone: tone},
		{label: infoVersionLabel, value: version, tone: cFg},
		{label: infoPathLabel, value: path, tone: cFg},
	}
}

func (m infoModel) linkRows(links []config.WebLink) []infoRow {
	start, end := m.links.Range(len(links))
	head := infoRow{
		label: infoLinksLabel,
		head:  true,
		note:  scrollNote(m.links.Above(), m.links.Below(len(links))),
	}

	rows := make([]infoRow, 0, 1+len(links))
	rows = append(rows, head)
	for i, l := range links {
		rows = append(rows, infoRow{
			label:  l.Name,
			value:  l.URL,
			tone:   cFg,
			link:   true,
			hidden: i < start || i >= end,
		})
	}

	return rows
}

func infoDSN(cfg config.Connection) string {
	const unescapedMask = "conmMaskedPassword"

	return strings.ReplaceAll(
		cfg.ConnectionString(unescapedMask),
		unescapedMask,
		strings.Repeat(gInputMask, 3),
	)
}

func infoValueRow(label, value string) infoRow {
	if value == "" {
		return infoRow{label: label, value: gEmpty, tone: cFaint}
	}

	return infoRow{label: label, value: value, tone: cFg}
}

func (m infoModel) cursorRow() (infoRow, bool) {
	pick := 0
	for _, r := range m.rows() {
		if r.head {
			continue
		}
		if pick == m.cursor {
			return r, true
		}
		pick++
	}

	return infoRow{}, false
}

func (m infoModel) pickCount() int {
	n := 0
	for _, r := range m.rows() {
		if !r.head {
			n++
		}
	}

	return n
}
