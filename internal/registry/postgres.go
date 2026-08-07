package registry

import (
	"errors"
	"fmt"
	"sync"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/network"
)

type CommonRegistry[C config.Connection] struct {
	file config.File[C]
	ncs  []network.Connection

	mx *sync.RWMutex
}

func newPostgresRegistry() (*CommonRegistry[config.Postgres], error) {
	configPath, err := config.PGFilePath()
	if err != nil {
		return nil, err
	}

	file, err := config.OpenConfig[*config.PostgresConfigWrapper, config.Postgres](configPath)
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
			conn, err = network.NewPGClient(config.Conm{}, c)
		} else {
			conn, err = network.NewNoClient(c)
		}
		if err != nil {
			return nil, fmt.Errorf("couldn't create connection: %w", err)
		}

		ncs = append(ncs, conn)
	}

	return &CommonRegistry[config.Postgres]{
		file: file,
		mx:   &sync.RWMutex{},
		ncs:  ncs,
	}, nil
}

func (r CommonRegistry[C]) Len() int {
	r.mx.RLock()
	defer r.mx.RUnlock()

	return len(r.ncs)
}

func (r CommonRegistry[C]) Get(i int) (network.Connection, bool) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	if i < 0 || i >= len(r.ncs) {
		return nil, false
	}

	return r.ncs[i], true
}

func (r CommonRegistry[C]) Add(cfg C) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if errs := cfg.Validate(); errs != nil {
		return fmt.Errorf("add: %w", errors.Join(errs...))
	}

	// TODO
	conn, err := network.NewConnection(config.Conm{}, cfg)
	if err != nil {
		return err
	}

	r.ncs = append(r.ncs, conn)
	r.file.Add(cfg)

	return nil
}
func (r CommonRegistry[C]) Edit(i int, cfg C) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if errs := cfg.Validate(); errs != nil {
		return fmt.Errorf("edit: %w", errors.Join(errs...))
	}

	if i < 0 || i >= len(r.ncs) {
		return fmt.Errorf("edit: connection index %d is out of range", i)
	}

	conn, err := network.NewConnection(config.Conm{}, cfg)
	if err != nil {
		return err
	}

	r.ncs[i] = conn
	r.file.Put(i, cfg)

	return r.file.Save()
}

func (r CommonRegistry[C]) Remove(i int) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if i < 0 || i >= len(r.ncs) {
		return fmt.Errorf("remove: connection index %d is out of range", i)
	}

	r.ncs = append(r.ncs[:i], r.ncs[i+1:]...)
	r.file.Remove(i)

	return r.file.Save()
}

func (r CommonRegistry[C]) Save() error {
	r.mx.Lock()
	defer r.mx.Unlock()

	return r.file.Save()
}
func (r CommonRegistry[C]) Close() error {
	r.mx.Lock()
	defer r.mx.Unlock()

	return r.file.Close()
}
