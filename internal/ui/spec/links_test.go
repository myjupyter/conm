package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/myjupyter/conm/internal/config"
)

func TestLinksRoundTrip(t *testing.T) {
	tests := map[string]struct {
		links []config.WebLink
		want  []config.WebLink
	}{
		"none": {links: nil, want: nil},
		"one": {
			links: []config.WebLink{{Name: "grafana", URL: "https://grafana.internal/d/db"}},
			want:  []config.WebLink{{Name: "grafana", URL: "https://grafana.internal/d/db"}},
		},
		"order is kept": {
			links: []config.WebLink{
				{Name: "grafana", URL: "https://grafana.internal"},
				{Name: "logs", URL: "https://kibana.internal"},
				{Name: "runbook", URL: "https://wiki.internal"},
			},
			want: []config.WebLink{
				{Name: "grafana", URL: "https://grafana.internal"},
				{Name: "logs", URL: "https://kibana.internal"},
				{Name: "runbook", URL: "https://wiki.internal"},
			},
		},
		"whitespace is trimmed": {
			links: []config.WebLink{{Name: "  grafana  ", URL: " https://grafana.internal "}},
			want:  []config.WebLink{{Name: "grafana", URL: "https://grafana.internal"}},
		},
		"an untouched row is not a link": {
			links: []config.WebLink{
				{Name: "grafana", URL: "https://grafana.internal"},
				{Name: "", URL: ""},
			},
			want: []config.WebLink{{Name: "grafana", URL: "https://grafana.internal"}},
		},
		"a half-written row is kept for the refusal": {
			links: []config.WebLink{{Name: "grafana", URL: ""}},
			want:  []config.WebLink{{Name: "grafana", URL: ""}},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			values := map[FormFieldKey]FormFieldValue{}
			SeedLinks(tt.links, values)

			assert.Equal(t, tt.want, LinksFrom(values))
		})
	}
}

func TestLinkKeys(t *testing.T) {
	is := assert.New(t)

	is.True(IsLinkKey(LinkNameKey(0)))
	is.True(IsLinkKey(LinkURLKey(7)))
	is.False(IsLinkKey("name"))
	is.False(IsLinkKey("link.notes"))

	i, ok := LinkIndexOf(LinkURLKey(3))
	is.True(ok)
	is.Equal(3, i)

	_, ok = LinkIndexOf("description")
	is.False(ok)

	is.Equal(0, LinkCount(map[FormFieldKey]FormFieldValue{"name": "x"}))
	is.Equal(2, LinkCount(map[FormFieldKey]FormFieldValue{
		"name":         "x",
		LinkNameKey(0): "grafana",
		LinkURLKey(0):  "https://grafana.internal",
		LinkNameKey(1): "",
		LinkURLKey(1):  "https://logs.internal",
	}))
	is.Equal(1, LinkCount(map[FormFieldKey]FormFieldValue{
		LinkNameKey(0): "grafana",
		LinkURLKey(2):  "https://logs.internal",
	}), "the run ends at the first index without a name")
}

func TestValidateLink(t *testing.T) {
	tests := map[string]struct {
		name, url         string
		wantName, wantURL bool
	}{
		"whole":            {name: "grafana", url: "https://grafana.internal"},
		"untouched row":    {name: "", url: ""},
		"name without url": {name: "grafana", url: "", wantURL: true},
		"url without name": {name: "", url: "https://grafana.internal", wantName: true},
		"bad url":          {name: "grafana", url: "grafana.internal", wantURL: true},
		"both bad":         {name: " ", url: "nope", wantName: true, wantURL: true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			nameErr, urlErr := ValidateLink(tt.name, tt.url)

			assert.Equal(t, tt.wantName, nameErr != nil)
			assert.Equal(t, tt.wantURL, urlErr != nil)
		})
	}
}

func TestBuildMetaCarriesLinks(t *testing.T) {
	values := map[FormFieldKey]FormFieldValue{
		"name": "prod-primary",
		"tags": "prod, critical",
	}
	SeedLinks([]config.WebLink{{Name: "grafana", URL: "https://grafana.internal"}}, values)

	meta := buildMeta(values, "name", "description", "tags")

	assert.Equal(t, []config.WebLink{{Name: "grafana", URL: "https://grafana.internal"}}, meta.Links)
	assert.Equal(t, "prod-primary", meta.Name)
}
