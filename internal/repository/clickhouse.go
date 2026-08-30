package repository

import (
	"github.com/myjupyter/conm/internal/config"
)

func NewClickHouseRepository(cfg config.Conm) (*ConnectionRepository, error) {
	return openConnections[*config.ClickHouseConfigWrapper](
		cfg,
		config.ClickHouseConnType,
		config.ClickHousePath(),
	)
}
