package ui

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui/spec"
)

func TestPasteText(t *testing.T) {
	tests := map[string]struct {
		raw  string
		want string
	}{
		"plain":            {raw: "https://grafana.internal/d/pg", want: "https://grafana.internal/d/pg"},
		"trailing newline": {raw: "https://grafana.internal\n", want: "https://grafana.internal"},
		"crlf":             {raw: "https://grafana.internal\r\n", want: "https://grafana.internal"},
		"surrounding ws":   {raw: "  grafana  ", want: "grafana"},
		"multi-line":       {raw: "one\ntwo", want: "one two"},
		"tabs":             {raw: "a\tb", want: "a b"},
		"empty":            {raw: "\n\n", want: ""},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, pasteText(tt.raw))
		})
	}
}

func TestFormPaste(t *testing.T) {
	url := "https://grafana.internal/d/pg"

	t.Run("into a link url", func(t *testing.T) {
		is := assert.New(t)

		m := newLinkForm(t, []config.WebLink{{Name: "grafana"}})
		m.idx = m.linkBase + 1

		m = m.applyPaste(url+"\n", nil)

		is.Equal(url, m.vals[spec.LinkURLKey(0)])
		is.True(m.insert, "pasting into a hovered field starts editing it")
	})

	t.Run("appends while editing", func(t *testing.T) {
		m := newLinkForm(t, []config.WebLink{{Name: "grafana", URL: "https://"}})
		m.idx = m.linkBase + 1
		m.insert = true

		m = m.applyPaste("grafana.internal", nil)

		assert.Equal(t, "https://grafana.internal", m.vals[spec.LinkURLKey(0)])
	})

	t.Run("a selector refuses the paste", func(t *testing.T) {
		is := assert.New(t)

		m := newLinkForm(t, nil)
		m.section = 0
		for i, f := range m.sectionFields(0) {
			if m.isSelector(f) {
				m.idx = i
				break
			}
		}
		was := m.vals[m.fields[m.currentField()].Key]

		m = m.applyPaste("nonsense", nil)

		is.False(m.insert)
		is.Equal(was, m.vals[m.fields[m.currentField()].Key])
		is.Equal(kindWarn, m.statusKind)
	})

	t.Run("a failed read is reported", func(t *testing.T) {
		m := newLinkForm(t, []config.WebLink{{Name: "grafana"}})
		m.idx = m.linkBase + 1

		m = m.applyPaste("", errors.New("no clipboard reader"))

		assert.Equal(t, kindErr, m.statusKind)
		assert.Empty(t, m.vals[spec.LinkURLKey(0)])
	})

	t.Run("an empty clipboard changes nothing", func(t *testing.T) {
		m := newLinkForm(t, []config.WebLink{{Name: "grafana"}})
		m.idx = m.linkBase + 1

		m = m.applyPaste("  \n ", nil)

		assert.Equal(t, kindWarn, m.statusKind)
		assert.Empty(t, m.vals[spec.LinkURLKey(0)])
	})
}
