package repository

import (
	"fmt"
	"sync"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/network"
	"github.com/myjupyter/conm/internal/secret"
)

var (
	_ Connections = (*ConnectionRepository)(nil)
	_ SecretUser  = (*ConnectionRepository)(nil)
)

func NewPostgresRepository(cfg config.Conm) (*ConnectionRepository, error) {
	file, err := config.OpenConfig[*config.PostgresConfigWrapper](config.PostgresPath())
	if err != nil {
		return nil, err
	}

	var sec secret.Provider = secret.Default()

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

	return &ConnectionRepository{
		cfg:  cfg,
		kind: config.PostgresConnType,
		file: file,
		mx:   &sync.RWMutex{},
		ncs:  ncs,
		sec:  sec,
	}, nil
}
