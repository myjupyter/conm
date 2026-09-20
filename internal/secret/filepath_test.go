package secret

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type rawRef string

func (r rawRef) SecretRef() string { return string(r) }

func TestFilepathResolve(t *testing.T) {
	existing := filepath.Join(t.TempDir(), "id_ed25519")
	require.NoError(t, os.WriteFile(existing, nil, 0o600))

	p := &FilepathProvider{}

	got, err := p.Resolve(context.Background(), rawRef(Ref(Filepath, existing)))
	require.NoError(t, err)
	assert.Equal(t, existing, got)

	_, err = p.Resolve(context.Background(), rawRef(Ref(Filepath, existing+".missing")))
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestExpandHome(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	is := assert.New(t)
	is.Equal(filepath.Join(home, ".ssh", "id"), expandHome("~/.ssh/id"))
	is.Equal(home, expandHome("~"))
	is.Equal("/etc/id", expandHome("/etc/id"))
	is.Equal("~user/id", expandHome("~user/id"))
}
