package spec

import (
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

const (
	keyringExampleID          = "prod-orders-rw"
	keyringExampleLocation    = "prod/orders-rw"
	keyringExamplePassword    = "hunter2hunter2"
	keyringExampleDescription = "Order write path credentials"
)

type keyringFormField = string

const (
	keyringFormFieldID          keyringFormField = "id"
	keyringFormFieldLocation    keyringFormField = "location"
	keyringFormFieldPassword    keyringFormField = "password"
	keyringFormFieldDescription keyringFormField = "description"
)

const (
	KeyringIDKey       FormFieldKey = keyringFormFieldID
	KeyringPasswordKey FormFieldKey = keyringFormFieldPassword
)

var KeyringFormFields = []FormField{
	{
		Key:          keyringFormFieldID,
		Label:        keyringFormFieldID,
		Example:      keyringExampleID,
		ValidateFunc: config.ValidateSecretID,
	},
	{
		Key:          keyringFormFieldLocation,
		Label:        keyringFormFieldLocation,
		Example:      keyringExampleLocation,
		ValidateFunc: config.ValidateSecretLocation,
	},
	{
		Key:      keyringFormFieldPassword,
		Label:    keyringFormFieldPassword,
		Kind:     HiddenFieldKind,
		Example:  keyringExamplePassword,
		Property: OptionalFieldProperty,
		ValidateFunc: func(string) error {
			return nil
		},
	},
	{
		Key:          keyringFormFieldDescription,
		Label:        keyringFormFieldDescription,
		Example:      keyringExampleDescription,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateSecretDescription,
	},
}

var KeyringFormSpec = FormSpec[config.Secret]{
	AddTitle:  "new entry",
	EditTitle: "editing an entry",
	Fields:    KeyringFormFields,
	Sections: []FormSection{
		{
			Title: "secret",
			Fields: []FormFieldKey{
				keyringFormFieldID,
				keyringFormFieldLocation,
				keyringFormFieldPassword,
				keyringFormFieldDescription,
			},
		},
	},
	SeedFunc: func(s config.Secret) map[FormFieldKey]FormFieldValue {
		k, ok := s.(config.Keyring)
		if !ok {
			return nil
		}
		return map[FormFieldKey]FormFieldValue{
			keyringFormFieldID:          k.SecretID,
			keyringFormFieldLocation:    k.LocationRef,
			keyringFormFieldPassword:    "",
			keyringFormFieldDescription: k.Desc,
		}
	},
	BuildFunc: func(values map[FormFieldKey]FormFieldValue) (config.Secret, error) {
		return config.Keyring{
			SecretID:    values[keyringFormFieldID],
			ProviderID:  secret.Keyring,
			LocationRef: values[keyringFormFieldLocation],
			Desc:        values[keyringFormFieldDescription],
		}, nil
	},
}
