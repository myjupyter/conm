package registry

import (
	"errors"
	"fmt"
	"sync"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/network"
	"github.com/myjupyter/conm/internal/secret"
)

type Connections[C config.Connection] interface {
	Len() int
	Get(int) (network.Connection, bool)
	Config(int) (C, bool)

	Add(cfg C) error
	Edit(int, C) error
	Remove(int) error
}

type ConnectionRegistry[C config.Connection] struct {
	cfg  config.Conm
	file config.File[C]
	ncs  []network.Connection
	sec  secret.Provider

	mx *sync.RWMutex
}

func (r *ConnectionRegistry[C]) Len() int {
	r.mx.RLock()
	defer r.mx.RUnlock()

	return len(r.ncs)
}

func (r *ConnectionRegistry[C]) Get(i int) (network.Connection, bool) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	if i < 0 || i >= len(r.ncs) {
		return nil, false
	}

	return r.ncs[i], true
}

func (r *ConnectionRegistry[C]) Config(i int) (C, bool) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	return r.at(i)
}

func (r *ConnectionRegistry[C]) Add(cfg C) error {
	r.mx.Lock()
	defer r.mx.Unlock()

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
func (r *ConnectionRegistry[C]) Edit(i int, cfg C) error {
	r.mx.Lock()
	defer r.mx.Unlock()

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

func (r *ConnectionRegistry[C]) Remove(i int) error {
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

func (r *ConnectionRegistry[C]) Save() error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if err := r.file.Save(); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	return nil
}

func (r *ConnectionRegistry[C]) Close() error {
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

func (r *ConnectionRegistry[S]) at(i int) (S, bool) {
	if i < 0 || i >= r.file.Len() {
		var zero S
		return zero, false
	}

	return r.file.Get(i), true
}
