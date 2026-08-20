package repository

import (
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

var _ Secrets = (*SecretRepository)(nil)

func NewKeyringRepository(usage UsageLookup) (*SecretRepository, error) {
	file, err := config.OpenConfig[*config.KeyringConfigWrapper](config.SecretConfigPath())
	if err != nil {
		return nil, err
	}

	return newSecretRepository(&secret.KeyringProvider{}, file, usage)
}
