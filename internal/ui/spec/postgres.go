package spec

import (
	"errors"
	"strconv"
	"strings"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/i18n"
)

const (
	postgresExampleHost     = "localhost"
	postgresExamplePort     = "5432"
	postgresExampleUsername = "postgres"
	postgresExampleDatabase = "master"
	postgresExampleSchema   = "public"
	postgresExampleName     = "prod-primary"
	postgresExampleTags     = "prod,eu-west,primary"
)

type postgresFormField = string

const (
	postgresFormFieldName        postgresFormField = "Name"
	postgresFormFieldDescription postgresFormField = "Description"
	postgresFormFieldTags        postgresFormField = "Tags"
	postgresFormFieldHost        postgresFormField = "Host"
	postgresFormFieldPort        postgresFormField = "Port"
	postgresFormFieldUsername    postgresFormField = "Username"
	postgresFormFieldPassword    postgresFormField = "Password"
	postgresFormFieldDatabase    postgresFormField = "Database"
	postgresFormFieldSchema      postgresFormField = "Schema"
	postgresFormFieldSSLMode     postgresFormField = "SSLMode"
)

// SSLModesOrder lists the selectable values for the SSL mode field, in the
// order they are cycled through on the form.
var SSLModesOrder = []config.PostgresSSLMode{
	config.PostgresSSLModeDisable,
	config.PostgresSSLModeAllow,
	config.PostgresSSLModePrefer,
	config.PostgresSSLModeRequire,
	config.PostgresSSLModeVerifyCA,
	config.PostgresSSLModeVerifyFull,
}

// PostgresFormFields is the order in which the fields are displayed.
var PostgresFormFields = []FormField{
	// Connection information
	{
		Key:          strings.ToLower(postgresFormFieldHost),
		Label:        i18n.Sprintf("Host"),
		Example:      postgresExampleHost,
		ValidateFunc: config.ValidatePostgresHost,
	},
	{
		Key:          strings.ToLower(postgresFormFieldPort),
		Label:        i18n.Sprintf("Port"),
		Example:      postgresExamplePort,
		Kind:         IntFieldKind,
		Property:     OptionalFieldProperty,
		DefaultValue: "5432",
		ValidateFunc: func(value string) error {
			if value == "" {
				return nil
			}
			port, err := strconv.Atoi(value)
			if err != nil {
				return errors.New(i18n.Sprintf("port must be a number"))
			}
			return config.ValidatePostgresPort(port)
		},
	},
	{
		Key:          strings.ToLower(postgresFormFieldUsername),
		Label:        i18n.Sprintf("Username"),
		Example:      postgresExampleUsername,
		ValidateFunc: config.ValidatePostgresUsername,
	},
	secretProviderField,
	passwordField,
	{
		Key:          strings.ToLower(postgresFormFieldDatabase),
		Label:        i18n.Sprintf("Database"),
		Example:      postgresExampleDatabase,
		ValidateFunc: config.ValidatePostgresDatabase,
	},
	{
		Key:          strings.ToLower(postgresFormFieldSchema),
		Label:        i18n.Sprintf("Schema"),
		Example:      postgresExampleSchema,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidatePostgresSchema,
	},
	{
		Key:          strings.ToLower(postgresFormFieldSSLMode),
		Label:        i18n.Sprintf("SSLMode"),
		Kind:         SelectFieldKind,
		DefaultValue: config.PostgresSSLModePrefer,
		Options:      SSLModesOrder,
		ValidateFunc: config.ValidatePostgresSSLMode,
	},
	// Meta information (optinal)
	{
		Key:          strings.ToLower(postgresFormFieldName),
		Label:        i18n.Sprintf("Name"),
		Example:      postgresExampleName,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidatePostgresName,
	},
	{
		Key:          strings.ToLower(postgresFormFieldDescription),
		Label:        i18n.Sprintf("Description"),
		Example:      i18n.Sprintf("Production primary database"),
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidatePostgresDescription,
	},
	{
		Key:      strings.ToLower(postgresFormFieldTags),
		Label:    i18n.Sprintf("Tags"),
		Example:  postgresExampleTags,
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

var PostgresFormSpec = FormSpec[config.Connection]{
	AddTitle:  i18n.Sprintf("Add a new Postgres connection"),
	EditTitle: i18n.Sprintf("Edit a Postgres connection"),
	Fields:    PostgresFormFields,
	Sections: []FormSection{
		{
			Title: connectionSectionTitle,
			Fields: []FormFieldKey{
				strings.ToLower(postgresFormFieldHost),
				strings.ToLower(postgresFormFieldPort),
				strings.ToLower(postgresFormFieldUsername),
				SecretProviderKey,
				SecretValueKey,
				strings.ToLower(postgresFormFieldDatabase),
				strings.ToLower(postgresFormFieldSchema),
				strings.ToLower(postgresFormFieldSSLMode),
			},
		},
		transportSection,
		metadataSection(
			strings.ToLower(postgresFormFieldName),
			strings.ToLower(postgresFormFieldDescription),
			strings.ToLower(postgresFormFieldTags),
		),
	},
	SeedFunc: func(c config.Connection) map[FormFieldKey]FormFieldValue {
		pg, ok := c.(config.Postgres)
		if !ok {
			return nil
		}
		mode, value := SplitSecret(pg.Password)
		return map[FormFieldKey]FormFieldValue{
			strings.ToLower(postgresFormFieldHost):        pg.Hostname,
			strings.ToLower(postgresFormFieldPort):        strconv.Itoa(pg.PortNumber),
			strings.ToLower(postgresFormFieldUsername):    pg.User,
			SecretProviderKey:                             mode,
			SecretValueKey:                                value,
			TransportTypeKey:                              transportTypeOf(pg.Transport),
			TransportKey:                                  transportNameOf(pg.Transport),
			strings.ToLower(postgresFormFieldDatabase):    pg.DBName,
			strings.ToLower(postgresFormFieldSchema):      pg.SchemaName,
			strings.ToLower(postgresFormFieldSSLMode):     pg.SSLMode,
			strings.ToLower(postgresFormFieldName):        pg.Metadata.Name,
			strings.ToLower(postgresFormFieldDescription): pg.Metadata.Description,
			strings.ToLower(postgresFormFieldTags):        strings.Join(pg.Metadata.Tags, ", "),
		}
	},
	BuildFunc: func(values map[FormFieldKey]FormFieldValue) (config.Connection, error) {
		port, err := intValue(values, strings.ToLower(postgresFormFieldPort), 5432)
		if err != nil {
			return nil, err
		}

		return config.Postgres{
			Transport: JoinTransport(values[TransportTypeKey], values[TransportKey]),
			Metadata: buildMeta(
				values,
				strings.ToLower(postgresFormFieldName),
				strings.ToLower(postgresFormFieldDescription),
				strings.ToLower(postgresFormFieldTags),
			),
			Hostname:   values[strings.ToLower(postgresFormFieldHost)],
			PortNumber: port,
			User:       values[strings.ToLower(postgresFormFieldUsername)],
			Password: JoinSecret(
				values[SecretProviderKey],
				values[SecretValueKey],
			),
			DBName:     values[strings.ToLower(postgresFormFieldDatabase)],
			SchemaName: values[strings.ToLower(postgresFormFieldSchema)],
			SSLMode:    values[strings.ToLower(postgresFormFieldSSLMode)],
		}, nil
	},
}
