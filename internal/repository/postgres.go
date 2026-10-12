package repository

import (
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/transport"
)

var (
	_ Connections = (*ConnectionRepository)(nil)
	_ SecretUser  = (*ConnectionRepository)(nil)
)

func NewPostgresRepository(
	cfg config.Conm,
	transports transport.Provider,
	usage UsageLookup,
) (*ConnectionRepository, error) {
	return openConnections[*config.PostgresConfigWrapper](
		cfg,
		config.PostgresConnType,
		config.PostgresPath(),
		transports,
		usage,
	)
}
