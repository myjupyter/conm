package spec

import (
	"slices"

	"github.com/myjupyter/conm/internal/secret"
)

type secretNestedFormFieldKey = string

const examplePassword = "4^@CP^S8\\le9"

const (
	secretNestedFormFieldSecretProvider secretNestedFormFieldKey = "provider"
	secretNestedFormFieldPassword       secretNestedFormFieldKey = "password"
)

// SecretProviderKey and SecretValueKey name the pair of fields that together
// say where a connection's password comes from: the provider picks literal,
// none or a store, and the value is either the password itself, nothing at all
// or a location inside that store. A spec without them is literal-only.
const (
	SecretProviderKey FormFieldKey = secretNestedFormFieldSecretProvider
	SecretValueKey    FormFieldKey = secretNestedFormFieldPassword
)

var SecretProvidersOrder = []string{
	secret.Literal,
	secret.None,
	secret.Keyring,
}

var (
	secretProviderField = FormField{
		Key:          secretNestedFormFieldSecretProvider,
		Label:        secretNestedFormFieldSecretProvider,
		Kind:         SelectFieldKind,
		DefaultValue: secret.Literal,
		Options:      SecretProvidersOrder,
	}

	passwordField = FormField{
		Key:      SecretValueKey,
		Label:    postgresFormFieldPassword,
		Kind:     HiddenFieldKind,
		Example:  examplePassword,
		Property: OptionalFieldProperty,
		ValidateFunc: func(value string) error {
			return nil
		},
	}
)

func JoinSecret(mode, value string) string {
	switch {
	case mode == secret.None:
		return secret.Ref(secret.None, "")
	case mode == "" || mode == secret.Literal || value == "":
		return value
	default:
		return secret.Ref(mode, value)
	}
}
func SplitSecret(raw string) (mode, value string) {
	scheme, location, ok := secret.ParseRef(raw)
	if !ok {
		return secret.Literal, raw
	}
	if scheme == secret.Literal {
		return secret.Literal, location
	}
	if slices.Contains(SecretProvidersOrder, scheme) {
		return scheme, location
	}
	return secret.Literal, raw
}
