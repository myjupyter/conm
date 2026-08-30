package repository

import (
	"github.com/myjupyter/conm/internal/config"
)

var (
	_ Connections = (*ConnectionRepository)(nil)
	_ SecretUser  = (*ConnectionRepository)(nil)
)

func NewPostgresRepository(cfg config.Conm) (*ConnectionRepository, error) {
	return openConnections[*config.PostgresConfigWrapper](
		cfg,
		config.PostgresConnType,
		config.PostgresPath(),
	)
}
