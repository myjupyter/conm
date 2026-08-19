// TODO: complete it
package registry

import (
	"errors"
	"fmt"
	"sync"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

var _ Secrets[config.Secret] = (*SecretRegistry[config.Secret])(nil)

// Secrets is the mutation surface the UI uses for the secret store.
type Secrets[S config.Secret] interface {
	Len() int
	Get(int) (S, bool)

	Add(S) error
	Edit(int, S) error
	Remove(int) error

	// TODO
	// Link()
	// Unlink()

	Save() error
	Close() error
}

type SecretRegistry[S config.Secret] struct {
	mx   sync.RWMutex
	file config.File[S]
	refs secretUsageSet
}

func (r *SecretRegistry[S]) Len() int {
	r.mx.RLock()
	defer r.mx.RUnlock()

	return r.file.Len()
}

func (r *SecretRegistry[S]) Get(i int) (S, bool) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	return r.at(i)
}

func (r *SecretRegistry[S]) Add(s S) error {
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

func (r *SecretRegistry[S]) Edit(i int, s S) error {
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
		if n := r.refs.count(key); n > 0 {
			return fmt.Errorf("edit secret %q: secret is still in use by %d connection(s)", old.ID(), n)
		}
	}

	r.file.Put(i, s)

	if err := r.file.Save(); err != nil {
		return fmt.Errorf("edit secret: %w", err)
	}

	return nil
}

func (r *SecretRegistry[S]) Remove(i int) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	old, ok := r.at(i)
	if !ok {
		return fmt.Errorf("remove secret: index %d is out of range", i)
	}

	if n := r.refs.count(makeSecRef(old)); n > 0 {
		return fmt.Errorf("remove secret %q: secret is still in use by %d connection(s)", old.ID(), n)
	}

	r.file.Remove(i)

	if err := r.file.Save(); err != nil {
		return fmt.Errorf("remove secret: %w", err)
	}

	return nil
}

func (r *SecretRegistry[S]) Save() error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if err := r.file.Save(); err != nil {
		return fmt.Errorf("save secrets: %w", err)
	}

	return nil
}

func (r *SecretRegistry[S]) Close() error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if err := r.file.Close(); err != nil {
		return fmt.Errorf("close secrets: %w", err)
	}

	return nil
}

func (r *SecretRegistry[S]) at(i int) (S, bool) {
	if i < 0 || i >= r.file.Len() {
		var zero S
		return zero, false
	}

	return r.file.Get(i), true
}

func (r *SecretRegistry[S]) indexOf(key string) (int, bool) {
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

type secretUsageSet struct {
	refs map[string]int
}

func (s *secretUsageSet) link(rawRef string) {
	// No need to track literal secrets
	scheme, _, ok := secret.ParseRef(rawRef)
	if !ok || scheme == secret.Literal {
		return
	}

	if s.refs == nil {
		s.refs = make(map[string]int)
	}
	s.refs[rawRef]++
}

func (s *secretUsageSet) count(key string) int { return s.refs[key] }
