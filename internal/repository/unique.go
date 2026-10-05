package repository

import (
	"iter"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/utils"
)

func validateUnique(existing iter.Seq[config.Connection], c config.Connection) error {
	if other, ok := utils.FirstDuplicate(existing, config.Connection.Identity, c); ok {
		return &DuplicateConnectionError{
			Kind:   c.ConnType(),
			Name:   other.Meta().Name,
			Target: c.ConnectionString(""),
		}
	}

	if c.Meta().Name == "" {
		return nil
	}

	keyFn := func(c config.Connection) string {
		return c.Meta().Name
	}

	if _, ok := utils.FirstDuplicate(existing, keyFn, c); ok {
		return &DuplicateNameError{Kind: c.ConnType(), Name: c.Meta().Name}
	}

	return nil
}

func (r *ConnectionRepository) others(self int) iter.Seq[config.Connection] {
	return func(yield func(config.Connection) bool) {
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
