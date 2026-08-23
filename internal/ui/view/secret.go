package view

import (
	"fmt"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/repository"
	"github.com/myjupyter/conm/internal/secret"
)

type SecretRef = Ref[secret.Scheme]

type Secrets struct {
	repos map[secret.Scheme]repository.Secrets

	picking bool
	picked  string

	tableState[secret.Scheme]
}

func NewSecrets(repos ...repository.Secrets) *Secrets {
	v := &Secrets{
		repos: make(map[secret.Scheme]repository.Secrets, len(repos)),
	}

	for _, repo := range repos {
		if repo == nil || !v.register(repo.Kind()) {
			continue
		}

		v.repos[repo.Kind()] = repo
	}

	v.Reindex()

	return v
}

func (v *Secrets) Reindex() {
	entries := make([]entry[secret.Scheme], 0, v.Total())

	for _, kind := range v.idx.order {
		repo := v.repos[kind]
		for i := range repo.Len() {
			sec, ok := repo.Get(i)
			if !ok {
				continue
			}

			entries = append(entries, entry[secret.Scheme]{
				ref:    SecretRef{Kind: kind, Index: i},
				fields: secretFields(kind, sec),
			})
		}
	}

	v.reset(entries)
}

func (v *Secrets) CountOf(kind secret.Scheme) int {
	repo, ok := v.repos[kind]
	if !ok {
		return 0
	}

	return repo.Len()
}

func (v *Secrets) At(i int) (config.Secret, bool) {
	ref, ok := v.RefAt(i)
	if !ok {
		return nil, false
	}

	return v.For(ref)
}

func (v *Secrets) For(ref SecretRef) (config.Secret, bool) {
	repo, ok := v.repos[ref.Kind]
	if !ok {
		return nil, false
	}

	return repo.Get(ref.Index)
}

func (v *Secrets) UsagesAt(i int) []repository.Usage {
	ref, repo, err := v.rowRepo("usages", i)
	if err != nil {
		return nil
	}

	return repo.UsagesAt(ref.Index)
}

// LocationsOf lists the entries a store holds, in repository order and
// independent of the current query: a password field offers the store it names,
// not the rows the table happens to show.
func (v *Secrets) LocationsOf(kind secret.Scheme) []string {
	repo, ok := v.repos[kind]
	if !ok {
		return nil
	}

	locations := make([]string, 0, repo.Len())
	for i := range repo.Len() {
		if s, ok := repo.Get(i); ok {
			locations = append(locations, s.Location())
		}
	}

	return locations
}

func (v *Secrets) rowRepo(action string, i int) (SecretRef, repository.Secrets, error) {
	ref, ok := v.RefAt(i)
	if !ok {
		return SecretRef{}, nil, fmt.Errorf("%s: row %d is out of range", action, i)
	}

	repo, ok := v.repos[ref.Kind]
	if !ok {
		return SecretRef{}, nil, fmt.Errorf("%s: no repository for secret store %q", action, ref.Kind)
	}

	return ref, repo, nil
}

func secretFields(kind secret.Scheme, s config.Secret) []string {
	params := s.Params()

	values := make([]string, 0, 5+2*len(params))
	values = append(values, s.ID(), s.Provider(), s.Location(), s.Description(), kind)
	for _, p := range params {
		values = append(values, p.Name, p.Value)
	}

	return searchable(values...)
}
