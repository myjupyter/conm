package spec

import (
	"errors"
	"strconv"
	"strings"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/i18n"
)

const (
	mysqlExampleHost     = "localhost"
	mysqlExamplePort     = "3306"
	mysqlExampleUsername = "root"
	mysqlExampleDatabase = "app"
	mysqlExampleName     = "prod-primary"
	mysqlExampleTags     = "prod,eu-west,primary"
)

const mysqlDefaultPort = 3306

type mysqlFormField = string

const (
	mysqlFormFieldName        mysqlFormField = "Name"
	mysqlFormFieldDescription mysqlFormField = "Description"
	mysqlFormFieldTags        mysqlFormField = "Tags"
	mysqlFormFieldHost        mysqlFormField = "Host"
	mysqlFormFieldPort        mysqlFormField = "Port"
	mysqlFormFieldUsername    mysqlFormField = "Username"
	mysqlFormFieldDatabase    mysqlFormField = "Database"
	mysqlFormFieldTLSMode     mysqlFormField = "TLS"
)

var MySQLTLSModesOrder = []config.MySQLTLSMode{
	config.MySQLTLSModeDisable,
	config.MySQLTLSModePreferred,
	config.MySQLTLSModeSkipVerify,
	config.MySQLTLSModeVerify,
}

var MySQLFormFields = []FormField{
	{
		Key:          strings.ToLower(mysqlFormFieldHost),
		Label:        i18n.Sprintf("Host"),
		Example:      mysqlExampleHost,
		ValidateFunc: config.ValidateMySQLHost,
	},
	{
		Key:          strings.ToLower(mysqlFormFieldPort),
		Label:        i18n.Sprintf("Port"),
		Example:      mysqlExamplePort,
		Kind:         IntFieldKind,
		Property:     OptionalFieldProperty,
		DefaultValue: strconv.Itoa(mysqlDefaultPort),
		ValidateFunc: func(value string) error {
			if value == "" {
				return nil
			}
			port, err := strconv.Atoi(value)
			if err != nil {
				return errors.New(i18n.Sprintf("port must be a number"))
			}
			return config.ValidateMySQLPort(port)
		},
	},
	{
		Key:          strings.ToLower(mysqlFormFieldUsername),
		Label:        i18n.Sprintf("Username"),
		Example:      mysqlExampleUsername,
		ValidateFunc: config.ValidateMySQLUsername,
	},
	secretProviderField,
	passwordField,
	{
		Key:          strings.ToLower(mysqlFormFieldDatabase),
		Label:        i18n.Sprintf("Database"),
		Example:      mysqlExampleDatabase,
		ValidateFunc: config.ValidateMySQLDatabase,
	},
	{
		Key:          strings.ToLower(mysqlFormFieldTLSMode),
		Label:        i18n.Sprintf("TLS"),
		Kind:         SelectFieldKind,
		DefaultValue: config.MySQLTLSModePreferred,
		Options:      MySQLTLSModesOrder,
		ValidateFunc: config.ValidateMySQLTLSMode,
	},
	{
		Key:          strings.ToLower(mysqlFormFieldName),
		Label:        i18n.Sprintf("Name"),
		Example:      mysqlExampleName,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateMySQLName,
	},
	{
		Key:          strings.ToLower(mysqlFormFieldDescription),
		Label:        i18n.Sprintf("Description"),
		Example:      i18n.Sprintf("Production primary database"),
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateMySQLDescription,
	},
	{
		Key:      strings.ToLower(mysqlFormFieldTags),
		Label:    i18n.Sprintf("Tags"),
		Example:  mysqlExampleTags,
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

var MySQLFormSpec = FormSpec[config.Connection]{
	AddTitle:  i18n.Sprintf("Add a new MySQL connection"),
	EditTitle: i18n.Sprintf("Edit a MySQL connection"),
	Fields:    MySQLFormFields,
	Sections: []FormSection{
		{
			Title: connectionSectionTitle,
			Fields: []FormFieldKey{
				strings.ToLower(mysqlFormFieldHost),
				strings.ToLower(mysqlFormFieldPort),
				strings.ToLower(mysqlFormFieldUsername),
				SecretProviderKey,
				SecretValueKey,
				strings.ToLower(mysqlFormFieldDatabase),
				strings.ToLower(mysqlFormFieldTLSMode),
			},
		},
		transportSection,
		metadataSection(
			strings.ToLower(mysqlFormFieldName),
			strings.ToLower(mysqlFormFieldDescription),
			strings.ToLower(mysqlFormFieldTags),
		),
	},
	SeedFunc: func(c config.Connection) map[FormFieldKey]FormFieldValue {
		my, ok := c.(config.MySQL)
		if !ok {
			return nil
		}
		mode, value := SplitSecret(my.Password)
		return map[FormFieldKey]FormFieldValue{
			strings.ToLower(mysqlFormFieldHost):        my.Hostname,
			strings.ToLower(mysqlFormFieldPort):        strconv.Itoa(my.PortNumber),
			strings.ToLower(mysqlFormFieldUsername):    my.User,
			SecretProviderKey:                          mode,
			SecretValueKey:                             value,
			TransportTypeKey:                           transportTypeOf(my.Transport),
			TransportKey:                               transportNameOf(my.Transport),
			strings.ToLower(mysqlFormFieldDatabase):    my.DBName,
			strings.ToLower(mysqlFormFieldTLSMode):     my.TLSMode,
			strings.ToLower(mysqlFormFieldName):        my.Metadata.Name,
			strings.ToLower(mysqlFormFieldDescription): my.Metadata.Description,
			strings.ToLower(mysqlFormFieldTags):        strings.Join(my.Metadata.Tags, ", "),
		}
	},
	BuildFunc: func(values map[FormFieldKey]FormFieldValue) (config.Connection, error) {
		port, err := intValue(values, strings.ToLower(mysqlFormFieldPort), mysqlDefaultPort)
		if err != nil {
			return nil, err
		}

		return config.MySQL{
			Transport: JoinTransport(values[TransportTypeKey], values[TransportKey]),
			Metadata: buildMeta(
				values,
				strings.ToLower(mysqlFormFieldName),
				strings.ToLower(mysqlFormFieldDescription),
				strings.ToLower(mysqlFormFieldTags),
			),
			Hostname:   values[strings.ToLower(mysqlFormFieldHost)],
			PortNumber: port,
			User:       values[strings.ToLower(mysqlFormFieldUsername)],
			Password: JoinSecret(
				values[SecretProviderKey],
				values[SecretValueKey],
			),
			DBName:  values[strings.ToLower(mysqlFormFieldDatabase)],
			TLSMode: values[strings.ToLower(mysqlFormFieldTLSMode)],
		}, nil
	},
}
