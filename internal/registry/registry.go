package registry

import (
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/network"
)

type Registry interface {
	Search(string)

	RestoreState()

	Len() int
	Get(int) (network.Connection, bool)

	Add(cfg config.Connection) error
	Edit(int, config.Connection) error
	Remove(int) error

	Save() error
	Close() error
}
