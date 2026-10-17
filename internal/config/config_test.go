package config

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteFileAtomic(t *testing.T) {
	t.Parallel()

	type args struct {
		data []byte
	}
	tests := []struct {
		name     string
		prepare  func(t *testing.T, dir, path string)
		args     args
		want     string
		wantMode fs.FileMode
		wantErr  bool
	}{
		{
			name: "replaces existing content and tightens mode to 0600",
			prepare: func(t *testing.T, _, path string) {
				t.Helper()
				require.NoError(t, os.WriteFile(path, []byte("old"), 0o644))
			},
			args:     args{data: []byte("new")},
			want:     "new",
			wantMode: 0o600,
		},
		{
			name:     "creates missing file",
			prepare:  func(*testing.T, string, string) {},
			args:     args{data: []byte("new")},
			want:     "new",
			wantMode: 0o600,
		},
		{
			name: "writes through a symlink",
			prepare: func(t *testing.T, dir, path string) {
				t.Helper()
				target := filepath.Join(dir, "target.toml")
				require.NoError(t, os.WriteFile(target, []byte("old"), 0o644))
				require.NoError(t, os.Symlink(target, path))
			},
			args:     args{data: []byte("new")},
			want:     "new",
			wantMode: 0o600,
		},
		{
			name: "creates the target of a dangling symlink",
			prepare: func(t *testing.T, _, path string) {
				t.Helper()
				require.NoError(t, os.Symlink("missing.toml", path))
			},
			args:     args{data: []byte("new")},
			want:     "new",
			wantMode: 0o600,
		},
		{
			name: "writes through a chain of symlinks",
			prepare: func(t *testing.T, dir, path string) {
				t.Helper()
				target := filepath.Join(dir, "target.toml")
				require.NoError(t, os.WriteFile(target, []byte("old"), 0o644))
				require.NoError(t, os.Symlink(target, filepath.Join(dir, "middle.toml")))
				require.NoError(t, os.Symlink("middle.toml", path))
			},
			args:     args{data: []byte("new")},
			want:     "new",
			wantMode: 0o600,
		},
		{
			name: "resolves a relative parent link through a symlinked directory",
			prepare: func(t *testing.T, dir, path string) {
				t.Helper()
				store := filepath.Join(dir, "real")
				require.NoError(t, os.MkdirAll(filepath.Join(store, "conm"), 0o700))
				require.NoError(t, os.WriteFile(filepath.Join(store, "shared.toml"), []byte("old"), 0o644))
				require.NoError(t, os.Symlink("../shared.toml", filepath.Join(store, "conm", "target.toml")))
				require.NoError(t, os.Symlink(filepath.Join(store, "conm"), filepath.Join(dir, "conm")))
				require.NoError(t, os.Symlink("conm/target.toml", path))
			},
			args:     args{data: []byte("new")},
			want:     "new",
			wantMode: 0o600,
		},
		{
			name: "symlink loop",
			prepare: func(t *testing.T, _, path string) {
				t.Helper()
				require.NoError(t, os.Symlink("conn.toml", path))
			},
			args:    args{data: []byte("new")},
			wantErr: true,
		},
		{
			name: "temp file can't be created",
			prepare: func(t *testing.T, dir, path string) {
				t.Helper()
				if os.Geteuid() == 0 {
					t.Skip("root ignores directory permissions")
				}
				require.NoError(t, os.WriteFile(path, []byte("old"), 0o600))
				require.NoError(t, os.Chmod(dir, 0o500))
				t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })
			},
			args:    args{data: []byte("new")},
			want:    "old",
			wantErr: true,
		},
		{
			name: "target is a directory",
			prepare: func(t *testing.T, _, path string) {
				t.Helper()
				require.NoError(t, os.MkdirAll(filepath.Join(path, "occupied"), 0o700))
			},
			args:    args{data: []byte("new")},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			path := filepath.Join(dir, "conn.toml")
			tt.prepare(t, dir, path)
			before := dirListing(t, dir)

			err := writeFileAtomic(path, tt.args.data)

			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, before, dirListing(t, dir))
			} else {
				require.NoError(t, err)
				info, statErr := os.Stat(path)
				require.NoError(t, statErr)
				assert.Equal(t, tt.wantMode, info.Mode().Perm())
			}
			if tt.want != "" {
				got, readErr := os.ReadFile(path)
				require.NoError(t, readErr)
				assert.Equal(t, tt.want, string(got))
			}
		})
	}
}

func TestWriteAndClose(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		prepare func(t *testing.T, f *os.File)
		wantErr bool
	}{
		{
			name:    "writes and closes",
			prepare: func(*testing.T, *os.File) {},
		},
		{
			name: "write fails",
			prepare: func(t *testing.T, f *os.File) {
				t.Helper()
				require.NoError(t, f.Close())
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			f, err := os.CreateTemp(t.TempDir(), "conn")
			require.NoError(t, err)
			tt.prepare(t, f)

			err = writeAndClose(f, []byte("new"))

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			got, readErr := os.ReadFile(f.Name())
			require.NoError(t, readErr)
			assert.Equal(t, "new", string(got))
		})
	}
}

func dirListing(t *testing.T, dir string) map[string]fs.FileMode {
	t.Helper()
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	listing := make(map[string]fs.FileMode, len(entries))
	for _, entry := range entries {
		listing[entry.Name()] = entry.Type()
	}
	return listing
}
