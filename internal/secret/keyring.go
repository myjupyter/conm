package secret

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/zalando/go-keyring"
)

var keyringService = "conm"

func UseKeyringService(service string) {
	keyringService = service
}

var errEmptyKeyringAccount = errors.New("keyring spec must be [service/]account")

var _ Provider = (*KeyringProvider)(nil)

type KeyringProvider struct {
	Service string
}

func (*KeyringProvider) Scheme() Scheme { return Keyring }

func (p *KeyringProvider) Resolve(_ context.Context, ref Reference) (string, error) {
	service, account := p.parse(ref)
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
	service, account := p.parse(ref)
	if account == "" {
		return errEmptyKeyringAccount
	}

	if err := keyring.Set(service, account, password); err != nil {
		return err
	}

	return nil
}

func (p *KeyringProvider) Remove(_ context.Context, ref Reference) error {
	service, account := p.parse(ref)
	if account == "" {
		return errEmptyKeyringAccount
	}

	err := keyring.Delete(service, account)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}

func NamesKeyringService(ref Reference) bool {
	return strings.Contains(keyringSpec(ref), "/")
}

func (p *KeyringProvider) parse(ref Reference) (service, account string) {
	service, account = parseKeyringSpec(keyringSpec(ref))
	if p.Service != "" && !NamesKeyringService(ref) {
		service = p.Service
	}
	return service, account
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
