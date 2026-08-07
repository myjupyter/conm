package registry

import (
	"errors"
	"fmt"
	"sync"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/network"
)

var _ Registry[config.Postgres] = (*CommonRegistry[config.Postgres])(nil)

type CommonRegistry[C config.Connection] struct {
	cfg  config.Conm
	file config.File[C]
	ncs  []network.Connection

	mx *sync.RWMutex
}

func NewPostgresRegistry(cfg config.Conm) (*CommonRegistry[config.Postgres], error) {
	configPath, err := config.PGFilePath()
	if err != nil {
		return nil, err
	}

	file, err := config.OpenConfig[*config.PostgresConfigWrapper](configPath)
	if err != nil {
		return nil, err
	}

	n := file.Len()
	ncs := make([]network.Connection, 0, n)
	for i := 0; i < n; i++ {
		c := file.Get(i)

		var (
			conn network.Connection
			err  error
		)
		if c.IsValid() {
			conn, err = network.NewPGClient(cfg, c)
		} else {
			conn, err = network.NewNoClient(c)
		}
		if err != nil {
			return nil, fmt.Errorf("couldn't create connection: %w", err)
		}

		ncs = append(ncs, conn)
	}

	return &CommonRegistry[config.Postgres]{
		cfg:  cfg,
		file: file,
		mx:   &sync.RWMutex{},
		ncs:  ncs,
	}, nil
}

func (r *CommonRegistry[C]) Len() int {
	r.mx.RLock()
	defer r.mx.RUnlock()

	return len(r.ncs)
}

func (r *CommonRegistry[C]) Get(i int) (network.Connection, bool) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	if i < 0 || i >= len(r.ncs) {
		return nil, false
	}

	return r.ncs[i], true
}

func (r *CommonRegistry[C]) Add(cfg C) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if errs := cfg.Validate(); errs != nil {
		return fmt.Errorf("add: validation error: %w", errors.Join(errs...))
	}

	conn, err := network.NewConnection(r.cfg, cfg)
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
func (r *CommonRegistry[C]) Edit(i int, cfg C) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if errs := cfg.Validate(); errs != nil {
		return fmt.Errorf("edit: validation error: %w", errors.Join(errs...))
	}

	if i < 0 || i >= len(r.ncs) {
		return fmt.Errorf("edit: connection index %d is out of range", i)
	}

	conn, err := network.NewConnection(r.cfg, cfg)
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

func (r *CommonRegistry[C]) Remove(i int) error {
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

func (r *CommonRegistry[C]) Save() error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if err := r.file.Save(); err != nil {
		return fmt.Errorf("save: %w", err)
	}

	return nil
}
func (r *CommonRegistry[C]) Close() error {
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
