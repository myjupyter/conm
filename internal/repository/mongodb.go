package repository

import (
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
	"github.com/myjupyter/conm/internal/transport"
)

func NewMongoDBRepository(
	cfg config.Conm,
	transports transport.Provider,
	secrets secret.Provider,
	usage UsageLookup,
) (*ConnectionRepository, error) {
	return openConnections[*config.MongoDBConfigWrapper](
		cfg,
		config.MongoDBConnType,
		config.MongoDBPath(),
		transports,
		secrets,
		usage,
	)
}
