package view

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/network"
)

type stubConnections struct {
	kind  config.ConnType
	confs []config.Connection
}

func (r stubConnections) Kind() config.ConnType                     { return r.kind }
func (stubConnections) ClientInfo(context.Context) cli.Info         { return cli.Info{} }
func (r stubConnections) Len() int                                  { return len(r.confs) }
func (stubConnections) IndexOf(string) (int, bool)                  { return -1, false }
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

func newKindsFixture(filled ...config.ConnType) *Connections {
	kinds := []config.ConnType{config.PostgresConnType, config.MySQLConnType, config.RedisConnType}
	repos := make([]stubConnections, 0, len(kinds))
	for _, kind := range kinds {
		repos = append(repos, stubConnections{kind: kind})
	}
	for _, kind := range filled {
		for i := range repos {
			if repos[i].kind == kind {
				repos[i].confs = append(repos[i].confs, config.Postgres{Metadata: config.ConnMeta{Name: kind.String()}})
			}
		}
	}

	return NewConnections(repos[0], repos[1], repos[2])
}

func TestConnections_Kinds(t *testing.T) {
	tests := []struct {
		name       string
		filled     []config.ConnType
		active     config.ConnType
		wantActive config.ConnType
		want       []config.ConnType
	}{
		{
			name:       "opens on the first filled kind",
			filled:     []config.ConnType{config.MySQLConnType},
			wantActive: config.MySQLConnType,
			want:       []config.ConnType{config.MySQLConnType},
		},
		{
			name:       "the active kind stays while empty",
			filled:     []config.ConnType{config.MySQLConnType},
			active:     config.RedisConnType,
			wantActive: config.RedisConnType,
			want:       []config.ConnType{config.MySQLConnType, config.RedisConnType},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := newKindsFixture(tt.filled...)
			v.SetActive(tt.active)

			assert.Equal(t, tt.wantActive, v.Active())
			assert.Equal(t, tt.want, v.Kinds())
		})
	}
}

func TestConnections_NextKind(t *testing.T) {
	tests := []struct {
		name   string
		filled []config.ConnType
		want   config.ConnType
		wantOK bool
	}{
		{
			name:   "steps over an empty kind",
			filled: []config.ConnType{config.PostgresConnType, config.RedisConnType},
			want:   config.RedisConnType,
			wantOK: true,
		},
		{
			name:   "a single filled kind cannot step",
			filled: []config.ConnType{config.PostgresConnType},
			want:   config.PostgresConnType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := newKindsFixture(tt.filled...)

			gotOK := v.NextKind()
			assert.Equal(t, tt.wantOK, gotOK)
			assert.Equal(t, tt.want, v.Active())
		})
	}
}
