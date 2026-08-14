package registry

import (
	"fmt"
	"sync"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/network"
	"github.com/myjupyter/conm/internal/secret"
)

var _ Registry[config.Postgres] = (*CommonRegistry[config.Postgres])(nil)

func NewPostgresRegistry(cfg config.Conm) (*CommonRegistry[config.Postgres], error) {
	configPath, err := config.PGFilePath()
	if err != nil {
		return nil, err
	}

	file, err := config.OpenConfig[*config.PostgresConfigWrapper](configPath)
	if err != nil {
		return nil, err
	}

	sec := secret.Default()

	n := file.Len()
	ncs := make([]network.Connection, 0, n)
	for i := 0; i < n; i++ {
		c := file.Get(i)

		sec.Track(c)

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

	return &CommonRegistry[config.Postgres]{
		cfg:  cfg,
		file: file,
		mx:   &sync.RWMutex{},
		ncs:  ncs,
		sec:  sec,
	}, nil
}
