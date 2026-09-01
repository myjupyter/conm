package repository

import (
	"github.com/myjupyter/conm/internal/config"
)

func NewMongoDBRepository(cfg config.Conm) (*ConnectionRepository, error) {
	return openConnections[*config.MongoDBConfigWrapper](
		cfg,
		config.MongoDBConnType,
		config.MongoDBPath(),
	)
}
