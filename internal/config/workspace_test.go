package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUseWorkspace(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name           string
		args           args
		wantErr        bool
		wantConmPath   string
		wantSecretPath string
	}{
		{
			name:           "default workspace is a folder like any other",
			args:           args{name: DefaultWorkspace},
			wantConmPath:   filepath.Join(xdg.ConfigHome, "conm", "default", "conm.toml"),
			wantSecretPath: filepath.Join(xdg.DataHome, "conm", "default", "secret.toml"),
		},
		{
			name:           "named workspace gets its own folder and secret workspace",
			args:           args{name: "corp"},
			wantConmPath:   filepath.Join(xdg.ConfigHome, "conm", "corp", "conm.toml"),
			wantSecretPath: filepath.Join(xdg.DataHome, "conm", "corp", "secret.toml"),
		},
		{
			name:           "any folder-legal name",
			args:           args{name: "My Servers"},
			wantConmPath:   filepath.Join(xdg.ConfigHome, "conm", "My Servers", "conm.toml"),
			wantSecretPath: filepath.Join(xdg.DataHome, "conm", "My Servers", "secret.toml"),
		},
		{name: "empty name", args: args{name: ""}, wantErr: true},
		{name: "path traversal", args: args{name: "../x"}, wantErr: true},
		{name: "parent dir", args: args{name: ".."}, wantErr: true},
		{name: "backslash", args: args{name: `a\b`}, wantErr: true},
		{name: "control character", args: args{name: "a\tb"}, wantErr: true},
		{name: "windows reserved name", args: args{name: "aux"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() { workspace = DefaultWorkspace })

			err := UseWorkspace(tt.args.name)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, DefaultWorkspace, Workspace())
				return
			}
			require.NoError(t, err)

			assert.Equal(t, tt.args.name, Workspace())
			assert.Equal(t, tt.wantConmPath, ConmPath())
			assert.Equal(t, tt.wantSecretPath, SecretConfigPath())
		})
	}
}

func TestSecretWorkspaceOf(t *testing.T) {
	type args struct {
		secretFile string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    string
	}{
		{name: "a missing id is generated", args: args{}},
		{name: "a missing id in an existing file is generated", args: args{secretFile: "[[keyring]]\nid = \"db\"\n"}},
		{
			name: "a stored id is kept",
			args: args{secretFile: "workspace_id = \"6f1c2a9e-4b7d-4e2a-9c3f-1a2b3c4d5e6f\"\n"},
			want: "conm.6f1c2a9e-4b7d-4e2a-9c3f-1a2b3c4d5e6f",
		},
		{name: "a hand-edited id is refused", args: args{secretFile: "workspace_id = \"corp\"\n"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			t.Setenv("XDG_DATA_HOME", t.TempDir())
			xdg.Reload()
			t.Cleanup(xdg.Reload)

			if tt.args.secretFile != "" {
				path := SecretConfigPathOf("corp")
				require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
				require.NoError(t, os.WriteFile(path, []byte(tt.args.secretFile), 0o600))
			}

			got, err := SecretWorkspaceOf("corp")
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			is := assert.New(t)
			if tt.want != "" {
				is.Equal(tt.want, got)
			}
			is.Regexp(`^conm\.[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, got)

			again, err := SecretWorkspaceOf("corp")
			require.NoError(t, err)
			is.Equal(got, again)

			require.NoError(t, RenameWorkspace("corp", "work"))
			renamed, err := SecretWorkspaceOf("work")
			require.NoError(t, err)
			is.Equal(got, renamed)
		})
	}
}

func TestSetStartupWorkspace(t *testing.T) {
	type args struct {
		existing []string
		name     string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    string
	}{
		{name: "nothing set opens the default", args: args{}, want: DefaultWorkspace},
		{name: "an existing workspace", args: args{existing: []string{"corp"}, name: "corp"}, want: "corp"},
		{name: "back to the default", args: args{existing: []string{"corp"}, name: DefaultWorkspace}, want: DefaultWorkspace},
		{name: "an unknown workspace is refused", args: args{name: "copr"}, wantErr: true, want: DefaultWorkspace},
		{name: "a name that is not a folder name", args: args{name: "a/b"}, wantErr: true, want: DefaultWorkspace},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			t.Setenv("XDG_DATA_HOME", t.TempDir())
			xdg.Reload()
			t.Cleanup(xdg.Reload)

			for _, ws := range tt.args.existing {
				require.NoError(t, CreateWorkspace(ws))
			}

			if tt.args.name != "" {
				err := SetStartupWorkspace(tt.args.name)
				if tt.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			}

			got, err := StartupWorkspace()
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStartupWorkspace_afterRenameAndRemove(t *testing.T) {
	type args struct {
		existing []string
		startup  string
		change   func() error
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "renaming the startup workspace moves startup along",
			args: args{existing: []string{"corp"}, startup: "corp", change: func() error { return RenameWorkspace("corp", "work") }},
			want: "work",
		},
		{
			name: "renaming the default while nothing is set moves startup along",
			args: args{change: func() error { return RenameWorkspace(DefaultWorkspace, "home") }},
			want: "home",
		},
		{
			name: "renaming another workspace leaves startup alone",
			args: args{existing: []string{"corp", "lab"}, startup: "corp", change: func() error { return RenameWorkspace("lab", "test") }},
			want: "corp",
		},
		{
			name: "removing the startup workspace falls back to the default",
			args: args{existing: []string{"corp"}, startup: "corp", change: func() error { return RemoveWorkspace("corp") }},
			want: DefaultWorkspace,
		},
		{
			name: "removing another workspace leaves startup alone",
			args: args{existing: []string{"corp", "lab"}, startup: "corp", change: func() error { return RemoveWorkspace("lab") }},
			want: "corp",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			t.Setenv("XDG_DATA_HOME", t.TempDir())
			xdg.Reload()
			t.Cleanup(xdg.Reload)

			for _, ws := range tt.args.existing {
				require.NoError(t, CreateWorkspace(ws))
			}
			if tt.args.startup != "" {
				require.NoError(t, SetStartupWorkspace(tt.args.startup))
			}

			require.NoError(t, tt.args.change())

			got, err := StartupWorkspace()
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
