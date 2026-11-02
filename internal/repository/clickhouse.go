package repository

import (
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
	"github.com/myjupyter/conm/internal/transport"
)

func NewClickHouseRepository(
	cfg config.Conm,
	transports transport.Provider,
	secrets secret.Provider,
	usage UsageLookup,
) (*ConnectionRepository, error) {
	return openConnections[*config.ClickHouseConfigWrapper](
		cfg,
		config.ClickHouseConnType,
		config.ClickHousePath(),
		transports,
		secrets,
		usage,
	)
}
