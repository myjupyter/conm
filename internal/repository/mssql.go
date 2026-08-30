package repository

import (
	"github.com/myjupyter/conm/internal/config"
)

func NewMSSQLRepository(cfg config.Conm) (*ConnectionRepository, error) {
	return openConnections[*config.MSSQLConfigWrapper](
		cfg,
		config.MSSQLConnType,
		config.MSSQLPath(),
	)
}
