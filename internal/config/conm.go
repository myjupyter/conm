package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

const conmDirName = "conm"
const pgConfigFileName = "postgres.toml"
const conmConfigFileName = "conm.toml"

type ConnMeta struct {
	Name        string   `toml:"name,omitempty"`
	Description string   `toml:"description,omitempty"`
	Tags        []string `toml:"tags,omitempty"`
}

const PostgresCli = "pgcli"

type Conm struct {
	PostgresCli string `toml:"postgres_cli,omitempty"`
}

func ConmDirPath() (string, error) {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return "", fmt.Errorf("$HOME is not set")
	}

	defaultXdgConfigHomeDir := filepath.Join(homeDir, ".config")

	var xdgConfigHomeDir string
	if xdgConfigHomeDir = os.Getenv("XDG_CONFIG_HOME"); xdgConfigHomeDir == "" {
		xdgConfigHomeDir = defaultXdgConfigHomeDir
	}

	return filepath.Join(xdgConfigHomeDir, conmDirName), nil
}

func ConmConfigPath() (string, error) {
	conmDirPath, err := ConmDirPath()
	if err != nil {
		return "", err
	}

	return filepath.Join(conmDirPath, conmConfigFileName), nil
}

func PGFilePath() (string, error) {
	conmDirPath, err := ConmDirPath()
	if err != nil {
		return "", err
	}

	return filepath.Join(conmDirPath, pgConfigFileName), nil
}

func CreateConm(filename string, conf Conm) error {
	raw, err := toml.Marshal(struct {
		Conm Conm `toml:"conm"`
	}{Conm: conf})
	if err != nil {
		return err
	}

	return os.WriteFile(filename, raw, 0600)
}

func ReadConm(filename string) (Conm, error) {
	t := struct {
		Conm Conm `toml:"conm"`
	}{}

	raw, err := os.ReadFile(filename)
	if err != nil {
		return Conm{}, err
	}

	err = toml.Unmarshal(raw, &t)
	if err != nil {
		return Conm{}, err
	}

	if t.Conm.PostgresCli == "" {
		t.Conm.PostgresCli = PostgresCli
	}

	return t.Conm, nil
}
