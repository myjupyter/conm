package config

import (
	"path/filepath"
	"strings"
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
		wantKeyring    string
	}{
		{
			name:           "default workspace is a folder like any other",
			args:           args{name: DefaultWorkspace},
			wantConmPath:   filepath.Join(xdg.ConfigHome, "conm", "default", "conm.toml"),
			wantSecretPath: filepath.Join(xdg.DataHome, "conm", "default", "secret.toml"),
			wantKeyring:    "conm.default",
		},
		{
			name:           "named workspace gets its own folder and keyring service",
			args:           args{name: "corp"},
			wantConmPath:   filepath.Join(xdg.ConfigHome, "conm", "corp", "conm.toml"),
			wantSecretPath: filepath.Join(xdg.DataHome, "conm", "corp", "secret.toml"),
			wantKeyring:    "conm.corp",
		},
		{
			name:           "any folder-legal name",
			args:           args{name: "My Servers"},
			wantConmPath:   filepath.Join(xdg.ConfigHome, "conm", "My Servers", "conm.toml"),
			wantSecretPath: filepath.Join(xdg.DataHome, "conm", "My Servers", "secret.toml"),
			wantKeyring:    "conm.My Servers",
		},
		{name: "empty name", args: args{name: ""}, wantErr: true},
		{name: "path traversal", args: args{name: "../x"}, wantErr: true},
		{name: "parent dir", args: args{name: ".."}, wantErr: true},
		{name: "backslash", args: args{name: `a\b`}, wantErr: true},
		{name: "control character", args: args{name: "a\tb"}, wantErr: true},
		{name: "too long", args: args{name: strings.Repeat("a", workspaceNameMaxLen+1)}, wantErr: true},
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
			assert.Equal(t, tt.wantKeyring, KeyringService())
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
