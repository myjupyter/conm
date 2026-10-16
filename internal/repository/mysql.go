package repository

import (
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/transport"
)

func NewMySQLRepository(
	cfg config.Conm,
	transports transport.Provider,
	usage UsageLookup,
) (*ConnectionRepository, error) {
	return openConnections[*config.MySQLConfigWrapper](
		cfg,
		config.MySQLConnType,
		config.MySQLPath(),
		transports,
		usage,
	)
}
