package e2e

import (
	"bytes"
	"context"
	"flag"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"

	"github.com/myjupyter/conm/cmd"
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/importer"
	"github.com/myjupyter/conm/internal/secret"
	"github.com/myjupyter/conm/internal/ui"
)

var updateGolden = flag.Bool("update", false, "rewrite the golden files under testdata")

var e2eTestdata = mustAbs("testdata")

func mustAbs(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}

	return abs
}

type e2eEnv struct {
	t    *testing.T
	root string
}

func newE2EEnv(t *testing.T) *e2eEnv {
	t.Helper()

	root := t.TempDir()
	for _, dir := range []string{"home", "config", "data", "work"} {
		require.NoError(t, os.MkdirAll(filepath.Join(root, dir), 0o700))
	}

	t.Setenv("HOME", filepath.Join(root, "home"))
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(root, "data"))
	t.Setenv("PGSERVICEFILE", "")
	t.Setenv("PGPASSFILE", filepath.Join(root, "home", ".pgpass"))
	xdg.Reload()
	t.Cleanup(xdg.Reload)
	keyring.MockInit()
	t.Chdir(filepath.Join(root, "work"))

	return &e2eEnv{t: t, root: root}
}

func (e *e2eEnv) path(name string) string {
	return filepath.Join(e.root, "work", name)
}

func (e *e2eEnv) write(name, content string) string {
	e.t.Helper()

	path := e.path(name)
	require.NoError(e.t, os.MkdirAll(filepath.Dir(path), 0o700))
	require.NoError(e.t, os.WriteFile(path, []byte(content), 0o600))

	return path
}

func (e *e2eEnv) read(path string) string {
	e.t.Helper()

	raw, err := os.ReadFile(path)
	require.NoError(e.t, err)

	return string(raw)
}

func (e *e2eEnv) run(args ...string) (string, error) {
	e.t.Helper()

	var out bytes.Buffer

	root := cmd.NewRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)
	err := root.ExecuteContext(context.Background())

	return out.String(), err
}

func (e *e2eEnv) mustRun(args ...string) string {
	e.t.Helper()

	out, err := e.run(args...)
	require.NoError(e.t, err, out)

	return out
}

func (e *e2eEnv) connections(t config.ConnType, workspace string) []config.Connection {
	e.t.Helper()

	require.NoError(e.t, config.UseWorkspace(workspace))
	conm, err := config.ReadConm(config.ConmPath())
	require.NoError(e.t, err)

	conns, err := ui.Connections(conm, t)
	require.NoError(e.t, err)

	return conns
}

func (e *e2eEnv) passwords(t config.ConnType, workspace string) map[string]string {
	e.t.Helper()

	stored := map[string]string{}
	conns := e.connections(t, workspace)
	service, err := config.SecretWorkspace()
	require.NoError(e.t, err)

	for _, conn := range conns {
		ref, ok := conn.(secret.Reference)
		if !ok {
			continue
		}

		scheme, location, _ := secret.ParseRef(ref.SecretRef())
		if scheme != secret.Keyring {
			continue
		}

		password, err := keyring.Get(service, location)
		require.NoError(e.t, err, conn.Meta().Name)
		stored[conn.Meta().Name] = password
	}

	return stored
}

func (e *e2eEnv) seed(t config.ConnType, conns ...config.Connection) {
	e.t.Helper()

	require.NoError(e.t, config.UseWorkspace(config.DefaultWorkspace))
	conm, err := config.ReadConm(config.ConmPath())
	require.NoError(e.t, err)

	entries := make([]importer.Entry, 0, len(conns))
	for _, conn := range conns {
		entries = append(entries, importer.Entry{Name: conn.Meta().Name, Conn: conn})
	}

	results, err := ui.ImportConnections(context.Background(), conm, t, entries, ui.ImportStrict)
	require.NoError(e.t, err)
	for _, res := range results {
		require.NotEqual(e.t, ui.ImportFailed, res.Status, "%s: %s", res.Name, res.Note)
	}
}

func names(conns []config.Connection) []string {
	found := make([]string, 0, len(conns))
	for _, conn := range conns {
		found = append(found, conn.Meta().Name)
	}
	slices.Sort(found)

	return found
}

func fileNames(t *testing.T, kind config.ConnType, path string) []string {
	t.Helper()

	entries, err := importer.Read(kind, path)
	require.NoError(t, err)

	found := make([]string, 0, len(entries))
	for _, entry := range entries {
		found = append(found, entry.Name)
	}
	slices.Sort(found)

	return found
}

func linesKept(kept, in string) bool {
	lines := strings.Split(in, "\n")
	for line := range strings.SplitSeq(strings.TrimRight(kept, "\n"), "\n") {
		at := slices.Index(lines, line)
		if at < 0 {
			return false
		}
		lines = lines[at+1:]
	}

	return true
}

func golden(t *testing.T, name, got string) string {
	t.Helper()

	path := filepath.Join(e2eTestdata, name)
	if *updateGolden {
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
		require.NoError(t, os.WriteFile(path, []byte(got), 0o600))
	}

	want, err := os.ReadFile(path)
	require.NoError(t, err, "run the tests with -update to create %s", path)

	return string(want)
}

func fixture(t *testing.T, name string) string {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join(e2eTestdata, name))
	require.NoError(t, err)

	return string(raw)
}
