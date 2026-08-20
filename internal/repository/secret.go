package repository

import (
	"errors"
	"fmt"
	"sync"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

var _ Secrets[config.Secret] = (*SecretRepository[config.Secret])(nil)

type Secrets[S config.Secret] interface {
	Len() int
	Get(int) (S, bool)
	UsagesAt(int) []Usage

	Add(S) error
	Edit(int, S) error
	Remove(int) error

	Save() error
	Close() error
}

type SecretRepository[S config.Secret] struct {
	mx    sync.RWMutex
	file  config.File[S]
	usage UsageLookup
}

func (r *SecretRepository[S]) Len() int {
	r.mx.RLock()
	defer r.mx.RUnlock()

	return r.file.Len()
}

func (r *SecretRepository[S]) Get(i int) (S, bool) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	return r.at(i)
}

func (r *SecretRepository[S]) Add(s S) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if errs := s.Validate(); errs != nil {
		return fmt.Errorf("add secret: validation error: %w", errors.Join(errs...))
	}

	if _, found := r.indexOf(makeSecRef(s)); found {
		return fmt.Errorf("add secret %q: secret already exists", s.ID())
	}

	r.file.Add(s)

	if err := r.file.Save(); err != nil {
		return fmt.Errorf("add secret: %w", err)
	}

	return nil
}

func (r *SecretRepository[S]) Edit(i int, s S) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if errs := s.Validate(); errs != nil {
		return fmt.Errorf("edit secret: validation error: %w", errors.Join(errs...))
	}

	old, ok := r.at(i)
	if !ok {
		return fmt.Errorf("edit secret: index %d is out of range", i)
	}

	// Re-pointing a secret would orphan every connection referencing it.
	if key := makeSecRef(old); key != makeSecRef(s) {
		if used := r.usages(key); len(used) > 0 {
			return fmt.Errorf("edit secret: %w", &ErrSecretInUse{ID: old.ID(), By: used})
		}
	}

	r.file.Put(i, s)

	if err := r.file.Save(); err != nil {
		return fmt.Errorf("edit secret: %w", err)
	}

	return nil
}

func (r *SecretRepository[S]) Remove(i int) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	old, ok := r.at(i)
	if !ok {
		return fmt.Errorf("remove secret: index %d is out of range", i)
	}

	if used := r.usages(makeSecRef(old)); len(used) > 0 {
		return fmt.Errorf("remove secret: %w", &ErrSecretInUse{ID: old.ID(), By: used})
	}

	r.file.Remove(i)

	if err := r.file.Save(); err != nil {
		return fmt.Errorf("remove secret: %w", err)
	}

	return nil
}

func (r *SecretRepository[S]) Save() error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if err := r.file.Save(); err != nil {
		return fmt.Errorf("save secrets: %w", err)
	}

	return nil
}

func (r *SecretRepository[S]) Close() error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if err := r.file.Close(); err != nil {
		return fmt.Errorf("close secrets: %w", err)
	}

	return nil
}

func (r *SecretRepository[S]) UsagesAt(i int) []Usage {
	r.mx.RLock()
	defer r.mx.RUnlock()

	s, ok := r.at(i)
	if !ok {
		return nil
	}

	return r.usages(makeSecRef(s))
}

func (r *SecretRepository[S]) usages(ref string) []Usage {
	if r.usage == nil {
		return nil
	}

	return r.usage.Usages(ref)
}

func (r *SecretRepository[S]) at(i int) (S, bool) {
	if i < 0 || i >= r.file.Len() {
		var zero S
		return zero, false
	}

	return r.file.Get(i), true
}

func (r *SecretRepository[S]) indexOf(key string) (int, bool) {
	for i := range r.file.Len() {
		if makeSecRef(r.file.Get(i)) == key {
			return i, true
		}
	}

	return 0, false
}

func makeSecRef(s config.Secret) string {
	return secret.Ref(s.Provider(), s.Location())
}
