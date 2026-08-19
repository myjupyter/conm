package spec

import (
	"strings"

	"github.com/myjupyter/conm/internal/config"
)

// KeyringFormFields is the order in which the fields are displayed
var KeyringFormFields = []FormField{
	// Secret information
	{
		Key:          strings.ToLower(config.KeyringFormFieldLocation),
		Label:        config.KeyringFormFieldLocation,
		Kind:         TextFieldKind,
		Example:      "/path/to/keyring/secret",
		DefaultValue: "",
		ValidateFunc: config.ValidateSecretLocation,
	},
	{
		Key:          strings.ToLower(config.KeyringFormFieldDescription),
		Label:        config.KeyringFormFieldDescription,
		Kind:         TextFieldKind,
		Example:      "",
		DefaultValue: "",
		ValidateFunc: config.ValidateSecretDescription,
	},
}

var KeyringFormSpec = FormSpec[config.Keyring]{
	AddTitle:  "Add a new Postgres connection",
	EditTitle: "Edit a Postgres connection",
	Fields:    PostgresFormFields,
	Sections: []FormSection{
		{
			Title: "secret",
			Note:  "secret location",
			Fields: []FormFieldKey{
				strings.ToLower(config.KeyringFormFieldLocation),
			},
		},
		{
			Title: "metadata",
			Note:  "secret metadata",
			Fields: []FormFieldKey{
				strings.ToLower(config.KeyringFormFieldDescription),
			},
		},
	},
	SeedFunc: func(k config.Keyring) map[FormFieldKey]FormFieldValue {
		return map[FormFieldKey]FormFieldValue{
			strings.ToLower(config.KeyringFormFieldLocation):    k.Location(),
			strings.ToLower(config.KeyringFormFieldDescription): k.Description(),
		}
	},
	BuildFunc: func(values map[FormFieldKey]FormFieldValue) (config.Keyring, error) {
		return config.Keyring{
			LocationRef: values[strings.ToLower(config.KeyringFormFieldLocation)],
			Desc:        values[strings.ToLower(config.KeyringFormFieldDescription)],
		}, nil
	},
}
