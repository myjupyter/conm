package repository

import (
	"iter"

	"github.com/myjupyter/conm/internal/config"
)

type ConnectionUser interface {
	Kind() string
	ConnectionRefs() iter.Seq2[string, string]
}

type ConnectionIndex struct {
	users []ConnectionUser
}

func NewConnectionIndex(users ...ConnectionUser) *ConnectionIndex {
	return &ConnectionIndex{users: users}
}

func (ix *ConnectionIndex) Usages(ref string) []Usage {
	if ix == nil || ref == "" {
		return nil
	}

	var used []Usage
	for _, u := range ix.users {
		for owner, r := range u.ConnectionRefs() {
			if r == ref {
				used = append(used, Usage{Kind: u.Kind(), Owner: owner})
			}
		}
	}

	return used
}

func connectionRef(t config.ConnType, name string) string {
	if name == "" {
		return ""
	}
	return t.String() + ":" + name
}
