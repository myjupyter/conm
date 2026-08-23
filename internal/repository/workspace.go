package repository

import (
	"errors"
	"fmt"

	"github.com/myjupyter/conm/internal/config"
)

var connectionRepositories = []func(config.Conm) (*ConnectionRepository, error){
	NewPostgresRepository,
}

type Workspace struct {
	Keyring *SecretRepository

	conns  []*ConnectionRepository
	byKind map[config.ConnType]Connections
}

func NewWorkspace(cfg config.Conm) (*Workspace, error) {
	w := &Workspace{byKind: make(map[config.ConnType]Connections, len(connectionRepositories))}

	users := make([]SecretUser, 0, len(connectionRepositories))
	for _, open := range connectionRepositories {
		repo, err := open(cfg)
		if err != nil {
			return nil, errors.Join(err, w.closeConnections())
		}
		if _, taken := w.byKind[repo.Kind()]; taken {
			return nil, errors.Join(
				fmt.Errorf("two repositories claim connection type %q", repo.Kind()),
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
