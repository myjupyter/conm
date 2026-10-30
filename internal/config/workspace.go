package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strings"
	"unicode"

	"github.com/adrg/xdg"
	"github.com/pelletier/go-toml/v2"
)

const (
	DefaultWorkspace     = "default"
	keyringServicePrefix = "conm"
	workspaceNameMaxLen  = 63
	startupFileName      = "workspace.toml"
)

type startupFile struct {
	Workspace string `toml:"workspace"`
}

var workspace = DefaultWorkspace

func ValidateWorkspaceName(name string) error {
	switch {
	case name == "":
		return errors.New("workspace name can't be empty")
	case len(name) > workspaceNameMaxLen:
		return fmt.Errorf("workspace name must be at most %d characters long", workspaceNameMaxLen)
	case name == "." || name == "..":
		return fmt.Errorf("workspace name %q is not a folder name", name)
	case strings.ContainsAny(name, `/\`):
		return errors.New("workspace name can't contain a slash")
	case strings.IndexFunc(name, unicode.IsControl) >= 0:
		return errors.New("workspace name can't contain control characters")
	}
	return nil
}

func UseWorkspace(name string) error {
	if err := ValidateWorkspaceName(name); err != nil {
		return err
	}
	workspace = name
	return nil
}

func Workspace() string {
	return workspace
}

func KeyringService() string {
	return KeyringServiceOf(workspace)
}

func KeyringServiceOf(ws string) string {
	return keyringServicePrefix + "." + ws
}

func Workspaces() ([]string, error) {
	entries, err := os.ReadDir(workspaceDir(xdg.ConfigHome, ""))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}

	names := []string{workspace}
	for _, e := range entries {
		if e.IsDir() && !slices.Contains(names, e.Name()) {
			names = append(names, e.Name())
		}
	}
	slices.Sort(names)

	return names, nil
}

func CreateWorkspace(name string) error {
	if err := ValidateWorkspaceName(name); err != nil {
		return err
	}
	if err := refuseExistingWorkspace(name); err != nil {
		return err
	}

	return os.MkdirAll(workspaceDir(xdg.ConfigHome, name), 0o700)
}

func RenameWorkspace(from, to string) error {
	if err := ValidateWorkspaceName(to); err != nil {
		return err
	}
	if err := refuseExistingWorkspace(to); err != nil {
		return err
	}

	if err := os.MkdirAll(workspaceDir(xdg.ConfigHome, from), 0o700); err != nil {
		return err
	}
	if err := os.Rename(workspaceDir(xdg.ConfigHome, from), workspaceDir(xdg.ConfigHome, to)); err != nil {
		return err
	}

	err := os.Rename(workspaceDir(xdg.DataHome, from), workspaceDir(xdg.DataHome, to))
	if err == nil || errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	return errors.Join(err, os.Rename(workspaceDir(xdg.ConfigHome, to), workspaceDir(xdg.ConfigHome, from)))
}

func RemoveWorkspace(name string) error {
	if err := ValidateWorkspaceName(name); err != nil {
		return err
	}

	return errors.Join(
		os.RemoveAll(workspaceDir(xdg.ConfigHome, name)),
		os.RemoveAll(workspaceDir(xdg.DataHome, name)),
	)
}

func refuseExistingWorkspace(name string) error {
	if name == DefaultWorkspace {
		return fmt.Errorf("workspace %q already exists", name)
	}
	for _, home := range []string{xdg.ConfigHome, xdg.DataHome} {
		_, err := os.Stat(workspaceDir(home, name))
		if err == nil {
			return fmt.Errorf("workspace %q already exists", name)
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}
	return nil
}

func StartupWorkspace() (string, error) {
	raw, err := os.ReadFile(startupPath())
	if errors.Is(err, fs.ErrNotExist) {
		return DefaultWorkspace, nil
	}
	if err != nil {
		return "", err
	}

	var f startupFile
	if err := toml.Unmarshal(raw, &f); err != nil {
		return "", fmt.Errorf("couldn't parse file %q: %w", startupPath(), err)
	}
	if f.Workspace == "" {
		return DefaultWorkspace, nil
	}

	return f.Workspace, ValidateWorkspaceName(f.Workspace)
}

func SetStartupWorkspace(name string) error {
	if err := ValidateWorkspaceName(name); err != nil {
		return err
	}

	names, err := Workspaces()
	if err != nil {
		return err
	}
	if name != DefaultWorkspace && !slices.Contains(names, name) {
		return fmt.Errorf("no workspace named %q", name)
	}

	raw, err := toml.Marshal(startupFile{Workspace: name})
	if err != nil {
		return err
	}
	if err := os.MkdirAll(workspaceDir(xdg.ConfigHome, ""), 0o700); err != nil {
		return err
	}

	return writeFileAtomic(startupPath(), raw)
}

func startupPath() string {
	return workspaceDir(xdg.ConfigHome, startupFileName)
}
