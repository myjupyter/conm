package spec

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/myjupyter/conm/internal/config"
)

const (
	clickhouseExampleHost        = "localhost"
	clickhouseExamplePort        = "9000"
	clickhouseExampleUsername    = "default"
	clickhouseExampleDatabase    = "default"
	clickhouseExampleName        = "analytics-primary"
	clickhouseExampleDescription = "Events warehouse"
	clickhouseExampleTags        = "prod,eu-west,analytics"
)

const clickhouseDefaultPort = 9000

const (
	clickhouseSecureEnabled  FormFieldValue = "true"
	clickhouseSecureDisabled FormFieldValue = "false"
)

type clickhouseFormField = string

const (
	clickhouseFormFieldName        clickhouseFormField = "Name"
	clickhouseFormFieldDescription clickhouseFormField = "Description"
	clickhouseFormFieldTags        clickhouseFormField = "Tags"
	clickhouseFormFieldHost        clickhouseFormField = "Host"
	clickhouseFormFieldPort        clickhouseFormField = "Port"
	clickhouseFormFieldUsername    clickhouseFormField = "Username"
	clickhouseFormFieldDatabase    clickhouseFormField = "Database"
	clickhouseFormFieldSecure      clickhouseFormField = "Secure"
)

var ClickHouseFormFields = []FormField{
	{
		Key:          strings.ToLower(clickhouseFormFieldHost),
		Label:        clickhouseFormFieldHost,
		Example:      clickhouseExampleHost,
		ValidateFunc: config.ValidateClickHouseHost,
	},
	{
		Key:          strings.ToLower(clickhouseFormFieldPort),
		Label:        clickhouseFormFieldPort,
		Example:      clickhouseExamplePort,
		Kind:         IntFieldKind,
		Property:     OptionalFieldProperty,
		DefaultValue: strconv.Itoa(clickhouseDefaultPort),
		ValidateFunc: func(value string) error {
			if value == "" {
				return nil
			}
			port, err := strconv.Atoi(value)
			if err != nil {
				return errors.New("port must be a number")
			}
			return config.ValidateClickHousePort(port)
		},
	},
	{
		Key:          strings.ToLower(clickhouseFormFieldUsername),
		Label:        clickhouseFormFieldUsername,
		Example:      clickhouseExampleUsername,
		DefaultValue: clickhouseExampleUsername,
		ValidateFunc: config.ValidateClickHouseUsername,
	},
	secretProviderField,
	passwordField,
	{
		Key:          strings.ToLower(clickhouseFormFieldDatabase),
		Label:        clickhouseFormFieldDatabase,
		Example:      clickhouseExampleDatabase,
		DefaultValue: clickhouseExampleDatabase,
		ValidateFunc: config.ValidateClickHouseDatabase,
	},
	{
		Key:          strings.ToLower(clickhouseFormFieldSecure),
		Label:        clickhouseFormFieldSecure,
		Kind:         SelectFieldKind,
		DefaultValue: clickhouseSecureDisabled,
		Options:      []string{clickhouseSecureDisabled, clickhouseSecureEnabled},
	},
	{
		Key:          strings.ToLower(clickhouseFormFieldName),
		Label:        clickhouseFormFieldName,
		Example:      clickhouseExampleName,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateClickHouseName,
	},
	{
		Key:          strings.ToLower(clickhouseFormFieldDescription),
		Label:        clickhouseFormFieldDescription,
		Example:      clickhouseExampleDescription,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateClickHouseDescription,
	},
	{
		Key:      strings.ToLower(clickhouseFormFieldTags),
		Label:    clickhouseFormFieldTags,
		Example:  clickhouseExampleTags,
		Property: OptionalFieldProperty,
		ValidateFunc: func(value string) error {
			if len(value) > tagsMaxLength {
				return fmt.Errorf("tags must be at most %d characters long", tagsMaxLength)
			}
			return nil
		},
	},
}

var ClickHouseFormSpec = FormSpec[config.Connection]{
	AddTitle:  "Add a new ClickHouse connection",
	EditTitle: "Edit a ClickHouse connection",
	Fields:    ClickHouseFormFields,
	Sections: []FormSection{
		{
			Title: connectionSectionTitle,
			Note:  connectionSectionNote,
			Fields: []FormFieldKey{
				strings.ToLower(clickhouseFormFieldHost),
				strings.ToLower(clickhouseFormFieldPort),
				strings.ToLower(clickhouseFormFieldUsername),
				SecretProviderKey,
				SecretValueKey,
				strings.ToLower(clickhouseFormFieldDatabase),
				strings.ToLower(clickhouseFormFieldSecure),
			},
		},
		metadataSection(
			strings.ToLower(clickhouseFormFieldName),
			strings.ToLower(clickhouseFormFieldDescription),
			strings.ToLower(clickhouseFormFieldTags),
		),
	},
	SeedFunc: func(c config.Connection) map[FormFieldKey]FormFieldValue {
		ch, ok := c.(config.ClickHouse)
		if !ok {
			return nil
		}
		mode, value := SplitSecret(ch.Password)
		return map[FormFieldKey]FormFieldValue{
			strings.ToLower(clickhouseFormFieldHost):        ch.Hostname,
			strings.ToLower(clickhouseFormFieldPort):        strconv.Itoa(ch.PortNumber),
			strings.ToLower(clickhouseFormFieldUsername):    ch.User,
			SecretProviderKey:                               mode,
			SecretValueKey:                                  value,
			strings.ToLower(clickhouseFormFieldDatabase):    ch.DBName,
			strings.ToLower(clickhouseFormFieldSecure):      strconv.FormatBool(ch.Secure),
			strings.ToLower(clickhouseFormFieldName):        ch.Meta.Name,
			strings.ToLower(clickhouseFormFieldDescription): ch.Meta.Description,
			strings.ToLower(clickhouseFormFieldTags):        strings.Join(ch.Meta.Tags, ", "),
		}
	},
	BuildFunc: func(values map[FormFieldKey]FormFieldValue) (config.Connection, error) {
		port, err := intValue(values, strings.ToLower(clickhouseFormFieldPort), clickhouseDefaultPort)
		if err != nil {
			return nil, err
		}

		return config.ClickHouse{
			Meta: buildMeta(
				values,
				strings.ToLower(clickhouseFormFieldName),
				strings.ToLower(clickhouseFormFieldDescription),
				strings.ToLower(clickhouseFormFieldTags),
			),
			Hostname:   values[strings.ToLower(clickhouseFormFieldHost)],
			PortNumber: port,
			User:       values[strings.ToLower(clickhouseFormFieldUsername)],
			Password: JoinSecret(
				values[SecretProviderKey],
				values[SecretValueKey],
			),
			DBName: values[strings.ToLower(clickhouseFormFieldDatabase)],
			Secure: values[strings.ToLower(clickhouseFormFieldSecure)] == clickhouseSecureEnabled,
		}, nil
	},
}
