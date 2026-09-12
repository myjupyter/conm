package ui

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/network"
	"github.com/myjupyter/conm/internal/ui/view"
)

type stubConnections struct {
	confs []config.Connection
}

func (stubConnections) Kind() config.ConnType                       { return config.PostgresConnType }
func (stubConnections) ClientInfo(context.Context) cli.Info         { return cli.Info{} }
func (r stubConnections) Len() int                                  { return len(r.confs) }
func (stubConnections) ConnectionAt(int) (network.Connection, bool) { return nil, false }
func (stubConnections) Add(config.Connection) error                 { return nil }
func (stubConnections) Edit(int, config.Connection) error           { return nil }
func (stubConnections) Remove(int) error                            { return nil }

func (r stubConnections) ConfigAt(i int) (config.Connection, bool) {
	if i < 0 || i >= len(r.confs) {
		return nil, false
	}
	return r.confs[i], true
}

const stubPassword = "hunter2"

func newInfoFixture(t *testing.T, links []config.WebLink) infoModel {
	t.Helper()

	cfg := config.Postgres{
		Metadata:   config.ConnMeta{Name: "prod-primary", Description: "orders", Links: links},
		Hostname:   "db-prod-01.internal",
		PortNumber: 5432,
		User:       "svc_api",
		Password:   "keyring:" + stubPassword,
		DBName:     "orders",
		SSLMode:    config.PostgresSSLModePrefer,
	}

	conns := view.NewConnections(stubConnections{confs: []config.Connection{cfg}})
	require.Equal(t, 1, conns.Len())

	return newInfoModel(conns)
}

func linkFixture(n int) []config.WebLink {
	links := make([]config.WebLink, 0, n)
	for i := range n {
		links = append(links, config.WebLink{
			Name: fmt.Sprintf("link-%d", i),
			URL:  fmt.Sprintf("https://example.test/%d", i),
		})
	}
	return links
}

func TestInfoCursorIsLazyAndWraps(t *testing.T) {
	is, must := assert.New(t), require.New(t)

	m := newInfoFixture(t, nil)
	_, ok := m.cursorRow()
	must.False(ok, "the sheet opens with no cursor")

	n := m.pickCount()
	must.Equal(6, n, "three meta fields and three cli fields, no heads")

	first, ok := m.move(1).cursorRow()
	must.True(ok)
	is.Equal("name", first.label, "j creates the cursor on the first field")

	last, ok := m.move(-1).cursorRow()
	must.True(ok)
	is.Equal(infoPathLabel, last.label, "k creates it on the last field")

	is.Equal(0, m.move(-1).move(1).cursor, "the cursor wraps past the end")
	is.Equal(n-1, m.move(1).move(-1).cursor, "and past the start")
}

func TestInfoCursorSkipsHeads(t *testing.T) {
	m := newInfoFixture(t, nil)

	for i := range m.pickCount() {
		row, ok := m.cursorRowAt(i)
		require.True(t, ok)
		assert.False(t, row.head, "the cursor never lands on a section head")
		assert.NotEmpty(t, row.label)
	}
}

func TestInfoNeverShowsSecretMaterial(t *testing.T) {
	is := assert.New(t)

	m := newInfoFixture(t, linkFixture(2))
	for tab := range infoTabCount {
		m.tab = tab
		for _, r := range m.rows() {
			is.NotContains(r.value, stubPassword, "no password material on the %s page", infoTabTitles[tab])
			is.NotContains(strings.ToLower(r.label), "password")
			is.NotContains(strings.ToLower(r.label), "provider")
		}
	}

	m.tab = infoEndpoint
	dsn, ok := m.cursorRowAt(m.pickCount() - 1)
	is.True(ok)
	is.Equal(infoDSNLabel, dsn.label)
	is.Contains(dsn.value, strings.Repeat(gInputMask, 3), "the dsn wears a mask where the password goes")
}

func TestInfoLinksScrollWithTheCursor(t *testing.T) {
	is, must := assert.New(t), require.New(t)

	m := newInfoFixture(t, linkFixture(12))
	m.tab = infoLinks
	must.Equal(12, m.pickCount())

	shown := func(m infoModel) []string {
		var out []string
		for _, r := range m.rows() {
			if !r.head && !r.hidden {
				out = append(out, r.label)
			}
		}
		return out
	}

	is.Len(shown(m), infoLinkWindow, "the page renders at most eight links")
	is.Equal("link-0", shown(m)[0])
	is.Equal(gScrollDown+" 4 below", m.rows()[0].note, "only what is hidden is counted")

	for range infoLinkWindow + 1 {
		m = m.move(1)
	}
	is.Equal("link-1", shown(m)[0], "stepping past the last visible link shifts the window by one")
	is.Equal(gScrollUp+" 1 above · "+gScrollDown+" 3 below", m.rows()[0].note)

	m = m.switchTab(1).switchTab(-1)
	is.Equal(-1, m.cursor, "switching page drops the cursor")
	is.Equal("link-0", shown(m)[0], "and resets the window")
}

func (m infoModel) cursorRowAt(pick int) (infoRow, bool) {
	m.cursor = pick
	return m.cursorRow()
}

func TestInfoHidesTheLinksTabWithoutLinks(t *testing.T) {
	is := assert.New(t)

	m := newInfoFixture(t, nil)
	is.Equal([]string{"general", "endpoint"}, m.tabTitles(), "an unlinked connection has no links page")
	is.Equal(infoEndpoint, m.switchTab(1).tab)
	is.Equal(infoGeneral, m.switchTab(1).switchTab(1).tab, "the pages cycle without ever landing on links")
	is.Equal(infoEndpoint, m.switchTab(-1).tab)

	linked := newInfoFixture(t, linkFixture(1))
	is.Equal([]string{"general", "endpoint", "links"}, linked.tabTitles())
	is.Equal(infoLinks, linked.switchTab(-1).tab, "one link is enough to bring the page back")
}
