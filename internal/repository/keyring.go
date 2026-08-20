package repository

import (
	"github.com/myjupyter/conm/internal/config"
)

var _ Secrets[config.Keyring] = (*SecretRepository[config.Keyring])(nil)

func NewKeyringRepository(usage UsageLookup) (*SecretRepository[config.Keyring], error) {
	file, err := config.OpenConfig[*config.KeyringConfigWrapper](config.SecretConfigPath())
	if err != nil {
		return nil, err
	}

	return &SecretRepository[config.Keyring]{
		file:  file,
		usage: usage,
	}, nil
}
