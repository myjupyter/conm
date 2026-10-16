package repository

import (
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/transport"
)

func NewRedisRepository(
	cfg config.Conm,
	transports transport.Provider,
	usage UsageLookup,
) (*ConnectionRepository, error) {
	return openConnections[*config.RedisConfigWrapper](
		cfg,
		config.RedisConnType,
		config.RedisPath(),
		transports,
		usage,
	)
}
