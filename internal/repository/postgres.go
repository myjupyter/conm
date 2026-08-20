package repository

import (
	"fmt"
	"sync"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/network"
	"github.com/myjupyter/conm/internal/secret"
)

var (
	_ Connections[config.Postgres] = (*ConnectionRepository[config.Postgres])(nil)
	_ SecretUser                   = (*ConnectionRepository[config.Postgres])(nil)
)

func NewPostgresRepository(cfg config.Conm) (*ConnectionRepository[config.Postgres], error) {
	file, err := config.OpenConfig[*config.PostgresConfigWrapper](config.PostgresPath())
	if err != nil {
		return nil, err
	}

	var sec secret.Provider
	sec = secret.Default()

	n := file.Len()
	ncs := make([]network.Connection, 0, n)
	for i := range n {
		c := file.Get(i)

		var (
			conn network.Connection
			err  error
		)
		if c.IsValid() {
			conn, err = network.NewPGClient(cfg, c, sec)
		} else {
			conn, err = network.NewNoClient(c)
		}
		if err != nil {
			return nil, fmt.Errorf("couldn't create connection: %w", err)
		}

		ncs = append(ncs, conn)
	}

	return &ConnectionRepository[config.Postgres]{
		cfg:  cfg,
		kind: config.PostgresConnType,
		file: file,
		mx:   &sync.RWMutex{},
		ncs:  ncs,
		sec:  sec,
	}, nil
}
