package repository

import (
	"iter"
	"net"
	"strconv"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/utils"
)

type connKey struct {
	Type     config.ConnType
	Host     string
	Port     int
	Database string
	Schema   string
	Username string
}

func keyOf(c config.Connection) connKey {
	return connKey{
		Type:     c.ConnType(),
		Host:     c.Host(),
		Port:     c.Port(),
		Database: c.Database(),
		Schema:   c.Schema(),
		Username: c.Username(),
	}
}

func (k connKey) String() string {
	target := k.Host
	if k.Port != 0 {
		target = net.JoinHostPort(k.Host, strconv.Itoa(k.Port))
	}
	if k.Username != "" {
		target = k.Username + "@" + target
	}
	if k.Database != "" {
		target += "/" + k.Database
	}
	if k.Schema != "" {
		target += "." + k.Schema
	}

	return target
}

func validateUnique(existing iter.Seq[config.Connection], c config.Connection) error {
	if other, ok := utils.FirstDuplicate(existing, keyOf, c); ok {
		return &DuplicateConnectionError{
			Kind:   c.ConnType(),
			Name:   other.Meta().Name,
			Target: keyOf(c).String(),
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
