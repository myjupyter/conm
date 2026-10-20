package spec

import (
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/i18n"
	"github.com/myjupyter/conm/internal/secret"
)

const (
	keyringExampleID       = "prod-orders-rw"
	keyringExampleLocation = "conm/prod-orders-rw"
	keyringExamplePassword = "hunter2hunter2"
)

type keyringFormField = string

const (
	keyringFormFieldID          keyringFormField = "id"
	keyringFormFieldLocation    keyringFormField = "location"
	keyringFormFieldPassword    keyringFormField = "password"
	keyringFormFieldDescription keyringFormField = "description"
)

// KeyringIDKey and KeyringPasswordKey name the two fields the keyring screens
// read back off the form: the entry's id, and the password typed for it.
const (
	KeyringIDKey       FormFieldKey = keyringFormFieldID
	KeyringPasswordKey FormFieldKey = keyringFormFieldPassword
)

var KeyringFormFields = []FormField{
	{
		Key:          keyringFormFieldID,
		Label:        i18n.Sprintf("id"),
		Example:      keyringExampleID,
		ValidateFunc: config.ValidateSecretID,
	},
	{
		Key:          keyringFormFieldLocation,
		Label:        i18n.Sprintf("location"),
		Example:      keyringExampleLocation,
		ValidateFunc: config.ValidateSecretLocation,
	},
	{
		Key:      keyringFormFieldPassword,
		Label:    i18n.Sprintf("password"),
		Kind:     HiddenFieldKind,
		Example:  keyringExamplePassword,
		Property: OptionalFieldProperty,
		ValidateFunc: func(string) error {
			return nil // empty keeps whatever the keyring already holds
		},
	},
	{
		Key:          keyringFormFieldDescription,
		Label:        i18n.Sprintf("description"),
		Example:      i18n.Sprintf("Order write path credentials"),
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateSecretDescription,
	},
}

// KeyringFormSpec drives both the add and the edit keyring form. It is a single
// section: an entry is four fields, too few to be worth splitting into tabs.
var KeyringFormSpec = FormSpec[config.Secret]{
	AddTitle:  i18n.Sprintf("new entry"),
	EditTitle: i18n.Sprintf("editing an entry"),
	Fields:    KeyringFormFields,
	Sections: []FormSection{
		{
			Title: i18n.Sprintf("secret"),
			Note:  i18n.Sprintf("where the password lives · conm never writes it to disk"),
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
