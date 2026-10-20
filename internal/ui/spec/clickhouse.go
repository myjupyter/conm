package spec

import (
	"errors"
	"strconv"
	"strings"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/i18n"
)

const (
	clickhouseExampleHost     = "localhost"
	clickhouseExamplePort     = "9000"
	clickhouseExampleUsername = "default"
	clickhouseExampleDatabase = "default"
	clickhouseExampleName     = "analytics-primary"
	clickhouseExampleTags     = "prod,eu-west,analytics"
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
		Label:        i18n.Sprintf("Host"),
		Example:      clickhouseExampleHost,
		ValidateFunc: config.ValidateClickHouseHost,
	},
	{
		Key:          strings.ToLower(clickhouseFormFieldPort),
		Label:        i18n.Sprintf("Port"),
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
				return errors.New(i18n.Sprintf("port must be a number"))
			}
			return config.ValidateClickHousePort(port)
		},
	},
	{
		Key:          strings.ToLower(clickhouseFormFieldUsername),
		Label:        i18n.Sprintf("Username"),
		Example:      clickhouseExampleUsername,
		DefaultValue: clickhouseExampleUsername,
		ValidateFunc: config.ValidateClickHouseUsername,
	},
	secretProviderField,
	passwordField,
	{
		Key:          strings.ToLower(clickhouseFormFieldDatabase),
		Label:        i18n.Sprintf("Database"),
		Example:      clickhouseExampleDatabase,
		DefaultValue: clickhouseExampleDatabase,
		ValidateFunc: config.ValidateClickHouseDatabase,
	},
	{
		Key:          strings.ToLower(clickhouseFormFieldSecure),
		Label:        i18n.Sprintf("Secure"),
		Kind:         SelectFieldKind,
		DefaultValue: clickhouseSecureDisabled,
		Options:      []string{clickhouseSecureDisabled, clickhouseSecureEnabled},
	},
	{
		Key:          strings.ToLower(clickhouseFormFieldName),
		Label:        i18n.Sprintf("Name"),
		Example:      clickhouseExampleName,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateClickHouseName,
	},
	{
		Key:          strings.ToLower(clickhouseFormFieldDescription),
		Label:        i18n.Sprintf("Description"),
		Example:      i18n.Sprintf("Events warehouse"),
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateClickHouseDescription,
	},
	{
		Key:      strings.ToLower(clickhouseFormFieldTags),
		Label:    i18n.Sprintf("Tags"),
		Example:  clickhouseExampleTags,
		Property: OptionalFieldProperty,
		ValidateFunc: func(value string) error {
			if len(value) > tagsMaxLength {
				return errors.New(i18n.Sprintf("tags must be at most %d characters long", tagsMaxLength))
			}
			return nil
		},
	},
	transportTypeField,
	transportField,
}

var ClickHouseFormSpec = FormSpec[config.Connection]{
	AddTitle:  i18n.Sprintf("Add a new ClickHouse connection"),
	EditTitle: i18n.Sprintf("Edit a ClickHouse connection"),
	Fields:    ClickHouseFormFields,
	Sections: []FormSection{
		{
			Title: connectionSectionTitle,
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
		transportSection,
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
			TransportTypeKey:                                transportTypeOf(ch.Transport),
			TransportKey:                                    transportNameOf(ch.Transport),
			strings.ToLower(clickhouseFormFieldDatabase):    ch.DBName,
			strings.ToLower(clickhouseFormFieldSecure):      strconv.FormatBool(ch.Secure),
			strings.ToLower(clickhouseFormFieldName):        ch.Metadata.Name,
			strings.ToLower(clickhouseFormFieldDescription): ch.Metadata.Description,
			strings.ToLower(clickhouseFormFieldTags):        strings.Join(ch.Metadata.Tags, ", "),
		}
	},
	BuildFunc: func(values map[FormFieldKey]FormFieldValue) (config.Connection, error) {
		port, err := intValue(values, strings.ToLower(clickhouseFormFieldPort), clickhouseDefaultPort)
		if err != nil {
			return nil, err
		}

		return config.ClickHouse{
			Transport: JoinTransport(values[TransportTypeKey], values[TransportKey]),
			Metadata: buildMeta(
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
