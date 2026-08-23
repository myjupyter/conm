package view

import (
	"fmt"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/network"
)

func (v *Connections) Connection() (network.Connection, bool) {
	return v.ConnectionAt(v.Cursor())
}

func (v *Connections) Config() (config.Connection, bool) {
	return v.ConfigAt(v.Cursor())
}

func (v *Connections) Add(cfg config.Connection) error {
	repo, ok := v.repos[cfg.ConnType()]
	if !ok {
		return fmt.Errorf("add: no repository for connection type %q", cfg.ConnType())
	}

	if err := repo.Add(cfg); err != nil {
		return err
	}

	v.Reindex()

	return nil
}

func (v *Connections) Edit(cfg config.Connection) error {
	ref, repo, err := v.rowRepo("edit", v.Cursor())
	if err != nil {
		return err
	}

	if err := repo.Edit(ref.Index, cfg); err != nil {
		return err
	}

	v.Reindex()

	return nil
}

func (v *Connections) Remove() error {
	ref, repo, err := v.rowRepo("remove", v.Cursor())
	if err != nil {
		return err
	}

	if err := repo.Remove(ref.Index); err != nil {
		return err
	}

	v.Reindex()

	return nil
}
