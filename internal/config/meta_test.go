package config

import (
	"strings"
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateWebLink(t *testing.T) {
	tests := map[string]struct {
		link WebLink
		errs int
	}{
		"valid https":      {link: WebLink{Name: "grafana", URL: "https://grafana.internal/d/db"}, errs: 0},
		"valid http":       {link: WebLink{Name: "logs", URL: "http://kibana.internal/app"}, errs: 0},
		"empty name":       {link: WebLink{Name: "", URL: "https://grafana.internal"}, errs: 1},
		"blank name":       {link: WebLink{Name: "   ", URL: "https://grafana.internal"}, errs: 1},
		"empty url":        {link: WebLink{Name: "grafana", URL: ""}, errs: 1},
		"relative url":     {link: WebLink{Name: "grafana", URL: "/d/db"}, errs: 1},
		"no scheme":        {link: WebLink{Name: "grafana", URL: "grafana.internal/d/db"}, errs: 1},
		"unwanted scheme":  {link: WebLink{Name: "runbook", URL: "ftp://files.internal/rb"}, errs: 1},
		"no host":          {link: WebLink{Name: "runbook", URL: "https:///d/db"}, errs: 1},
		"nothing at all":   {link: WebLink{}, errs: 2},
		"name and no host": {link: WebLink{Name: "", URL: "https://"}, errs: 2},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			errs := 0
			for _, err := range []error{
				ValidateWebLinkName(tt.link.Name),
				ValidateWebLinkURL(tt.link.URL),
			} {
				if err != nil {
					errs++
				}
			}
			assert.Equal(t, tt.errs, errs)
		})
	}
}

func TestConnMetaLinksRoundTrip(t *testing.T) {
	tests := map[string]struct {
		links   []WebLink
		written bool
	}{
		"none": {links: nil, written: false},
		"one":  {links: []WebLink{{Name: "grafana", URL: "https://grafana.internal/d/db"}}, written: true},
		"many": {
			links: []WebLink{
				{Name: "grafana", URL: "https://grafana.internal/d/db"},
				{Name: "logs", URL: "https://kibana.internal/app/logs"},
				{Name: "runbook", URL: "https://wiki.internal/runbooks/orders"},
			},
			written: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			is, must := assert.New(t), require.New(t)

			pg := Postgres{
				Metadata:   ConnMeta{Name: "prod-primary", Links: tt.links},
				Hostname:   "db-prod-01.internal",
				PortNumber: 5432,
				User:       "svc_api",
				Password:   "hunter2",
				DBName:     "orders",
			}

			raw, err := toml.Marshal(pg)
			must.NoError(err)
			is.Equal(tt.written, strings.Contains(string(raw), "links"))

			var read Postgres
			must.NoError(toml.Unmarshal(raw, &read))
			is.Equal(tt.links, read.Meta().Links)

			// A link is the user's own note, never something the server is told.
			for _, l := range tt.links {
				is.NotContains(pg.ConnectionString("hunter2"), l.URL)
			}
		})
	}
}
