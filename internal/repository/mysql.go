package repository

import (
	"github.com/myjupyter/conm/internal/config"
)

func NewMySQLRepository(cfg config.Conm) (*ConnectionRepository, error) {
	return openConnections[*config.MySQLConfigWrapper](
		cfg,
		config.MySQLConnType,
		config.MySQLPath(),
	)
}
