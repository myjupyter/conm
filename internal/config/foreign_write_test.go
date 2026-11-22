package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteForeign(t *testing.T) {
	path := filepath.Join(t.TempDir(), "made", "file")

	require.NoError(t, WriteForeign(path, []byte("first")))
	require.NoError(t, WriteForeign(path, []byte("second")))

	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "second", string(raw))

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}
