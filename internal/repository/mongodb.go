package repository

import (
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/transport"
)

func NewMongoDBRepository(
	cfg config.Conm,
	transports transport.Provider,
	usage UsageLookup,
) (*ConnectionRepository, error) {
	return openConnections[*config.MongoDBConfigWrapper](
		cfg,
		config.MongoDBConnType,
		config.MongoDBPath(),
		transports,
		usage,
	)
}
