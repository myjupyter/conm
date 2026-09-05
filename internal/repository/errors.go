package repository

import (
	"fmt"
	"strings"

	"github.com/myjupyter/conm/internal/config"
)

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

type DuplicateConnectionError struct {
	Kind   config.ConnType
	Name   string
	Target string
}

func (e *DuplicateConnectionError) Error() string {
	if e.Name == "" {
		return fmt.Sprintf("%s already has a connection to %s", e.Kind, e.Target)
	}

	return fmt.Sprintf("%s already has this connection, named %q", e.Kind, e.Name)
}

type DuplicateNameError struct {
	Kind config.ConnType
	Name string
}

func (e *DuplicateNameError) Error() string {
	return fmt.Sprintf("%s already has a connection named %q", e.Kind, e.Name)
}
