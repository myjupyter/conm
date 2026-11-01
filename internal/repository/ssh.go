package repository

import (
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
	"github.com/myjupyter/conm/internal/transport"
)

func NewSSHRepository(
	cfg config.Conm,
	transports transport.Provider,
	secrets secret.Provider,
	usage UsageLookup,
) (*ConnectionRepository, error) {
	return openConnections[*config.SSHConfigWrapper](
		cfg,
		config.SSHConnType,
		config.SSHPath(),
		transports,
		secrets,
		usage,
	)
}
