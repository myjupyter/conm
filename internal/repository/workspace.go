package repository

import (
	"errors"

	"github.com/myjupyter/conm/internal/config"
)

type Workspace struct {
	Postgres *ConnectionRepository[config.Postgres]
	Keyring  *SecretRepository[config.Keyring]
}

func NewWorkspace(cfg config.Conm) (*Workspace, error) {
	pg, err := NewPostgresRepository(cfg)
	if err != nil {
		return nil, err
	}

	index := NewIndex(pg)

	kr, err := NewKeyringRepository(index)
	if err != nil {
		return nil, errors.Join(err, pg.Close())
	}

	return &Workspace{
		Postgres: pg,
		Keyring:  kr,
	}, nil
}

func (w *Workspace) Close() error {
	return errors.Join(w.Keyring.Close(), w.Postgres.Close())
}
