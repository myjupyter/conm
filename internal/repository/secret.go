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

// SecretRepository owns both halves of a secret: the record — a reference plus
// metadata, persisted as TOML — and the material it points at, held by the
// provider. A password only ever crosses this API as an argument; it is never
// part of a config.Secret and never stored on the repository.
//
// Every operation touches the provider first and the file second. A record
// that outlives its material is visible and can be fixed; material without a
// record is invisible, and is reclaimed by Add adopting it.
//
// Nothing here is specific to a provider: one repository serves one scheme,
// whichever it is, and the provider is the only thing that knows where the
// material actually lives.
//
// TODO: type the material errors (material already exists at ref / no material
// found at ref) so callers can tell them apart from a failed add: both are
// returned after the record is saved, not instead of saving it.
type SecretRepository struct {
	mx       sync.RWMutex
	file     config.File[config.Secret]
	provider secret.Provider
	usage    UsageLookup
}

// newSecretRepository binds a file to the provider that backs it. The scheme
// is taken from the provider rather than passed alongside it, so a repository
// cannot be built that stores one scheme and resolves another.
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

// Get returns the record only. The material behind it is write-only from here;
// reading it is the resolver's job, on the connection path.
func (r *SecretRepository) Get(i int) (config.Secret, bool) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	return r.at(i)
}

// Add registers s and, when password is not empty, writes it to the provider.
//
// An empty password means the material already lives in the backend and only
// needs a record. If material is found where a password was given, the record
// is still saved — so the entry becomes visible and manageable — but the
// password is discarded rather than overwriting what is already there, and the
// returned error says so.
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

// Edit replaces the record at i and, when password is not empty, writes it to
// the provider. Overwriting the material of an unchanged reference is the
// point of the call, so it happens without complaint; a reference that moves
// is treated like Add — pre-existing material at the new location is adopted,
// not clobbered. The material at the old location is left alone: nothing
// points at it anymore, but the user did not ask for it to be deleted.
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

// Remove drops the material first and the record second, so a failure in
// between leaves a record the user can see and retry on.
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

// hasMaterial reports whether the provider already holds a password at ref.
//
// TODO: a failed Resolve reads as "nothing there", which also swallows a
// backend that is merely unreachable or locked. Providers have no shared way
// to say "absent" yet; give secret a not-found sentinel and split the two.
func (r *SecretRepository) hasMaterial(ctx context.Context, ref secretRef) bool {
	_, err := r.provider.Resolve(ctx, ref)

	return err == nil
}

// checkAdd validates s and returns its reference. It holds the read lock only:
// the provider call that follows must not run under the write lock.
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

// The commit half of each operation re-reads what the check half saw: the lock
// was dropped for the provider call, so the list may have moved underneath.

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

// secretRef adapts a record to secret.Reference: what the config model holds is
// the pointer to the material, never the material itself.
type secretRef string

func (r secretRef) SecretRef() string { return string(r) }

func makeSecRef(s config.Secret) secretRef {
	return secretRef(secret.Ref(s.Provider(), s.Location()))
}
