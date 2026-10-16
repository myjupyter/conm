package config

import (
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
		name    string
		prepare func(t *testing.T, dir, path string)
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "replaces existing content",
			prepare: func(t *testing.T, _, path string) {
				t.Helper()
				require.NoError(t, os.WriteFile(path, []byte("old"), 0o644))
			},
			args: args{data: []byte("new")},
			want: "new",
		},
		{
			name:    "creates missing file",
			prepare: func(*testing.T, string, string) {},
			args:    args{data: []byte("new")},
			want:    "new",
		},
		{
			name: "writes through a symlink",
			prepare: func(t *testing.T, dir, path string) {
				t.Helper()
				target := filepath.Join(dir, "target.toml")
				require.NoError(t, os.WriteFile(target, []byte("old"), 0o600))
				require.NoError(t, os.Symlink(target, path))
			},
			args: args{data: []byte("new")},
			want: "new",
		},
		{
			name: "temp file can't be created",
			prepare: func(t *testing.T, dir, path string) {
				t.Helper()
				require.NoError(t, os.WriteFile(path, []byte("old"), 0o600))
				require.NoError(t, os.Chmod(dir, 0o500))
				t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })
			},
			args:    args{data: []byte("new")},
			want:    "old",
			wantErr: true,
		},
		{
			name: "rename fails",
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
			before, err := os.ReadDir(dir)
			require.NoError(t, err)

			err = writeFileAtomic(path, tt.args.data)

			after, readErr := os.ReadDir(dir)
			require.NoError(t, readErr)
			if tt.wantErr {
				require.Error(t, err)
				assert.Len(t, after, len(before))
			} else {
				require.NoError(t, err)
				info, statErr := os.Stat(path)
				require.NoError(t, statErr)
				assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
			}
			if tt.want != "" {
				got, readErr := os.ReadFile(path)
				require.NoError(t, readErr)
				assert.Equal(t, tt.want, string(got))
			}
		})
	}
}
