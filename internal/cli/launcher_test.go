package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

			conm := config.Conm{Connections: []config.ConnectionSettings{{Type: tt.kind, CLI: tt.cli}}}

			assert.Equal(t, tt.want, For(conm, tt.kind).Info(t.Context()))
		})
	}
}

func TestLauncherInfoResolvesConfiguredClient(t *testing.T) {
	t.Parallel()

	conm := config.Conm{Connections: []config.ConnectionSettings{{Type: config.PostgresConnType, CLI: Psql}}}

	info := For(conm, config.PostgresConnType).Info(t.Context())

	path, err := exec.LookPath(Psql)

	assert.Equal(t, Psql, info.Name)
	assert.Equal(t, err == nil, info.Installed)
	assert.Equal(t, path, info.Path)
	assert.Equal(t, Version(t.Context(), Psql), info.Version)
}

func TestLauncherInfoFallsBackToInstalledClient(t *testing.T) {
	tests := []struct {
		name      string
		installed []string
		want      Info
	}{
		{name: "nothing installed", want: Info{}},
		{name: "first installed client", installed: []string{Pgcli, USQL}, want: Info{Name: Pgcli, Installed: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, name := range tt.installed {
				require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\n"), 0o700))
			}
			t.Setenv("PATH", dir)

			conm := config.Conm{Connections: []config.ConnectionSettings{{Type: config.PostgresConnType}}}
			got := For(conm, config.PostgresConnType).Info(t.Context())

			if tt.want.Installed {
				tt.want.Path = filepath.Join(dir, tt.want.Name)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
