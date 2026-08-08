package spec

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/myjupyter/conm/internal/config"
)

const (
	postgresExampleHost        = "localhost"
	postgresExamplePort        = "5432"
	postgresExampleUsername    = "postgres"
	postgresExamplePassword    = "4^@CP^S8\\le9"
	postgresExampleDatabase    = "master"
	postgresExampleSchema      = "public"
	postgresExampleName        = "prod-primary"
	postgresExampleDescription = "Production primary database"
	postgresExampleTags        = "prod,eu-west,primary"
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

// PostgresFormFields is the order in which the fields are displayed
var PostgresFormFields = []FormField{
	// Connection information
	{
		Key:          strings.ToLower(config.PostgresFormFieldHost),
		Label:        config.PostgresFormFieldHost,
		Example:      postgresExampleHost,
		ValidateFunc: config.ValidatePostgresHost,
	},
	{
		Key:          strings.ToLower(config.PostgresFormFieldPort),
		Label:        config.PostgresFormFieldPort,
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
				return fmt.Errorf("port must be a number")
			}
			return config.ValidatePostgresPort(port)
		},
	},
	{
		Key:          strings.ToLower(config.PostgresFormFieldUsername),
		Label:        config.PostgresFormFieldUsername,
		Example:      postgresExampleUsername,
		ValidateFunc: config.ValidatePostgresUsername,
	},
	{
		Key:      strings.ToLower(config.PostgresFormFieldPassword),
		Label:    config.PostgresFormFieldPassword,
		Kind:     HiddenFieldKind,
		Example:  postgresExamplePassword,
		Property: OptionalFieldProperty,
		ValidateFunc: func(value string) error {
			return nil // no specific validation for password
		},
	},
	{
		Key:          strings.ToLower(config.PostgresFormFieldDatabase),
		Label:        config.PostgresFormFieldDatabase,
		Example:      postgresExampleDatabase,
		ValidateFunc: config.ValidatePostgresDatabase,
	},
	{
		Key:          strings.ToLower(config.PostgresFormFieldSchema),
		Label:        config.PostgresFormFieldSchema,
		Example:      postgresExampleSchema,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidatePostgresSchema,
	},
	{
		Key:          strings.ToLower(config.PostgresFormFieldSSLMode),
		Label:        config.PostgresFormFieldSSLMode,
		Kind:         SelectFieldKind,
		DefaultValue: config.PostgresSSLModePrefer,
		Options:      SSLModesOrder,
		ValidateFunc: config.ValidatePostgresSSLMode,
	},
	// Meta information (optinal)
	{
		Key:          strings.ToLower(config.PostgresFormFieldName),
		Label:        config.PostgresFormFieldName,
		Example:      postgresExampleName,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidatePostgresName,
	},
	{
		Key:          strings.ToLower(config.PostgresFormFieldDescription),
		Label:        config.PostgresFormFieldDescription,
		Example:      postgresExampleDescription,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidatePostgresDescription,
	},
	{
		Key:      strings.ToLower(config.PostgresFormFieldTags),
		Label:    config.PostgresFormFieldTags,
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

var PostgresFormSpec = FormSpec{
	AddTitle:  "Add a new Postgres connection",
	EditTitle: "Edit a Postgres connection",
	Fields:    PostgresFormFields,
	Sections: []FormSection{
		{
			Title: "connection",
			Note:  "how conm reaches the server",
			Fields: []FormFieldKey{
				strings.ToLower(config.PostgresFormFieldHost),
				strings.ToLower(config.PostgresFormFieldPort),
				strings.ToLower(config.PostgresFormFieldUsername),
				strings.ToLower(config.PostgresFormFieldPassword),
				strings.ToLower(config.PostgresFormFieldDatabase),
				strings.ToLower(config.PostgresFormFieldSchema),
				strings.ToLower(config.PostgresFormFieldSSLMode),
			},
		},
		{
			Title: "metadata",
			Note:  "yours — never sent to the server",
			Fields: []FormFieldKey{
				strings.ToLower(config.PostgresFormFieldName),
				strings.ToLower(config.PostgresFormFieldDescription),
				strings.ToLower(config.PostgresFormFieldTags),
			},
		},
	},
	SeedFunc: func(c config.Connection) map[FormFieldKey]FormFieldValue {
		pg, ok := c.(config.Postgres)
		if !ok {
			return nil
		}
		return map[FormFieldKey]FormFieldValue{
			strings.ToLower(config.PostgresFormFieldHost):        pg.Hostname,
			strings.ToLower(config.PostgresFormFieldPort):        strconv.Itoa(pg.PortNumber),
			strings.ToLower(config.PostgresFormFieldUsername):    pg.User,
			strings.ToLower(config.PostgresFormFieldPassword):    pg.Password,
			strings.ToLower(config.PostgresFormFieldDatabase):    pg.DBName,
			strings.ToLower(config.PostgresFormFieldSchema):      pg.SchemaName,
			strings.ToLower(config.PostgresFormFieldSSLMode):     pg.SSLMode,
			strings.ToLower(config.PostgresFormFieldName):        pg.Meta.Name,
			strings.ToLower(config.PostgresFormFieldDescription): pg.Meta.Description,
			strings.ToLower(config.PostgresFormFieldTags):        strings.Join(pg.Meta.Tags, ", "),
		}
	},
	BuildFunc: func(values map[FormFieldKey]FormFieldValue) (config.Connection, error) {
		port := 5432 // default port
		if rawPort, ok := values[strings.ToLower(config.PostgresFormFieldPort)]; ok && rawPort != "" {
			p, err := strconv.Atoi(rawPort)
			if err != nil {
				return nil, fmt.Errorf("invalid port value: %v", err)
			}
			port = p
		}

		return config.Postgres{
			Meta: config.ConnMeta{
				Name:        values[strings.ToLower(config.PostgresFormFieldName)],
				Description: values[strings.ToLower(config.PostgresFormFieldDescription)],
				Tags:        parseTags(values[strings.ToLower(config.PostgresFormFieldTags)]),
			},
			Hostname:   values[strings.ToLower(config.PostgresFormFieldHost)],
			PortNumber: port,
			User:       values[strings.ToLower(config.PostgresFormFieldUsername)],
			Password:   values[strings.ToLower(config.PostgresFormFieldPassword)],
			DBName:     values[strings.ToLower(config.PostgresFormFieldDatabase)],
			SchemaName: values[strings.ToLower(config.PostgresFormFieldSchema)],
			SSLMode:    values[strings.ToLower(config.PostgresFormFieldSSLMode)],
		}, nil
	},
}
