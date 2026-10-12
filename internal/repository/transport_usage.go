package repository

import (
	"iter"

	"github.com/myjupyter/conm/internal/config"
)

type TransportUser interface {
	Kind() config.ConnType
	TransportRefs() iter.Seq2[string, string]
}

type TransportIndex struct {
	users []TransportUser
}

func NewTransportIndex(users ...TransportUser) *TransportIndex {
	return &TransportIndex{users: users}
}

func (ix *TransportIndex) Usages(ref string) []Usage {
	if ix == nil || ref == "" {
		return nil
	}

	var used []Usage
	for _, u := range ix.users {
		for owner, r := range u.TransportRefs() {
			if r == ref {
				used = append(used, Usage{Kind: u.Kind().String(), Owner: owner})
			}
		}
	}

	return used
}
