package config

import (
	"fmt"

	"github.com/pelletier/go-toml/v2"
)

var _ Secret = SecretSpec{}
var _ ConfigWrapper[SecretSpec] = &SecretConfigWrapper{}

const (
	secretIDMaxLength          = 63
	secretDescriptionMaxLength = 255
)

type Secret interface {
	ID() string
	Provider() string
	Location() string
	Description() string

	IsValid() bool
	Validate() []error
}

type SecretSpec struct {
	SecretID    string `toml:"id"`
	ProviderID  string `toml:"provider"`
	LocationRef string `toml:"location"`
	Desc        string `toml:"description,omitempty"`

	validationErrs []error
}

type SecretConfigWrapper struct {
	Secrets []SecretSpec `toml:"secret"`
}

func ValidateSecretID(id string) error {
	if id == "" {
		return fmt.Errorf("secret id can't be empty")
	}
	if len(id) > secretIDMaxLength {
		return fmt.Errorf("secret id must be at most %d characters long", secretIDMaxLength)
	}
	return nil
}

func ValidateSecretProvider(provider string) error {
	if provider == "" {
		return fmt.Errorf("secret provider can't be empty")
	}
	return nil
}

func ValidateSecretLocation(location string) error {
	if location == "" {
		return fmt.Errorf("secret location can't be empty")
	}
	return nil
}

func ValidateSecretDescription(desc string) error {
	if len(desc) > secretDescriptionMaxLength {
		return fmt.Errorf("description must be at most %d characters long", secretDescriptionMaxLength)
	}
	return nil
}

func (w *SecretConfigWrapper) Add(spec SecretSpec) {
	w.Secrets = append(w.Secrets, spec)
}

func (w *SecretConfigWrapper) Validate() {
	for i := range w.Secrets {
		w.Secrets[i].validationErrs = w.Secrets[i].Validate()
	}
}

func (w *SecretConfigWrapper) Unmarshal(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	return toml.Unmarshal(data, w)
}

func (w *SecretConfigWrapper) Marshal() ([]byte, error) {
	return toml.Marshal(w)
}

func (w *SecretConfigWrapper) Len() int {
	return len(w.Secrets)
}

func (w *SecretConfigWrapper) Get(i int) SecretSpec {
	return w.Secrets[i]
}

func (w *SecretConfigWrapper) Put(i int, spec SecretSpec) {
	w.Secrets[i] = spec
}

func (w *SecretConfigWrapper) Remove(i int) {
	if i < 0 || i >= len(w.Secrets) {
		return
	}
	w.Secrets = append(w.Secrets[:i], w.Secrets[i+1:]...)
}

func (s SecretSpec) ID() string {
	return s.SecretID
}

func (s SecretSpec) Provider() string {
	return s.ProviderID
}

func (s SecretSpec) Location() string {
	return s.LocationRef
}

func (s SecretSpec) Description() string {
	return s.Desc
}

func (s SecretSpec) Validate() []error {
	var errs []error
	if err := ValidateSecretID(s.SecretID); err != nil {
		errs = append(errs, err)
	}
	if err := ValidateSecretProvider(s.ProviderID); err != nil {
		errs = append(errs, err)
	}
	if err := ValidateSecretLocation(s.LocationRef); err != nil {
		errs = append(errs, err)
	}
	if err := ValidateSecretDescription(s.Desc); err != nil {
		errs = append(errs, err)
	}
	return errs
}

func (s SecretSpec) IsValid() bool {
	return len(s.validationErrs) == 0
}
