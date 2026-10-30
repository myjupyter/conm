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
	Rename(ctx context.Context, i int, name string) error
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

func (r *WorkspaceRepository) Rename(ctx context.Context, i int, name string) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	old, ok := r.at(i)
	if !ok {
		return fmt.Errorf("rename workspace: index %d is out of range", i)
	}
	if old == name {
		return nil
	}
	if err := config.ValidateWorkspaceName(name); err != nil {
		return fmt.Errorf("rename workspace: %w", err)
	}

	refs, err := keyringRefs(old)
	if err != nil {
		return fmt.Errorf("rename workspace: %w", err)
	}

	from, to := workspaceKeyring(old), workspaceKeyring(name)
	moved, err := copyMaterial(ctx, refs, from, to)
	if err == nil {
		err = config.RenameWorkspace(old, name)
	}
	if err != nil {
		return fmt.Errorf("rename workspace: %w", errors.Join(err, removeMaterial(ctx, moved, to)))
	}

	if err := removeMaterial(ctx, moved, from); err != nil {
		return fmt.Errorf("rename workspace: old keychain entries left behind: %w", err)
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
	if name == config.Workspace() {
		return &ActiveWorkspaceError{Name: name}
	}

	refs, err := keyringRefs(name)
	if err != nil {
		return fmt.Errorf("remove workspace: %w", err)
	}
	if err := removeMaterial(ctx, refs, workspaceKeyring(name)); err != nil {
		return fmt.Errorf("remove workspace: %w", err)
	}
	if err := config.RemoveWorkspace(name); err != nil {
		return fmt.Errorf("remove workspace: %w", err)
	}

	if r.active == name {
		r.active = config.Workspace()
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

func workspaceKeyring(name string) *secret.KeyringProvider {
	return &secret.KeyringProvider{Service: config.KeyringServiceOf(name)}
}

func keyringRefs(name string) ([]secretRef, error) {
	file, err := config.OpenConfig[*config.KeyringConfigWrapper](config.SecretConfigPathOf(name))
	if err != nil {
		return nil, err
	}

	var refs []secretRef
	for i := range file.Len() {
		ref := makeSecRef(file.Get(i))
		if !secret.NamesKeyringService(ref) {
			refs = append(refs, ref)
		}
	}

	return refs, nil
}

func copyMaterial(ctx context.Context, refs []secretRef, from, to secret.Provider) ([]secretRef, error) {
	copied := make([]secretRef, 0, len(refs))
	for _, ref := range refs {
		password, err := from.Resolve(ctx, ref)
		if err != nil {
			continue
		}
		if err := to.Store(ctx, ref, password); err != nil {
			return copied, err
		}
		copied = append(copied, ref)
	}

	return copied, nil
}

func removeMaterial(ctx context.Context, refs []secretRef, from secret.Provider) error {
	errs := make([]error, 0, len(refs))
	for _, ref := range refs {
		errs = append(errs, from.Remove(ctx, ref))
	}

	return errors.Join(errs...)
}
