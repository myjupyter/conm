package repository

import (
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

	Add(config.Secret) error
	Edit(int, config.Secret) error
	Remove(int) error

	Save() error
	Close() error
}

type SecretRepository struct {
	mx    sync.RWMutex
	kind  secret.Scheme
	file  config.File[config.Secret]
	usage UsageLookup
}

// Kind is the provider every secret in this repository is stored under; it is
// what callers type-assert Get's result against.
func (r *SecretRepository) Kind() secret.Scheme {
	return r.kind
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

func (r *SecretRepository) Add(s config.Secret) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if err := r.check(s); err != nil {
		return fmt.Errorf("add secret: %w", err)
	}

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

func (r *SecretRepository) Edit(i int, s config.Secret) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if err := r.check(s); err != nil {
		return fmt.Errorf("edit secret: %w", err)
	}

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

func (r *SecretRepository) Remove(i int) error {
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

// check rejects a secret the underlying file can't store: without the generic
// parameter the compiler no longer does it for us.
func (r *SecretRepository) check(s config.Secret) error {
	if s.Provider() != r.kind {
		return fmt.Errorf("secret provider %q doesn't belong to the %q repository", s.Provider(), r.kind)
	}

	return nil
}

func (r *SecretRepository) usages(ref string) []Usage {
	if r.usage == nil {
		return nil
	}

	return r.usage.Usages(ref)
}

func (r *SecretRepository) at(i int) (config.Secret, bool) {
	if i < 0 || i >= r.file.Len() {
		return nil, false
	}

	return r.file.Get(i), true
}

func (r *SecretRepository) indexOf(key string) (int, bool) {
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
