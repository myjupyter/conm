package repository

import (
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/transport"
)

func NewSSHRepository(
	cfg config.Conm,
	transports transport.Provider,
	usage UsageLookup,
) (*ConnectionRepository, error) {
	return openConnections[*config.SSHConfigWrapper](
		cfg,
		config.SSHConnType,
		config.SSHPath(),
		transports,
		usage,
	)
}
