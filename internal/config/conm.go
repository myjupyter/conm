package config

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

type ConnMeta struct {
	Name        string   `toml:"name,omitempty"`
	Description string   `toml:"description,omitempty"`
	Tags        []string `toml:"tags,omitempty"`
}

type CLIInfo struct {
	Name   string
	Path   string
	Exists bool
}

type ConmPostgresSection struct {
	CLI string `toml:"cli"`
}

type Conm struct {
	Postgres ConmPostgresSection `toml:"postgres"`
}

func (c *Conm) SetCLI(t ConnType, cli string) {
	switch t {
	case PostgresConnType:
		c.Postgres = ConmPostgresSection{
			CLI: cli,
		}
	}
}

type ConmConfigWrapper struct {
	Conm Conm `toml:"conm"`
}

func (w *ConmConfigWrapper) Add(_ Conm) {
}

func (w *ConmConfigWrapper) Len() int {
	return 1
}

func (w *ConmConfigWrapper) Get(_ int) Conm {
	return w.Conm
}

func (w *ConmConfigWrapper) Put(_ int, conm Conm) {
	w.Conm = conm
}

func (w *ConmConfigWrapper) ConnectionConfigs() []Connection {
	return nil
}

func (w *ConmConfigWrapper) Remove(_ int) {}

func (w *ConmConfigWrapper) Validate() {
	//TODO: validate conm config
}

func (w *ConmConfigWrapper) Unmarshal(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	return toml.Unmarshal(data, w)
}

func (w *ConmConfigWrapper) Marshal() ([]byte, error) {
	return toml.Marshal(w)
}

// CreateConmConfigPath creates both the config dir (conm.toml, postgres.toml)
// and the data dir (secret.toml); XDG may place them under different roots.
func CreateConmConfigPath() error {
	dirs := []string{filepath.Dir(conmConfigPath), filepath.Dir(secretConfigPath)}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}

	return nil
}

func CreateConm(conf Conm) error {
	raw, err := toml.Marshal(struct {
		Conm Conm `toml:"conm"`
	}{Conm: conf})
	if err != nil {
		return err
	}

	return os.WriteFile(ConmPath(), raw, 0600)
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

	if t.Conm.Postgres.CLI == "" {
		// TODO
		t.Conm.Postgres.CLI = "psql"
	}

	return t.Conm, nil
}
