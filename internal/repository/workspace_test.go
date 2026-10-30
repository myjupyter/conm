package repository

import (
	"context"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

func testWorkspaces(t *testing.T) *WorkspaceRepository {
	t.Helper()

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	xdg.Reload()
	keyring.MockInit()
	require.NoError(t, config.UseWorkspace(config.DefaultWorkspace))
	t.Cleanup(xdg.Reload)

	repo, err := NewWorkspaceRepository()
	require.NoError(t, err)

	return repo
}

func withSecret(t *testing.T, ws, account, password string) {
	t.Helper()

	file, err := config.OpenConfig[*config.KeyringConfigWrapper](config.SecretConfigPathOf(ws))
	require.NoError(t, err)
	file.Add(config.Keyring{SecretID: account, ProviderID: secret.Keyring, LocationRef: account})
	require.NoError(t, file.Save())
	require.NoError(t, keyring.Set(config.KeyringServiceOf(ws), account, password))
}

func indexOfWorkspace(t *testing.T, repo *WorkspaceRepository, name string) int {
	t.Helper()

	for i := range repo.Len() {
		if got, _ := repo.Get(i); got == name {
			return i
		}
	}
	require.Failf(t, "workspace not listed", "%q", name)
	return -1
}

func TestWorkspaceRepository(t *testing.T) {
	type args struct {
		act func(*testing.T, *WorkspaceRepository) error
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantNames  []string
		wantActive string
		check      func(*testing.T)
	}{
		{
			name: "add lists the new workspace next to the current one",
			args: args{act: func(t *testing.T, r *WorkspaceRepository) error {
				t.Helper()
				return r.Add("corp")
			}},
			wantNames:  []string{"corp", "default"},
			wantActive: "default",
		},
		{
			name: "add refuses an existing name",
			args: args{act: func(t *testing.T, r *WorkspaceRepository) error {
				t.Helper()
				require.NoError(t, r.Add("corp"))
				return r.Add("corp")
			}},
			wantErr:    true,
			wantNames:  []string{"corp", "default"},
			wantActive: "default",
		},
		{
			name: "add refuses the default workspace even before its folder exists",
			args: args{act: func(t *testing.T, r *WorkspaceRepository) error {
				t.Helper()
				return r.Add(config.DefaultWorkspace)
			}},
			wantErr:    true,
			wantNames:  []string{"default"},
			wantActive: "default",
		},
		{
			name: "add refuses a name that is not a folder name",
			args: args{act: func(t *testing.T, r *WorkspaceRepository) error {
				t.Helper()
				return r.Add("a/b")
			}},
			wantErr:    true,
			wantNames:  []string{"default"},
			wantActive: "default",
		},
		{
			name: "rename moves the keychain entries to the new name",
			args: args{act: func(t *testing.T, r *WorkspaceRepository) error {
				t.Helper()
				require.NoError(t, r.Add("corp"))
				withSecret(t, "corp", "db", "s3cret")
				return r.Rename(context.Background(), indexOfWorkspace(t, r, "corp"), "work")
			}},
			wantNames:  []string{"default", "work"},
			wantActive: "default",
			check: func(t *testing.T) {
				t.Helper()
				got, err := keyring.Get(config.KeyringServiceOf("work"), "db")
				require.NoError(t, err)
				assert.Equal(t, "s3cret", got)

				_, err = keyring.Get(config.KeyringServiceOf("corp"), "db")
				assert.ErrorIs(t, err, keyring.ErrNotFound)
			},
		},
		{
			name: "renaming the current workspace makes the new name current",
			args: args{act: func(t *testing.T, r *WorkspaceRepository) error {
				t.Helper()
				return r.Rename(context.Background(), indexOfWorkspace(t, r, "default"), "home")
			}},
			wantNames:  []string{"default", "home"},
			wantActive: "home",
		},
		{
			name: "remove deletes the workspace and its keychain entries",
			args: args{act: func(t *testing.T, r *WorkspaceRepository) error {
				t.Helper()
				require.NoError(t, r.Add("corp"))
				withSecret(t, "corp", "db", "s3cret")
				return r.Remove(context.Background(), indexOfWorkspace(t, r, "corp"))
			}},
			wantNames:  []string{"default"},
			wantActive: "default",
			check: func(t *testing.T) {
				t.Helper()
				_, err := keyring.Get(config.KeyringServiceOf("corp"), "db")
				assert.ErrorIs(t, err, keyring.ErrNotFound)
			},
		},
		{
			name: "remove refuses the workspace conm is running in",
			args: args{act: func(t *testing.T, r *WorkspaceRepository) error {
				t.Helper()
				return r.Remove(context.Background(), indexOfWorkspace(t, r, "default"))
			}},
			wantErr:    true,
			wantNames:  []string{"default"},
			wantActive: "default",
		},
		{
			name: "use picks the workspace to switch to",
			args: args{act: func(t *testing.T, r *WorkspaceRepository) error {
				t.Helper()
				require.NoError(t, r.Add("corp"))
				return r.Use(indexOfWorkspace(t, r, "corp"))
			}},
			wantNames:  []string{"corp", "default"},
			wantActive: "corp",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := testWorkspaces(t)

			err := tt.args.act(t, r)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			names := make([]string, 0, r.Len())
			for i := range r.Len() {
				name, _ := r.Get(i)
				names = append(names, name)
			}
			assert.Equal(t, tt.wantNames, names)
			assert.Equal(t, tt.wantActive, r.Active())
			if tt.check != nil {
				tt.check(t)
			}
		})
	}
}
