package repository

import (
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/transport"
)

func NewMSSQLRepository(
	cfg config.Conm,
	transports transport.Provider,
	usage UsageLookup,
) (*ConnectionRepository, error) {
	return openConnections[*config.MSSQLConfigWrapper](
		cfg,
		config.MSSQLConnType,
		config.MSSQLPath(),
		transports,
		usage,
	)
}
