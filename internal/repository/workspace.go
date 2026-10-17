package repository

import (
	"errors"
	"fmt"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/transport"
)

type openRepository func(config.Conm, transport.Provider, UsageLookup) (*ConnectionRepository, error)

var connectionRepositories = map[config.ConnType]openRepository{
	config.PostgresConnType:   NewPostgresRepository,
	config.MySQLConnType:      NewMySQLRepository,
	config.MSSQLConnType:      NewMSSQLRepository,
	config.ClickHouseConnType: NewClickHouseRepository,
	config.RedisConnType:      NewRedisRepository,
	config.MongoDBConnType:    NewMongoDBRepository,
	config.SSHConnType:        NewSSHRepository,
}

var _ ConnectionFinder = (*Workspace)(nil)

type Workspace struct {
	Keyring    *SecretRepository
	Transports *TransportRepository

	conns  []*ConnectionRepository
	byKind map[config.ConnType]Connections
}

func NewWorkspace(cfg config.Conm) (*Workspace, error) {
	w := &Workspace{byKind: make(map[config.ConnType]Connections, len(connectionRepositories))}

	trs, err := NewTransportRepository(UsageLookupFunc(w.transportUsages), w)
	if err != nil {
		return nil, err
	}
	w.Transports = trs
	connectionUsage := NewConnectionIndex(trs)

	users := make([]SecretUser, 0, len(connectionRepositories))
	for _, db := range cfg.Connections {
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

		repo, err := open(cfg, trs, connectionUsage)
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
	return w.closeConnections()
}

func (w *Workspace) Connection(t config.ConnType, name string) (config.Connection, bool) {
	repo, ok := w.byKind[t].(*ConnectionRepository)
	if !ok {
		return nil, false
	}

	return repo.byName(name)
}

func (w *Workspace) transportUsages(ref string) []Usage {
	users := make([]TransportUser, 0, len(w.conns))
	for _, repo := range w.conns {
		users = append(users, repo)
	}

	return NewTransportIndex(users...).Usages(ref)
}

func (w *Workspace) closeConnections() error {
	errs := make([]error, 0, len(w.conns))
	for _, repo := range w.conns {
		errs = append(errs, repo.Close())
	}

	return errors.Join(errs...)
}
