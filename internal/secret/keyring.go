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

type KeyringProvider struct {
	usages usageSet
}

func (*KeyringProvider) Scheme() (Scheme, SchemeMethods) {
	return Keyring, Resolve | Store | Remove | Usages
}

func (*KeyringProvider) Resolve(_ context.Context, spec Secretable) (string, error) {
	service, account := parseKeyringSpec(keyringSpec(spec))
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

func (p *KeyringProvider) Store(_ context.Context, spec Secretable, password string) error {
	raw := keyringSpec(spec)

	service, account := parseKeyringSpec(raw)
	if account == "" {
		return errEmptyKeyringAccount
	}

	if err := keyring.Set(service, account, password); err != nil {
		return err
	}

	p.usages.track(raw)

	return nil
}

func (p *KeyringProvider) Remove(_ context.Context, spec Secretable) error {
	raw := keyringSpec(spec)

	service, account := parseKeyringSpec(raw)
	if account == "" {
		return errEmptyKeyringAccount
	}

	p.usages.untrack(raw)

	if _, shared := p.usages.refs[raw]; shared {
		return nil
	}

	err := keyring.Delete(service, account)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil
	}
	return err
}

func (p *KeyringProvider) Track(spec Secretable) { p.usages.track(keyringSpec(spec)) }

func (p *KeyringProvider) Untrack(spec Secretable) { p.usages.untrack(keyringSpec(spec)) }

func (p *KeyringProvider) Usages(_ Scheme) []Usage { return p.usages.list() }

func keyringSpec(spec Secretable) string {
	return strings.TrimPrefix(spec.Secret(), string(Keyring)+":")
}

func parseKeyringSpec(spec string) (service, account string) {
	service, account, found := strings.Cut(spec, "/")
	if found {
		return service, account
	}
	return keyringService, spec
}
