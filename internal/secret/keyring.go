package secret

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/zalando/go-keyring"
)

var (
	errEmptyKeyringAccount = errors.New("keyring spec must be [service/]account")
	errNoKeyringWorkspace  = errors.New("keyring provider has no workspace to store secrets in")
)

var _ Provider = (*KeyringProvider)(nil)

type KeyringProvider struct {
	Workspace string
}

func (*KeyringProvider) Scheme() Scheme { return Keyring }

func (p *KeyringProvider) Resolve(_ context.Context, ref Reference) (string, error) {
	service, account, err := p.parse(ref)
	if err != nil {
		return "", err
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
	service, account, err := p.parse(ref)
	if err != nil {
		return err
	}

	if err := keyring.Set(service, account, password); err != nil {
		return err
	}

	return nil
}

func (p *KeyringProvider) Remove(_ context.Context, ref Reference) error {
	service, account, err := p.parse(ref)
	if err != nil {
		return err
	}

	err = keyring.Delete(service, account)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}

func (p *KeyringProvider) parse(ref Reference) (service, account string, err error) {
	service, account = parseKeyringSpec(keyringSpec(ref))
	if service == "" {
		service = p.Workspace
	}

	switch {
	case account == "":
		return "", "", errEmptyKeyringAccount
	case service == "":
		return "", "", errNoKeyringWorkspace
	}
	return service, account, nil
}

func keyringSpec(ref Reference) string {
	return strings.TrimPrefix(ref.SecretRef(), Keyring+":")
}

func parseKeyringSpec(spec string) (service, account string) {
	service, account, found := strings.Cut(spec, "/")
	if found {
		return service, account
	}
	return "", spec
}
