package registry

import (
	"fmt"
	"sync"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/network"
	"github.com/myjupyter/conm/internal/secret"
)

var _ Connections[config.Postgres] = (*ConnectionRegistry[config.Postgres])(nil)

func NewPostgresRegistry(cfg config.Conm) (*ConnectionRegistry[config.Postgres], error) {
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

	return &ConnectionRegistry[config.Postgres]{
		cfg:  cfg,
		file: file,
		mx:   &sync.RWMutex{},
		ncs:  ncs,
		sec:  sec,
	}, nil
}
