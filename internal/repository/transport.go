package repository

import (
	"errors"
	"fmt"
	"iter"
	"sync"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
	"github.com/myjupyter/conm/internal/transport"
	"github.com/myjupyter/conm/internal/utils"
)

type Transports interface {
	transport.Provider

	Len() int
	Get(int) (config.Transport, bool)
	UsagesAt(int) []Usage

	Add(config.Transport) error
	Edit(int, config.Transport) error
	Remove(int) error
}

type ConnectionFinder interface {
	Connection(t config.ConnType, name string) (config.Connection, bool)
}

var (
	_ Transports         = (*TransportRepository)(nil)
	_ transport.Provider = (*TransportRepository)(nil)
	_ ConnectionUser     = (*TransportRepository)(nil)
)

type TransportRepository struct {
	mx    sync.RWMutex
	file  config.File[config.Transport]
	usage UsageLookup
	conns ConnectionFinder
	cache *transport.Pool
}

func NewTransportRepository(usage UsageLookup, conns ConnectionFinder) (*TransportRepository, error) {
	file, err := config.OpenConfig[*config.TransportConfigWrapper](config.TransportPath())
	if err != nil {
		return nil, err
	}

	return newTransportRepository(file, usage, conns), nil
}

func newTransportRepository(
	file config.File[config.Transport],
	usage UsageLookup,
	conns ConnectionFinder,
) *TransportRepository {
	return &TransportRepository{
		file:  file,
		usage: usage,
		conns: conns,
		cache: transport.NewPool(cli.TransportLauncher{}, secret.Default()),
	}
}

func (r *TransportRepository) Kind() string {
	return "transport"
}

func (r *TransportRepository) Len() int {
	r.mx.RLock()
	defer r.mx.RUnlock()

	return r.file.Len()
}

func (r *TransportRepository) Get(i int) (config.Transport, bool) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	return r.at(i)
}

func (r *TransportRepository) UsagesAt(i int) []Usage {
	t, ok := r.Get(i)
	if !ok {
		return nil
	}

	return r.usage.Usages(ref(t))
}

func (r *TransportRepository) Resolve(target config.Connection) (transport.Transport, error) {
	ref, ok := target.(transport.Reference)
	if !ok {
		return nil, fmt.Errorf("%s connections cannot go through a transport", target.ConnType())
	}

	rec, err := r.record(ref.TransportRef())
	if err != nil {
		return nil, err
	}

	via, err := r.via(rec)
	if err != nil {
		return nil, err
	}

	return r.cache.Get(rec, via, target)
}

func (r *TransportRepository) ConnectionRefs() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		r.mx.RLock()
		defer r.mx.RUnlock()

		for i := range r.file.Len() {
			t := r.file.Get(i)
			through, ok := transport.Through(t.Kind)
			if !ok {
				continue
			}
			if !yield(t.Name, connectionRef(through, t.Via)) {
				return
			}
		}
	}
}

func (r *TransportRepository) Add(t config.Transport) error {
	if _, err := r.via(t); err != nil {
		return fmt.Errorf("add transport: %w", err)
	}
	if err := r.add(t); err != nil {
		return fmt.Errorf("add transport: %w", err)
	}

	return nil
}

func (r *TransportRepository) Edit(i int, t config.Transport) error {
	if _, err := r.via(t); err != nil {
		return fmt.Errorf("edit transport: %w", err)
	}
	if old, ok := r.Get(i); ok && ref(old) != ref(t) {
		if used := r.usage.Usages(ref(old)); len(used) > 0 {
			return fmt.Errorf("edit transport: %w", &TransportInUseError{Ref: ref(old), By: used})
		}
	}
	if err := r.edit(i, t); err != nil {
		return fmt.Errorf("edit transport: %w", err)
	}

	return nil
}

func (r *TransportRepository) Remove(i int) error {
	if old, ok := r.Get(i); ok {
		if used := r.usage.Usages(ref(old)); len(used) > 0 {
			return fmt.Errorf("remove transport: %w", &TransportInUseError{Ref: ref(old), By: used})
		}
	}

	r.mx.Lock()
	defer r.mx.Unlock()

	if _, ok := r.at(i); !ok {
		return fmt.Errorf("remove transport: index %d is out of range", i)
	}

	r.file.Remove(i)
	if err := r.file.Save(); err != nil {
		return fmt.Errorf("remove transport: %w", err)
	}

	return nil
}

func (r *TransportRepository) record(raw string) (config.Transport, error) {
	kind, name, ok := transport.ParseRef(raw)
	if !ok {
		return config.Transport{}, &TransportNotFoundError{Ref: raw}
	}

	r.mx.RLock()
	defer r.mx.RUnlock()

	for i := range r.file.Len() {
		if t := r.file.Get(i); t.Name == name && t.Kind == kind {
			return t, nil
		}
	}

	return config.Transport{}, &TransportNotFoundError{Ref: raw}
}

func (r *TransportRepository) via(t config.Transport) (config.Connection, error) {
	through, ok := transport.Through(t.Kind)
	if !ok {
		return nil, fmt.Errorf("unknown transport kind %q", t.Kind)
	}

	via, ok := r.conns.Connection(through, t.Via)
	if !ok {
		return nil, &ConnectionNotFoundError{Ref: connectionRef(through, t.Via)}
	}

	return via, nil
}

func (r *TransportRepository) add(t config.Transport) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if err := r.check(t, -1); err != nil {
		return err
	}

	r.file.Add(t)
	return r.file.Save()
}

func (r *TransportRepository) edit(i int, t config.Transport) error {
	r.mx.Lock()
	defer r.mx.Unlock()

	if _, ok := r.at(i); !ok {
		return fmt.Errorf("index %d is out of range", i)
	}
	if err := r.check(t, i); err != nil {
		return err
	}

	r.file.Put(i, t)
	return r.file.Save()
}

func (r *TransportRepository) check(t config.Transport, self int) error {
	if errs := t.Validate(); errs != nil {
		return fmt.Errorf("validation error: %w", errors.Join(errs...))
	}

	if _, ok := utils.FirstDuplicate(r.others(self), config.Transport.Identity, t); ok {
		return &DuplicateTransportError{Name: t.Name}
	}

	return nil
}

func (r *TransportRepository) others(self int) iter.Seq[config.Transport] {
	return func(yield func(config.Transport) bool) {
		for i := range r.file.Len() {
			if i == self {
				continue
			}
			if !yield(r.file.Get(i)) {
				return
			}
		}
	}
}

func (r *TransportRepository) at(i int) (config.Transport, bool) {
	if i < 0 || i >= r.file.Len() {
		return config.Transport{}, false
	}

	return r.file.Get(i), true
}

func ref(t config.Transport) string {
	return transport.Ref(t.Kind, t.Name)
}
