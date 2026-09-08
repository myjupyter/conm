package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui/spec"
)

func newLinkForm(t *testing.T, links []config.WebLink) formModel {
	t.Helper()

	initial := spec.PostgresFormSpec.SeedFunc(config.Postgres{
		Metadata:   config.ConnMeta{Name: "prod-primary"},
		Hostname:   "db-prod-01.internal",
		PortNumber: 5432,
		User:       "svc_api",
		Password:   "hunter2",
		DBName:     "orders",
		SSLMode:    config.PostgresSSLModePrefer,
	})
	require.NotNil(t, initial)
	spec.SeedLinks(links, initial)

	m := newFormModel(config.PostgresConnType, spec.PostgresFormSpec, "edit", initial, true, nil)
	require.True(t, m.hasLinks())
	m.section = m.linkSect

	return m
}

func TestFormSeedsLinks(t *testing.T) {
	tests := map[string]int{"none": 0, "one": 1, "three": 3}

	for name, n := range tests {
		t.Run(name, func(t *testing.T) {
			links := make([]config.WebLink, n)
			for i := range links {
				links[i] = config.WebLink{Name: "grafana", URL: "https://grafana.internal"}
			}

			m := newLinkForm(t, links)

			assert.Equal(t, n, m.linkCount())
			assert.Len(t, m.sectionFields(m.linkSect), m.linkBase+n*2)
		})
	}
}

func TestFormAddLink(t *testing.T) {
	is, must := assert.New(t), require.New(t)

	m := newLinkForm(t, []config.WebLink{{Name: "grafana", URL: "https://grafana.internal"}})
	before := m.linkCount()
	host := m.vals["host"]

	m.addLink()

	must.Equal(before+1, m.linkCount())
	is.True(m.insert, "a new link is typed straight away")
	is.Equal(spec.LinkNameKey(before), m.fields[m.currentField()].Key, "the cursor lands on the new name")
	is.Empty(m.vals[spec.LinkNameKey(before)])
	is.Empty(m.vals[spec.LinkURLKey(before)])
	is.Equal(host, m.vals["host"], "no other field is touched")
}

func TestFormRemoveLink(t *testing.T) {
	links := []config.WebLink{
		{Name: "grafana", URL: "https://grafana.internal"},
		{Name: "logs", URL: "https://kibana.internal"},
		{Name: "runbook", URL: "https://wiki.internal"},
	}

	tests := map[string]struct {
		at   int
		want []config.WebLink
	}{
		"first":  {at: 0, want: links[1:]},
		"middle": {at: 1, want: []config.WebLink{links[0], links[2]}},
		"last":   {at: 2, want: links[:2]},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			is := assert.New(t)

			m := newLinkForm(t, links)
			host := m.vals["host"]

			m.removeLink(tt.at)

			is.Equal(len(tt.want), m.linkCount())
			is.Equal(tt.want, spec.LinksFrom(m.values()), "the survivors are renumbered without a gap")
			is.False(m.insert)
			is.Less(m.idx, len(m.sectionFields(m.linkSect)), "the cursor stays on a field that exists")
			is.Equal(host, m.vals["host"], "no other field is touched")

			gone := len(links) - 1
			is.NotContains(m.vals, spec.LinkNameKey(gone), "the trailing keys are dropped")
			is.NotContains(m.vals, spec.LinkURLKey(gone))
		})
	}
}

func TestFormRemoveLastLink(t *testing.T) {
	m := newLinkForm(t, []config.WebLink{{Name: "grafana", URL: "https://grafana.internal"}})
	m.idx = len(m.sectionFields(m.linkSect)) - 1

	m.removeLink(0)

	assert.Zero(t, m.linkCount())
	assert.Nil(t, spec.LinksFrom(m.values()))
	assert.Less(t, m.idx, len(m.sectionFields(m.linkSect)))
}

func TestFormLinkErrors(t *testing.T) {
	tests := map[string]struct {
		link            config.WebLink
		nameErr, urlErr bool
	}{
		"whole":            {link: config.WebLink{Name: "grafana", URL: "https://grafana.internal"}},
		"untouched row":    {link: config.WebLink{}},
		"name without url": {link: config.WebLink{Name: "grafana"}, urlErr: true},
		"url without name": {link: config.WebLink{URL: "https://grafana.internal"}, nameErr: true},
		"not a url":        {link: config.WebLink{Name: "grafana", URL: "grafana.internal"}, urlErr: true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			is := assert.New(t)

			m := newLinkForm(t, []config.WebLink{tt.link})

			is.Equal(tt.nameErr, m.linkError(spec.LinkNameKey(0)) != "")
			is.Equal(tt.urlErr, m.linkError(spec.LinkURLKey(0)) != "")

			// A refused link keeps the form open on the section it came from.
			model, _ := m.submit()
			refused, ok := model.(formModel)
			require.True(t, ok)
			is.Equal(tt.nameErr || tt.urlErr, !refused.submitted)
			if tt.nameErr || tt.urlErr {
				is.Equal(m.linkSect, refused.section)
				is.NotEmpty(refused.status)
			}
		})
	}
}
