package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

type Secrets interface {
	Kind() secret.Scheme

	Len() int
	Get(int) (config.Secret, bool)
	UsagesAt(int) []Usage

	Add(ctx context.Context, s config.Secret, password string) error
	Edit(ctx context.Context, i int, s config.Secret, password string) error
	Remove(ctx context.Context, i int) error

	Save() error
	Close() error
}

type SecretRepository struct {
	mx       sync.RWMutex
	file     config.File[config.Secret]
	provider secret.Provider
	usage    UsageLookup
}

func newSecretRepository(p secret.Provider, file config.File[config.Secret], usage UsageLookup) (*SecretRepository, error) {
	if p.Scheme() == "" {
		return nil, fmt.Errorf("secret provider has no scheme of its own")
	}

	return &SecretRepository{
		file:     file,
		provider: p,
		usage:    usage,
	}, nil
}

func (r *SecretRepository) Kind() secret.Scheme {
	return r.provider.Scheme()
}

func (r *SecretRepository) Len() int {
	r.mx.RLock()
	defer r.mx.RUnlock()

	return r.file.Len()
}

func (r *SecretRepository) Get(i int) (config.Secret, bool) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	return r.at(i)
}

func (r *SecretRepository) Add(ctx context.Context, s config.Secret, password string) error {
	ref, err := r.checkAdd(s)
	if err != nil {
		return fmt.Errorf("add secret: %w", err)
	}

	// Outside the lock: a provider call is arbitrarily slow — it may prompt,
	// unlock a store, or cross the network — and the render path reads this
	// repository.
	exists := r.hasMaterial(ctx, ref)

	if !exists && password != "" {
		if err := r.provider.Store(ctx, ref, password); err != nil {
			return fmt.Errorf("add secret %q: %w", s.ID(), err)
		}
	}

	if err := r.commitAdd(s, ref); err != nil {
		return fmt.Errorf("add secret: %w", err)
	}

	// Both cases are informational: the record is on disk either way.
	switch {
	case exists && password != "":
		return fmt.Errorf("add secret %q: material already exists at %s, the password entered was not written", s.ID(), ref)
	case !exists && password == "":
		return fmt.Errorf("add secret %q: no material found at %s", s.ID(), ref)
	}

	return nil
}

func (r *SecretRepository) Edit(ctx context.Context, i int, s config.Secret, password string) error {
	old, ref, err := r.checkEdit(i, s)
	if err != nil {
		return fmt.Errorf("edit secret: %w", err)
	}

	// An unchanged reference is not probed: a metadata-only edit has no reason
	// to reach the provider at all.
	repointed := makeSecRef(old) != ref
	exists := repointed && r.hasMaterial(ctx, ref)

	if password != "" && !exists {
		if err := r.provider.Store(ctx, ref, password); err != nil {
			return fmt.Errorf("edit secret %q: %w", s.ID(), err)
		}
	}

	if err := r.commitEdit(i, s, old); err != nil {
		return fmt.Errorf("edit secret: %w", err)
	}

	switch {
	case exists && password != "":
		return fmt.Errorf("edit secret %q: material already exists at %s, the password entered was not written", s.ID(), ref)
	case repointed && !exists && password == "":
		return fmt.Errorf("edit secret %q: no material found at %s", s.ID(), ref)
	}

	return nil
}

func (r *SecretRepository) Remove(ctx context.Context, i int) error {
	old, ref, err := r.checkRemove(i)
	if err != nil {
		return fmt.Errorf("remove secret: %w", err)
	}

	if err := r.provider.Remove(ctx, ref); err != nil {
		return fmt.Errorf("remove secret %q: %w", old.ID(), err)
	}

	if err := r.commitRemove(i, ref); err != nil {
		return fmt.Errorf("remove secret: %w", err)
	}

	return nil
}

func (r *SecretRepository) Save() error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if err := r.file.Save(); err != nil {
		return fmt.Errorf("save secrets: %w", err)
	}

	return nil
}

func (r *SecretRepository) Close() error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if err := r.file.Close(); err != nil {
		return fmt.Errorf("close secrets: %w", err)
	}

	return nil
}

func (r *SecretRepository) UsagesAt(i int) []Usage {
	r.mx.RLock()
	defer r.mx.RUnlock()

	s, ok := r.at(i)
	if !ok {
		return nil
	}

	return r.usages(makeSecRef(s))
}

func (r *SecretRepository) hasMaterial(ctx context.Context, ref secretRef) bool {
	_, err := r.provider.Resolve(ctx, ref)

	return err == nil
}

func (r *SecretRepository) checkAdd(s config.Secret) (secretRef, error) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	if err := r.check(s); err != nil {
		return "", err
	}

	if errs := s.Validate(); errs != nil {
		return "", fmt.Errorf("validation error: %w", errors.Join(errs...))
	}

	ref := makeSecRef(s)
	if _, found := r.indexOf(ref); found {
		return "", fmt.Errorf("secret %q already exists", s.ID())
	}

	return ref, nil
}

func (r *SecretRepository) checkEdit(i int, s config.Secret) (config.Secret, secretRef, error) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	if err := r.check(s); err != nil {
		return nil, "", err
	}

	if errs := s.Validate(); errs != nil {
		return nil, "", fmt.Errorf("validation error: %w", errors.Join(errs...))
	}

	old, ok := r.at(i)
	if !ok {
		return nil, "", fmt.Errorf("index %d is out of range", i)
	}

	ref := makeSecRef(s)

	if key := makeSecRef(old); key != ref {
		// Re-pointing a secret would orphan every connection referencing it.
		if used := r.usages(key); len(used) > 0 {
			return nil, "", &ErrSecretInUse{ID: old.ID(), By: used}
		}

		if j, found := r.indexOf(ref); found && j != i {
			return nil, "", fmt.Errorf("secret %q already points at %s", r.file.Get(j).ID(), ref)
		}
	}

	return old, ref, nil
}

func (r *SecretRepository) checkRemove(i int) (config.Secret, secretRef, error) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	old, ok := r.at(i)
	if !ok {
		return nil, "", fmt.Errorf("index %d is out of range", i)
	}

	ref := makeSecRef(old)
	if used := r.usages(ref); len(used) > 0 {
		return nil, "", &ErrSecretInUse{ID: old.ID(), By: used}
	}

	return old, ref, nil
}

func (r *SecretRepository) commitAdd(s config.Secret, ref secretRef) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if _, found := r.indexOf(ref); found {
		return fmt.Errorf("secret %q already exists", s.ID())
	}

	r.file.Add(s)

	return r.file.Save()
}

func (r *SecretRepository) commitEdit(i int, s, old config.Secret) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	cur, ok := r.at(i)
	if !ok || makeSecRef(cur) != makeSecRef(old) {
		return fmt.Errorf("secret at index %d changed while it was being edited", i)
	}

	r.file.Put(i, s)

	return r.file.Save()
}

func (r *SecretRepository) commitRemove(i int, ref secretRef) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	cur, ok := r.at(i)
	if !ok || makeSecRef(cur) != ref {
		return fmt.Errorf("secret at index %d changed while it was being removed", i)
	}

	r.file.Remove(i)

	return r.file.Save()
}

func (r *SecretRepository) check(s config.Secret) error {
	if kind := r.Kind(); s.Provider() != kind {
		return fmt.Errorf("secret provider %q doesn't belong to the %q repository", s.Provider(), kind)
	}

	return nil
}

func (r *SecretRepository) usages(ref secretRef) []Usage {
	if r.usage == nil {
		return nil
	}

	return r.usage.Usages(ref.SecretRef())
}

func (r *SecretRepository) at(i int) (config.Secret, bool) {
	if i < 0 || i >= r.file.Len() {
		return nil, false
	}

	return r.file.Get(i), true
}

func (r *SecretRepository) indexOf(key secretRef) (int, bool) {
	for i := range r.file.Len() {
		if makeSecRef(r.file.Get(i)) == key {
			return i, true
		}
	}

	return 0, false
}

type secretRef string

func (r secretRef) SecretRef() string { return string(r) }

func makeSecRef(s config.Secret) secretRef {
	return secretRef(secret.Ref(s.Provider(), s.Location()))
}
