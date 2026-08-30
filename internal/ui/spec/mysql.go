package spec

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/myjupyter/conm/internal/config"
)

const (
	mysqlExampleHost        = "localhost"
	mysqlExamplePort        = "3306"
	mysqlExampleUsername    = "root"
	mysqlExampleDatabase    = "app"
	mysqlExampleName        = "prod-primary"
	mysqlExampleDescription = "Production primary database"
	mysqlExampleTags        = "prod,eu-west,primary"
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
		Label:        mysqlFormFieldHost,
		Example:      mysqlExampleHost,
		ValidateFunc: config.ValidateMySQLHost,
	},
	{
		Key:          strings.ToLower(mysqlFormFieldPort),
		Label:        mysqlFormFieldPort,
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
				return errors.New("port must be a number")
			}
			return config.ValidateMySQLPort(port)
		},
	},
	{
		Key:          strings.ToLower(mysqlFormFieldUsername),
		Label:        mysqlFormFieldUsername,
		Example:      mysqlExampleUsername,
		ValidateFunc: config.ValidateMySQLUsername,
	},
	secretProviderField,
	passwordField,
	{
		Key:          strings.ToLower(mysqlFormFieldDatabase),
		Label:        mysqlFormFieldDatabase,
		Example:      mysqlExampleDatabase,
		ValidateFunc: config.ValidateMySQLDatabase,
	},
	{
		Key:          strings.ToLower(mysqlFormFieldTLSMode),
		Label:        mysqlFormFieldTLSMode,
		Kind:         SelectFieldKind,
		DefaultValue: config.MySQLTLSModePreferred,
		Options:      MySQLTLSModesOrder,
		ValidateFunc: config.ValidateMySQLTLSMode,
	},
	{
		Key:          strings.ToLower(mysqlFormFieldName),
		Label:        mysqlFormFieldName,
		Example:      mysqlExampleName,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateMySQLName,
	},
	{
		Key:          strings.ToLower(mysqlFormFieldDescription),
		Label:        mysqlFormFieldDescription,
		Example:      mysqlExampleDescription,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateMySQLDescription,
	},
	{
		Key:      strings.ToLower(mysqlFormFieldTags),
		Label:    mysqlFormFieldTags,
		Example:  mysqlExampleTags,
		Property: OptionalFieldProperty,
		ValidateFunc: func(value string) error {
			if len(value) > tagsMaxLength {
				return fmt.Errorf("tags must be at most %d characters long", tagsMaxLength)
			}
			return nil
		},
	},
}

var MySQLFormSpec = FormSpec[config.Connection]{
	AddTitle:  "Add a new MySQL connection",
	EditTitle: "Edit a MySQL connection",
	Fields:    MySQLFormFields,
	Sections: []FormSection{
		{
			Title: connectionSectionTitle,
			Note:  connectionSectionNote,
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
		{
			Title: metadataSectionTitle,
			Note:  metadataSectionNote,
			Fields: []FormFieldKey{
				strings.ToLower(mysqlFormFieldName),
				strings.ToLower(mysqlFormFieldDescription),
				strings.ToLower(mysqlFormFieldTags),
			},
		},
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
			strings.ToLower(mysqlFormFieldDatabase):    my.DBName,
			strings.ToLower(mysqlFormFieldTLSMode):     my.TLSMode,
			strings.ToLower(mysqlFormFieldName):        my.Meta.Name,
			strings.ToLower(mysqlFormFieldDescription): my.Meta.Description,
			strings.ToLower(mysqlFormFieldTags):        strings.Join(my.Meta.Tags, ", "),
		}
	},
	BuildFunc: func(values map[FormFieldKey]FormFieldValue) (config.Connection, error) {
		port := mysqlDefaultPort
		if rawPort, ok := values[strings.ToLower(mysqlFormFieldPort)]; ok && rawPort != "" {
			p, err := strconv.Atoi(rawPort)
			if err != nil {
				return nil, fmt.Errorf("invalid port value: %w", err)
			}
			port = p
		}

		return config.MySQL{
			Meta: config.ConnMeta{
				Name:        values[strings.ToLower(mysqlFormFieldName)],
				Description: values[strings.ToLower(mysqlFormFieldDescription)],
				Tags:        parseTags(values[strings.ToLower(mysqlFormFieldTags)]),
			},
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
