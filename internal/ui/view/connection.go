package view

import (
	"context"
	"fmt"
	"slices"
	"strconv"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/network"
	"github.com/myjupyter/conm/internal/repository"
)

type ConnRef = Ref[config.ConnType]

type Connections struct {
	repos map[config.ConnType]repository.Connections

	tableState[config.ConnType]
}

func NewConnections(repos ...repository.Connections) *Connections {
	v := &Connections{
		repos: make(map[config.ConnType]repository.Connections, len(repos)),
	}

	for _, repo := range repos {
		if repo == nil || !v.register(repo.Kind()) {
			continue
		}

		v.repos[repo.Kind()] = repo
	}

	v.Reindex()
	if i := slices.IndexFunc(v.tableState.Kinds(), v.filled); i >= 0 {
		v.SetActive(v.tableState.Kinds()[i])
	}

	return v
}

func (v *Connections) Kinds() []config.ConnType {
	return slices.DeleteFunc(v.tableState.Kinds(), func(kind config.ConnType) bool {
		return kind != v.Active() && !v.filled(kind)
	})
}

func (v *Connections) NextKind() bool {
	return v.stepKind(1)
}

func (v *Connections) PrevKind() bool {
	return v.stepKind(-1)
}

func (v *Connections) stepKind(step int) bool {
	kinds := v.Kinds()
	if len(kinds) < 2 {
		return false
	}

	i := slices.Index(kinds, v.Active())

	return v.SetActive(kinds[(i+step+len(kinds))%len(kinds)])
}

func (v *Connections) filled(kind config.ConnType) bool {
	return v.CountOf(kind) > 0
}

func (v *Connections) Reindex() {
	entries := make([]entry[config.ConnType], 0, v.Total())

	for _, kind := range v.idx.order {
		repo := v.repos[kind]
		for i := range repo.Len() {
			cfg, ok := repo.ConfigAt(i)
			if !ok {
				continue
			}

			entries = append(entries, entry[config.ConnType]{
				ref:    ConnRef{Kind: kind, Index: i},
				fields: connectionFields(cfg),
			})
		}
	}

	v.reset(entries)
}

func (v *Connections) CountOf(kind config.ConnType) int {
	repo, ok := v.repos[kind]
	if !ok {
		return 0
	}

	return repo.Len()
}

func (v *Connections) Named(kind config.ConnType, name string) (config.Connection, bool) {
	repo, ok := v.repos[kind]
	if !ok || name == "" {
		return nil, false
	}

	for i := range repo.Len() {
		if cfg, ok := repo.ConfigAt(i); ok && cfg.Meta().Name == name {
			return cfg, true
		}
	}

	return nil, false
}

func (v *Connections) NamesOf(kind config.ConnType) []string {
	repo, ok := v.repos[kind]
	if !ok {
		return nil
	}

	names := make([]string, 0, repo.Len())
	for i := range repo.Len() {
		if cfg, ok := repo.ConfigAt(i); ok && cfg.Meta().Name != "" {
			names = append(names, cfg.Meta().Name)
		}
	}

	return names
}

func (v *Connections) ClientInfo(ctx context.Context) (cli.Info, bool) {
	repo, ok := v.repos[v.Active()]
	if !ok {
		return cli.Info{}, false
	}

	return repo.ClientInfo(ctx), true
}

func (v *Connections) ConnectionAt(i int) (network.Connection, bool) {
	ref, ok := v.RefAt(i)
	if !ok {
		return nil, false
	}

	return v.ConnectionFor(ref)
}

func (v *Connections) ConfigAt(i int) (config.Connection, bool) {
	ref, ok := v.RefAt(i)
	if !ok {
		return nil, false
	}

	return v.ConfigFor(ref)
}

func (v *Connections) ConnectionFor(ref ConnRef) (network.Connection, bool) {
	repo, ok := v.repos[ref.Kind]
	if !ok {
		return nil, false
	}

	return repo.ConnectionAt(ref.Index)
}

func (v *Connections) ConfigFor(ref ConnRef) (config.Connection, bool) {
	repo, ok := v.repos[ref.Kind]
	if !ok {
		return nil, false
	}

	return repo.ConfigAt(ref.Index)
}

func (v *Connections) rowRepo(action string, i int) (ConnRef, repository.Connections, error) {
	ref, ok := v.RefAt(i)
	if !ok {
		return ConnRef{}, nil, fmt.Errorf("%s: row %d is out of range", action, i)
	}

	repo, ok := v.repos[ref.Kind]
	if !ok {
		return ConnRef{}, nil, fmt.Errorf("%s: no repository for connection type %q", action, ref.Kind)
	}

	return ref, repo, nil
}

func connectionFields(c config.Connection) []string {
	meta := c.Meta()
	values := append([]string{
		meta.Name,
		meta.Description,
		c.Username(),
		c.Host(),
		strconv.Itoa(c.Port()),
		c.ConnType().String(),
	}, meta.Tags...)
	if db, ok := c.(config.DBConnection); ok {
		values = append(values, db.Database(), db.Schema())
	}

	return searchable(values...)
}
