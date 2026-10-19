package repository

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"sync"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/network"
	"github.com/myjupyter/conm/internal/secret"
	"github.com/myjupyter/conm/internal/transport"
)

type Connections interface {
	Kind() config.ConnType
	ClientInfo(ctx context.Context) cli.Info

	Len() int
	IndexOf(name string) (int, bool)
	ConnectionAt(int) (network.Connection, bool)
	ConfigAt(int) (config.Connection, bool)

	Add(cfg config.Connection) error
	Edit(int, config.Connection) error
	Remove(int) error
}

type ConnectionRepository struct {
	launcher cli.Launcher
	kind     config.ConnType
	file     config.File[config.Connection]
	ncs      []network.Connection
	sec      secret.Provider

	transports transport.Provider
	usage      UsageLookup

	mx *sync.RWMutex
}

func (r *ConnectionRepository) TransportRefs() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		r.mx.RLock()
		defer r.mx.RUnlock()

		for i := range r.file.Len() {
			c := r.file.Get(i)
			ref, ok := c.(transport.Reference)
			if !ok || ref.TransportRef() == "" {
				continue
			}
			if !yield(c.Meta().Name, ref.TransportRef()) {
				return
			}
		}
	}
}

func (r *ConnectionRepository) Kind() config.ConnType {
	return r.kind
}

func (r *ConnectionRepository) ClientInfo(ctx context.Context) cli.Info {
	return r.launcher.Info(ctx)
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

			if !yield(c.Meta().Name, ref) {
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
	if err := r.checkTransport(cfg); err != nil {
		return fmt.Errorf("add: %w", err)
	}

	r.mx.Lock()
	defer r.mx.Unlock()

	const noSelf = -1
	if err := r.check(cfg, noSelf); err != nil {
		return fmt.Errorf("add: %w", err)
	}

	if errs := cfg.Validate(); errs != nil {
		return fmt.Errorf("add: validation error: %w", errors.Join(errs...))
	}

	conn, err := r.build(cfg)
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
	if err := r.checkTransport(cfg); err != nil {
		return fmt.Errorf("edit: %w", err)
	}
	if old, ok := r.ConfigAt(i); ok && old.Meta().Name != cfg.Meta().Name {
		if err := r.checkUnused(old); err != nil {
			return fmt.Errorf("edit: %w", err)
		}
	}

	r.mx.Lock()
	defer r.mx.Unlock()

	if err := r.check(cfg, i); err != nil {
		return fmt.Errorf("edit: %w", err)
	}

	if errs := cfg.Validate(); errs != nil {
		return fmt.Errorf("edit: validation error: %w", errors.Join(errs...))
	}

	if i < 0 || i >= len(r.ncs) {
		return fmt.Errorf("edit: connection index %d is out of range", i)
	}

	conn, err := r.build(cfg)
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
	if old, ok := r.ConfigAt(i); ok {
		if err := r.checkUnused(old); err != nil {
			return fmt.Errorf("remove: %w", err)
		}
	}

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

	return nil
}

func openConnections[W config.ConfigWrapper[config.Connection]](
	cfg config.Conm,
	kind config.ConnType,
	path string,
	transports transport.Provider,
	usage UsageLookup,
) (*ConnectionRepository, error) {
	file, err := config.OpenConfig[W](path)
	if err != nil {
		return nil, err
	}

	r := &ConnectionRepository{
		launcher: cli.For(cfg, kind),
		kind:     kind,
		file:     file,
		mx:       &sync.RWMutex{},
		sec:      secret.Default(),

		transports: transports,
		usage:      usage,
	}

	ncs := make([]network.Connection, 0, file.Len())
	for i := range file.Len() {
		conn, err := r.build(file.Get(i))
		if err != nil {
			return nil, fmt.Errorf("couldn't create connection: %w", err)
		}
		ncs = append(ncs, conn)
	}
	r.ncs = ncs

	return r, nil
}

func (r *ConnectionRepository) build(c config.Connection) (network.Connection, error) {
	if !c.IsValid() {
		return network.NewNoClient(c)
	}

	return network.NewRouted(r.launcher, c, r.transports, r.sec)
}

func (r *ConnectionRepository) IndexOf(name string) (int, bool) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	return r.indexOf(name)
}

func (r *ConnectionRepository) byName(name string) (config.Connection, bool) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	i, ok := r.indexOf(name)
	if !ok {
		return nil, false
	}

	return r.file.Get(i), true
}

func (r *ConnectionRepository) indexOf(name string) (int, bool) {
	for i := range r.file.Len() {
		if r.file.Get(i).Meta().Name == name {
			return i, true
		}
	}

	return -1, false
}

func (r *ConnectionRepository) check(cfg config.Connection, self int) error {
	if cfg.ConnType() != r.kind {
		return fmt.Errorf("connection type %q doesn't belong to the %q repository", cfg.ConnType(), r.kind)
	}

	return validateUnique(r.others(self), cfg)
}

func (r *ConnectionRepository) at(i int) (config.Connection, bool) {
	if i < 0 || i >= r.file.Len() {
		return nil, false
	}

	return r.file.Get(i), true
}

func (r *ConnectionRepository) checkTransport(cfg config.Connection) error {
	ref, ok := cfg.(transport.Reference)
	if !ok || ref.TransportRef() == "" {
		return nil
	}

	_, err := r.transports.Resolve(cfg)
	return err
}

func (r *ConnectionRepository) checkUnused(c config.Connection) error {
	ref := connectionRef(c.ConnType(), c.Meta().Name)
	if used := r.usage.Usages(ref); len(used) > 0 {
		return &ConnectionInUseError{Ref: ref, By: used}
	}

	return nil
}
