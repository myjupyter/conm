package cli

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/myjupyter/conm/internal/config"
)

func TestLauncherInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		kind config.ConnType
		cli  string
		want Info
	}{
		{
			name: "no client configured",
			kind: config.PostgresConnType,
			want: Info{},
		},
		{
			name: "client of another database",
			kind: config.PostgresConnType,
			cli:  RedisCLI,
			want: Info{Name: RedisCLI},
		},
		{
			name: "unknown client",
			kind: config.PostgresConnType,
			cli:  "sh",
			want: Info{Name: "sh"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			conm := config.Conm{Databases: []config.Database{{Type: tt.kind, CLI: tt.cli}}}

			assert.Equal(t, tt.want, For(conm, tt.kind).Info(t.Context()))
		})
	}
}

func TestLauncherInfoResolvesConfiguredClient(t *testing.T) {
	t.Parallel()

	conm := config.Conm{Databases: []config.Database{{Type: config.PostgresConnType, CLI: Psql}}}

	info := For(conm, config.PostgresConnType).Info(t.Context())

	path, err := exec.LookPath(Psql)

	assert.Equal(t, Psql, info.Name)
	assert.Equal(t, err == nil, info.Installed)
	assert.Equal(t, path, info.Path)
	assert.Equal(t, Version(t.Context(), Psql), info.Version)
}
