package view

import (
	"fmt"
	"strconv"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/repository"
	"github.com/myjupyter/conm/internal/transport"
)

type TransportRef = Ref[config.TransportKind]

type Transports struct {
	repo repository.Transports

	tableState[config.TransportKind]
}

func NewTransports(repo repository.Transports) *Transports {
	v := &Transports{repo: repo}
	for _, kind := range config.TransportKinds {
		v.register(kind)
	}

	v.Reindex()

	return v
}

func (v *Transports) Resolve(target config.Connection) (transport.Transport, error) {
	return v.repo.Resolve(target)
}

func (v *Transports) Reindex() {
	entries := make([]entry[config.TransportKind], 0, v.repo.Len())
	for i := range v.repo.Len() {
		t, ok := v.repo.Get(i)
		if !ok {
			continue
		}

		entries = append(entries, entry[config.TransportKind]{
			ref:    TransportRef{Kind: t.Kind, Index: i},
			fields: transportFields(t),
		})
	}

	v.reset(entries)
}

func (v *Transports) CountOf(kind config.TransportKind) int {
	n := 0
	for i := range v.repo.Len() {
		if t, ok := v.repo.Get(i); ok && t.Kind == kind {
			n++
		}
	}

	return n
}

func (v *Transports) At(i int) (config.Transport, bool) {
	ref, ok := v.RefAt(i)
	if !ok {
		return config.Transport{}, false
	}

	return v.repo.Get(ref.Index)
}

func (v *Transports) Transport() (config.Transport, bool) {
	return v.At(v.Cursor())
}

func (v *Transports) UsagesAt(i int) []repository.Usage {
	ref, ok := v.RefAt(i)
	if !ok {
		return nil
	}

	return v.repo.UsagesAt(ref.Index)
}

func (v *Transports) Usages() []repository.Usage {
	return v.UsagesAt(v.Cursor())
}

func (v *Transports) NamesOf(kind config.TransportKind) []string {
	var names []string
	for i := range v.repo.Len() {
		if t, ok := v.repo.Get(i); ok && t.Kind == kind {
			names = append(names, t.Name)
		}
	}

	return names
}

func (v *Transports) Add(t config.Transport) error {
	err := v.repo.Add(t)
	v.Reindex()

	return err
}

func (v *Transports) Edit(t config.Transport) error {
	ref, ok := v.Ref()
	if !ok {
		return fmt.Errorf("edit: row %d is out of range", v.Cursor())
	}

	err := v.repo.Edit(ref.Index, t)
	v.Reindex()

	return err
}

func (v *Transports) Remove() error {
	ref, ok := v.Ref()
	if !ok {
		return fmt.Errorf("remove: row %d is out of range", v.Cursor())
	}

	err := v.repo.Remove(ref.Index)
	v.Reindex()

	return err
}

func transportFields(t config.Transport) []string {
	return searchable(t.Name, t.Kind, t.Via, strconv.Itoa(t.LocalPort), t.Description)
}
