package repository

import (
	"github.com/myjupyter/conm/internal/config"
)

func NewRedisRepository(cfg config.Conm) (*ConnectionRepository, error) {
	return openConnections[*config.RedisConfigWrapper](
		cfg,
		config.RedisConnType,
		config.RedisPath(),
	)
}
