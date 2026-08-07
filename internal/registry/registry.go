package registry

import (
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/network"
)

type Registry[C config.Connection] interface {
	Len() int
	Get(int) (network.Connection, bool)
	Config(int) (C, bool)

	Add(cfg C) error
	Edit(int, C) error
	Remove(int) error

	Save() error
	Close() error
}
