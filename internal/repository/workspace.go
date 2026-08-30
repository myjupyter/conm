package repository

import (
	"errors"
	"fmt"

	"github.com/myjupyter/conm/internal/config"
)

var connectionRepositories = map[config.ConnType]func(config.Conm) (*ConnectionRepository, error){
	config.PostgresConnType: NewPostgresRepository,
	config.MySQLConnType:    NewMySQLRepository,
	config.MSSQLConnType:    NewMSSQLRepository,
	config.RedisConnType:    NewRedisRepository,
}

type Workspace struct {
	Keyring *SecretRepository

	conns  []*ConnectionRepository
	byKind map[config.ConnType]Connections
}

func NewWorkspace(cfg config.Conm) (*Workspace, error) {
	w := &Workspace{byKind: make(map[config.ConnType]Connections, len(connectionRepositories))}

	users := make([]SecretUser, 0, len(connectionRepositories))
	for _, db := range cfg.Databases {
		if !db.Enabled {
			continue
		}

		open, known := connectionRepositories[db.Type]
		if !known {
			return nil, errors.Join(
				fmt.Errorf("no repository for connection type %q", db.Type),
				w.closeConnections(),
			)
		}

		repo, err := open(cfg)
		if err != nil {
			return nil, errors.Join(err, w.closeConnections())
		}
		if repo.Kind() != db.Type {
			return nil, errors.Join(
				fmt.Errorf("repository for %q claims connection type %q", db.Type, repo.Kind()),
				repo.Close(), w.closeConnections(),
			)
		}

		w.conns = append(w.conns, repo)
		w.byKind[repo.Kind()] = repo
		users = append(users, repo)
	}

	kr, err := NewKeyringRepository(NewIndex(users...))
	if err != nil {
		return nil, errors.Join(err, w.closeConnections())
	}
	w.Keyring = kr

	return w, nil
}

func (w *Workspace) Connections() []Connections {
	repos := make([]Connections, 0, len(w.conns))
	for _, repo := range w.conns {
		repos = append(repos, repo)
	}

	return repos
}

func (w *Workspace) ConnectionsOf(t config.ConnType) (Connections, bool) {
	repo, ok := w.byKind[t]
	return repo, ok
}

func (w *Workspace) Close() error {
	return errors.Join(w.Keyring.Close(), w.closeConnections())
}

func (w *Workspace) closeConnections() error {
	errs := make([]error, 0, len(w.conns))
	for _, repo := range w.conns {
		errs = append(errs, repo.Close())
	}

	return errors.Join(errs...)
}
