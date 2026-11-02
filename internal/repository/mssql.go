package repository

import (
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
	"github.com/myjupyter/conm/internal/transport"
)

func NewMSSQLRepository(
	cfg config.Conm,
	transports transport.Provider,
	secrets secret.Provider,
	usage UsageLookup,
) (*ConnectionRepository, error) {
	return openConnections[*config.MSSQLConfigWrapper](
		cfg,
		config.MSSQLConnType,
		config.MSSQLPath(),
		transports,
		secrets,
		usage,
	)
}
