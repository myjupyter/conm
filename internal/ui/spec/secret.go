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
// say where a connection's password comes from: the provider picks literal or
// a store, and the value is either the password itself or a location inside
// that store. A spec without them is literal-only.
const (
	SecretProviderKey FormFieldKey = secretNestedFormFieldSecretProvider
	SecretValueKey    FormFieldKey = secretNestedFormFieldPassword
)

var SecretProvidersOrder = []string{
	secret.Literal,
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
			// In a store mode this field holds a location, and the form
			// validates it against the entries that actually exist.
			return nil
		},
	}
)

// JoinSecret folds a mode and a value back into the reference stored on the
// connection. A literal is kept bare — that is what the resolver falls back to,
// so existing configs keep working untouched.
func JoinSecret(mode, value string) string {
	if mode == "" || mode == secret.Literal || value == "" {
		return value
	}
	return secret.Ref(mode, value)
}

// SplitSecret is JoinSecret's inverse. Only a known store counts as a
// reference: a bare password that happens to contain a colon is still a
// password, which is exactly how secret.Resolver reads it.
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
