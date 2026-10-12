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
	return fmt.Sprintf("secret %q is still in use by %d connection(s): %s", e.ID, len(e.By), usageList(e.By))
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

type TransportInUseError struct {
	Ref string
	By  []Usage
}

func (e *TransportInUseError) Error() string {
	return fmt.Sprintf("transport %q is still in use by %d connection(s): %s", e.Ref, len(e.By), usageList(e.By))
}

type ConnectionInUseError struct {
	Ref string
	By  []Usage
}

func (e *ConnectionInUseError) Error() string {
	return fmt.Sprintf("connection %q is still in use by %d transport(s): %s", e.Ref, len(e.By), usageList(e.By))
}

type TransportNotFoundError struct {
	Ref string
}

func (e *TransportNotFoundError) Error() string {
	return fmt.Sprintf("transport %q does not exist", e.Ref)
}

type ConnectionNotFoundError struct {
	Ref string
}

func (e *ConnectionNotFoundError) Error() string {
	return fmt.Sprintf("no connection %q to go through", e.Ref)
}

type DuplicateTransportError struct {
	Name string
}

func (e *DuplicateTransportError) Error() string {
	return fmt.Sprintf("a transport named %q already exists", e.Name)
}

func usageList(by []Usage) string {
	owners := make([]string, 0, len(by))
	for _, u := range by {
		owners = append(owners, u.String())
	}
	return strings.Join(owners, ", ")
}
