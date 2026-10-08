package repository

import (
	"github.com/myjupyter/conm/internal/config"
)

func NewSSHRepository(cfg config.Conm) (*ConnectionRepository, error) {
	return openConnections[*config.SSHConfigWrapper](
		cfg,
		config.SSHConnType,
		config.SSHPath(),
	)
}
