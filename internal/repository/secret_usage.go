package repository

import (
	"fmt"
	"iter"
	"strings"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

type SecretUser interface {
	Kind() config.ConnType
	SecretRefs() iter.Seq2[string, secret.Reference]
}

type Usage struct {
	Kind  string
	Owner string
}

func (u Usage) String() string { return u.Kind + "/" + u.Owner }

type UsageLookup interface {
	Usages(ref string) []Usage
}

type Index struct {
	users []SecretUser
}

func NewIndex(users ...SecretUser) *Index {
	return &Index{users: users}
}

func (ix *Index) Usages(ref string) []Usage {
	if ix == nil || !trackable(ref) {
		return nil
	}

	var used []Usage
	for _, u := range ix.users {
		for owner, r := range u.SecretRefs() {
			if r.SecretRef() == ref {
				used = append(used, Usage{Kind: u.Kind().String(), Owner: owner})
			}
		}
	}

	return used
}

func trackable(raw string) bool {
	scheme, _, ok := secret.ParseRef(raw)
	return ok && secret.IsStore(scheme)
}

type SecretInUseError struct {
	ID string
	By []Usage
}

func (e *SecretInUseError) Error() string {
	owners := make([]string, 0, len(e.By))
	for _, u := range e.By {
		owners = append(owners, u.String())
	}

	return fmt.Sprintf("secret %q is still in use by %d connection(s): %s",
		e.ID, len(e.By), strings.Join(owners, ", "))
}
