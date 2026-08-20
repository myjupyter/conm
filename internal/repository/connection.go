package repository

import (
	"errors"
	"fmt"
	"iter"
	"sync"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/network"
	"github.com/myjupyter/conm/internal/secret"
)

type Connections interface {
	Kind() config.ConnType

	Len() int
	ConnectionAt(int) (network.Connection, bool)
	ConfigAt(int) (config.Connection, bool)

	Add(cfg config.Connection) error
	Edit(int, config.Connection) error
	Remove(int) error
}

type ConnectionRepository struct {
	cfg  config.Conm
	kind config.ConnType
	file config.File[config.Connection]
	ncs  []network.Connection
	sec  secret.Provider

	mx *sync.RWMutex
}

func (r *ConnectionRepository) Kind() config.ConnType {
	return r.kind
}

func (r *ConnectionRepository) SecretRefs() iter.Seq2[string, secret.Reference] {
	return func(yield func(string, secret.Reference) bool) {
		r.mx.RLock()
		defer r.mx.RUnlock()

		for i := range r.file.Len() {
			c := r.file.Get(i)

			ref, ok := c.(secret.Reference)
			if !ok {
				// TODO: log and maybe action
				continue
			}

			if !yield(c.Name(), ref) {
				return
			}
		}
	}
}

func (r *ConnectionRepository) Len() int {
	r.mx.RLock()
	defer r.mx.RUnlock()

	return len(r.ncs)
}

func (r *ConnectionRepository) ConnectionAt(i int) (network.Connection, bool) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	if i < 0 || i >= len(r.ncs) {
		return nil, false
	}

	return r.ncs[i], true
}

func (r *ConnectionRepository) ConfigAt(i int) (config.Connection, bool) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	return r.at(i)
}

func (r *ConnectionRepository) Add(cfg config.Connection) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if err := r.check(cfg); err != nil {
		return fmt.Errorf("add: %w", err)
	}

	if errs := cfg.Validate(); errs != nil {
		return fmt.Errorf("add: validation error: %w", errors.Join(errs...))
	}

	conn, err := network.NewConnection(r.cfg, cfg, r.sec)
	if err != nil {
		return fmt.Errorf("add: open new connection error: %w", err)
	}

	r.ncs = append(r.ncs, conn)
	r.file.Add(cfg)

	if err := r.file.Save(); err != nil {
		return fmt.Errorf("add: %w", err)
	}

	return nil
}
func (r *ConnectionRepository) Edit(i int, cfg config.Connection) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if err := r.check(cfg); err != nil {
		return fmt.Errorf("edit: %w", err)
	}

	if errs := cfg.Validate(); errs != nil {
		return fmt.Errorf("edit: validation error: %w", errors.Join(errs...))
	}

	if i < 0 || i >= len(r.ncs) {
		return fmt.Errorf("edit: connection index %d is out of range", i)
	}

	conn, err := network.NewConnection(r.cfg, cfg, r.sec)
	if err != nil {
		return fmt.Errorf("edit: open new connection error: %w", err)
	}

	oldNC := r.ncs[i]
	r.ncs[i] = conn
	r.file.Put(i, cfg)

	if err := r.file.Save(); err != nil {
		return fmt.Errorf("edit: %w", err)
	}

	// TODO: log error
	_ = oldNC.Close()

	return nil
}

func (r *ConnectionRepository) Remove(i int) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if i < 0 || i >= len(r.ncs) {
		return fmt.Errorf("remove: connection index %d is out of range", i)
	}

	oldNC := r.ncs[i]
	r.ncs[i] = nil
	r.ncs = append(r.ncs[:i], r.ncs[i+1:]...)
	r.file.Remove(i)

	if err := r.file.Save(); err != nil {
		return fmt.Errorf("remove: %w", err)
	}

	// TODO: log error
	_ = oldNC.Close()

	return nil
}

func (r *ConnectionRepository) Save() error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if err := r.file.Save(); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	return nil
}

func (r *ConnectionRepository) Close() error {
	r.mx.Lock()
	defer r.mx.Unlock()

	for _, c := range r.ncs {
		// silently close connections even if it fails
		// TODO: log error
		_ = c.Close()
	}

	if err := r.file.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
	}

	return nil
}

func (r *ConnectionRepository) check(cfg config.Connection) error {
	if cfg.ConnType() != r.kind {
		return fmt.Errorf("connection type %q doesn't belong to the %q repository", cfg.ConnType(), r.kind)
	}

	return nil
}

func (r *ConnectionRepository) at(i int) (config.Connection, bool) {
	if i < 0 || i >= r.file.Len() {
		return nil, false
	}

	return r.file.Get(i), true
}
