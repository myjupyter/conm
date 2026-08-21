package secret

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/zalando/go-keyring"
)

const keyringService = "conm"

var errEmptyKeyringAccount = errors.New("keyring spec must be [service/]account")

var _ Provider = (*KeyringProvider)(nil)

type KeyringProvider struct{}

func (*KeyringProvider) Scheme() Scheme { return Keyring }

func (*KeyringProvider) Resolve(_ context.Context, ref Reference) (string, error) {
	service, account := parseKeyringSpec(keyringSpec(ref))
	if account == "" {
		return "", errEmptyKeyringAccount
	}

	password, err := keyring.Get(service, account)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", fmt.Errorf("no secret stored for %q in %q", account, service)
		}
		return "", err
	}

	return password, nil
}

func (p *KeyringProvider) Store(_ context.Context, ref Reference, password string) error {
	raw := keyringSpec(ref)

	service, account := parseKeyringSpec(raw)
	if account == "" {
		return errEmptyKeyringAccount
	}

	if err := keyring.Set(service, account, password); err != nil {
		return err
	}

	return nil
}

func (p *KeyringProvider) Remove(_ context.Context, ref Reference) error {
	raw := keyringSpec(ref)

	service, account := parseKeyringSpec(raw)
	if account == "" {
		return errEmptyKeyringAccount
	}

	err := keyring.Delete(service, account)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}

func keyringSpec(ref Reference) string {
	return strings.TrimPrefix(ref.SecretRef(), Keyring+":")
}

func parseKeyringSpec(spec string) (service, account string) {
	service, account, found := strings.Cut(spec, "/")
	if found {
		return service, account
	}
	return keyringService, spec
}
