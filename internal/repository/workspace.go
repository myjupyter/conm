package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

type Workspaces interface {
	Len() int
	Get(int) (string, bool)
	Active() string
	Use(int) error

	Add(name string) error
	Rename(i int, name string) error
	Remove(ctx context.Context, i int) error
}

var _ Workspaces = (*WorkspaceRepository)(nil)

type WorkspaceRepository struct {
	mx     sync.RWMutex
	names  []string
	active string
}

func NewWorkspaceRepository() (*WorkspaceRepository, error) {
	r := &WorkspaceRepository{active: config.Workspace()}
	if err := r.reload(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *WorkspaceRepository) Len() int {
	r.mx.RLock()
	defer r.mx.RUnlock()

	return len(r.names)
}

func (r *WorkspaceRepository) Get(i int) (string, bool) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	return r.at(i)
}

func (r *WorkspaceRepository) Active() string {
	r.mx.RLock()
	defer r.mx.RUnlock()

	return r.active
}

func (r *WorkspaceRepository) Use(i int) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	name, ok := r.at(i)
	if !ok {
		return fmt.Errorf("index %d is out of range", i)
	}
	r.active = name

	return nil
}

func (r *WorkspaceRepository) Add(name string) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if err := config.CreateWorkspace(name); err != nil {
		return fmt.Errorf("add workspace: %w", err)
	}

	return r.reload()
}

func (r *WorkspaceRepository) Rename(i int, name string) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	old, ok := r.at(i)
	if !ok {
		return fmt.Errorf("rename workspace: index %d is out of range", i)
	}
	if old == name {
		return nil
	}
	if err := config.RenameWorkspace(old, name); err != nil {
		return fmt.Errorf("rename workspace: %w", err)
	}

	if r.active == old {
		r.active = name
	}

	return r.reload()
}

func (r *WorkspaceRepository) Remove(ctx context.Context, i int) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	name, ok := r.at(i)
	if !ok {
		return fmt.Errorf("remove workspace: index %d is out of range", i)
	}
	if name == r.active {
		return &ActiveWorkspaceError{Name: name}
	}

	if err := removeMaterial(ctx, name); err != nil {
		return fmt.Errorf("remove workspace: %w", err)
	}
	if err := config.RemoveWorkspace(name); err != nil {
		return fmt.Errorf("remove workspace: %w", err)
	}

	return r.reload()
}

func (r *WorkspaceRepository) reload() error {
	names, err := config.Workspaces()
	if err != nil {
		return err
	}
	r.names = names

	return nil
}

func (r *WorkspaceRepository) at(i int) (string, bool) {
	if i < 0 || i >= len(r.names) {
		return "", false
	}

	return r.names[i], true
}

func removeMaterial(ctx context.Context, name string) error {
	workspace, err := config.SecretWorkspaceOf(name)
	if err != nil {
		return err
	}
	file, err := config.OpenConfig[*config.KeyringConfigWrapper](config.SecretConfigPathOf(name))
	if err != nil {
		return err
	}

	resolver := secret.DefaultFor(workspace)
	errs := make([]error, 0, file.Len())
	for i := range file.Len() {
		errs = append(errs, resolver.Remove(ctx, makeSecRef(file.Get(i))))
	}

	return errors.Join(errs...)
}
