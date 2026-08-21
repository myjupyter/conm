package spec

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/myjupyter/conm/internal/config"
)

const (
	postgresExampleHost        = "localhost"
	postgresExamplePort        = "5432"
	postgresExampleUsername    = "postgres"
	postgresExampleDatabase    = "master"
	postgresExampleSchema      = "public"
	postgresExampleName        = "prod-primary"
	postgresExampleDescription = "Production primary database"
	postgresExampleTags        = "prod,eu-west,primary"
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
		Label:        postgresFormFieldHost,
		Example:      postgresExampleHost,
		ValidateFunc: config.ValidatePostgresHost,
	},
	{
		Key:          strings.ToLower(postgresFormFieldPort),
		Label:        postgresFormFieldPort,
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
				return errors.New("port must be a number")
			}
			return config.ValidatePostgresPort(port)
		},
	},
	{
		Key:          strings.ToLower(postgresFormFieldUsername),
		Label:        postgresFormFieldUsername,
		Example:      postgresExampleUsername,
		ValidateFunc: config.ValidatePostgresUsername,
	},
	secretProviderField,
	passwordField,
	{
		Key:          strings.ToLower(postgresFormFieldDatabase),
		Label:        postgresFormFieldDatabase,
		Example:      postgresExampleDatabase,
		ValidateFunc: config.ValidatePostgresDatabase,
	},
	{
		Key:          strings.ToLower(postgresFormFieldSchema),
		Label:        postgresFormFieldSchema,
		Example:      postgresExampleSchema,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidatePostgresSchema,
	},
	{
		Key:          strings.ToLower(postgresFormFieldSSLMode),
		Label:        postgresFormFieldSSLMode,
		Kind:         SelectFieldKind,
		DefaultValue: config.PostgresSSLModePrefer,
		Options:      SSLModesOrder,
		ValidateFunc: config.ValidatePostgresSSLMode,
	},
	// Meta information (optinal)
	{
		Key:          strings.ToLower(postgresFormFieldName),
		Label:        postgresFormFieldName,
		Example:      postgresExampleName,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidatePostgresName,
	},
	{
		Key:          strings.ToLower(postgresFormFieldDescription),
		Label:        postgresFormFieldDescription,
		Example:      postgresExampleDescription,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidatePostgresDescription,
	},
	{
		Key:      strings.ToLower(postgresFormFieldTags),
		Label:    postgresFormFieldTags,
		Example:  postgresExampleTags,
		Property: OptionalFieldProperty,
		ValidateFunc: func(value string) error {
			if len(value) > tagsMaxLength {
				return fmt.Errorf("tags must be at most %d characters long", tagsMaxLength)
			}
			return nil
		},
	},
}

var PostgresFormSpec = FormSpec[config.Connection]{
	AddTitle:  "Add a new Postgres connection",
	EditTitle: "Edit a Postgres connection",
	Fields:    PostgresFormFields,
	Sections: []FormSection{
		{
			Title: "connection",
			Note:  "how conm reaches the server",
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
		{
			Title: "metadata",
			Note:  "yours — never sent to the server",
			Fields: []FormFieldKey{
				strings.ToLower(postgresFormFieldName),
				strings.ToLower(postgresFormFieldDescription),
				strings.ToLower(postgresFormFieldTags),
			},
		},
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
			strings.ToLower(postgresFormFieldDatabase):    pg.DBName,
			strings.ToLower(postgresFormFieldSchema):      pg.SchemaName,
			strings.ToLower(postgresFormFieldSSLMode):     pg.SSLMode,
			strings.ToLower(postgresFormFieldName):        pg.Meta.Name,
			strings.ToLower(postgresFormFieldDescription): pg.Meta.Description,
			strings.ToLower(postgresFormFieldTags):        strings.Join(pg.Meta.Tags, ", "),
		}
	},
	BuildFunc: func(values map[FormFieldKey]FormFieldValue) (config.Connection, error) {
		port := 5432 // default port
		if rawPort, ok := values[strings.ToLower(postgresFormFieldPort)]; ok && rawPort != "" {
			p, err := strconv.Atoi(rawPort)
			if err != nil {
				return nil, fmt.Errorf("invalid port value: %w", err)
			}
			port = p
		}

		return config.Postgres{
			Meta: config.ConnMeta{
				Name:        values[strings.ToLower(postgresFormFieldName)],
				Description: values[strings.ToLower(postgresFormFieldDescription)],
				Tags:        parseTags(values[strings.ToLower(postgresFormFieldTags)]),
			},
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
