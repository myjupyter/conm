package config

import (
	"errors"
	"fmt"

	"github.com/pelletier/go-toml/v2"
)

var _ Secret = (*Keyring)(nil)
var _ ConfigWrapper[Secret] = (*KeyringConfigWrapper)(nil)

const (
	secretIDMaxLength          = 63
	secretDescriptionMaxLength = 255
)

type Keyring struct {
	SecretID    string `toml:"id"`
	ProviderID  string `toml:"provider"`
	LocationRef string `toml:"location"`
	Desc        string `toml:"description,omitempty"`

	validationErrs []error
}

type KeyringConfigWrapper struct {
	Secrets []Keyring `toml:"secret"`
}

func ValidateSecretID(id string) error {
	if id == "" {
		return errors.New("secret id can't be empty")
	}
	if len(id) > secretIDMaxLength {
		return fmt.Errorf("secret id must be at most %d characters long", secretIDMaxLength)
	}
	return nil
}

func ValidateSecretProvider(provider string) error {
	if provider == "" {
		return errors.New("secret provider can't be empty")
	}
	return nil
}

func ValidateSecretLocation(location string) error {
	// TODO: validate location properly
	if location == "" {
		return errors.New("secret location can't be empty")
	}
	return nil
}

func ValidateSecretDescription(desc string) error {
	if len(desc) > secretDescriptionMaxLength {
		return fmt.Errorf("description must be at most %d characters long", secretDescriptionMaxLength)
	}
	return nil
}

func (w *KeyringConfigWrapper) Add(spec Secret) {
	k, ok := spec.(Keyring)
	if !ok {
		return
	}

	w.Secrets = append(w.Secrets, k)
}

func (w *KeyringConfigWrapper) Validate() {
	for i := range w.Secrets {
		w.Secrets[i].validationErrs = w.Secrets[i].Validate()
	}
}

func (w *KeyringConfigWrapper) Unmarshal(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	return toml.Unmarshal(data, w)
}

func (w *KeyringConfigWrapper) Marshal() ([]byte, error) {
	return toml.Marshal(w)
}

func (w *KeyringConfigWrapper) Len() int {
	return len(w.Secrets)
}

func (w *KeyringConfigWrapper) Get(i int) Secret {
	return w.Secrets[i]
}

func (w *KeyringConfigWrapper) Put(i int, spec Secret) {
	k, ok := spec.(Keyring)
	if !ok {
		return
	}

	w.Secrets[i] = k
}

func (w *KeyringConfigWrapper) Remove(i int) {
	if i < 0 || i >= len(w.Secrets) {
		return
	}
	w.Secrets = append(w.Secrets[:i], w.Secrets[i+1:]...)
}

func (k Keyring) ID() string {
	return k.SecretID
}

func (k Keyring) Provider() string {
	return k.ProviderID
}

func (k Keyring) Location() string {
	return k.LocationRef
}

func (k Keyring) Description() string {
	return k.Desc
}

func (k Keyring) Params() []SecretParam {
	return nil
}

func (k Keyring) Validate() []error {
	var errs []error
	if err := ValidateSecretID(k.SecretID); err != nil {
		errs = append(errs, err)
	}
	if err := ValidateSecretProvider(k.ProviderID); err != nil {
		errs = append(errs, err)
	}
	if err := ValidateSecretLocation(k.LocationRef); err != nil {
		errs = append(errs, err)
	}
	if err := ValidateSecretDescription(k.Desc); err != nil {
		errs = append(errs, err)
	}
	return errs
}

func (k Keyring) IsValid() bool {
	return len(k.validationErrs) == 0
}
