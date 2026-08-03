package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/myjupyter/conm/internal/config"
)

const tagsMaxLength = 255

type FieldProperty int

type FieldKind int

const (
	// TextFieldKind is a default kind if isn't set
	TextFieldKind FieldKind = iota
	HiddenFieldKind
	IntFieldKind
	SelectFieldKind
)

const (
	// RequiredFieldProperty is a default property if isn't set
	RequiredFieldProperty FieldProperty = iota
	OptionalFieldProperty
)

func (f FieldProperty) String() string {
	switch f {
	case RequiredFieldProperty:
		return "required"
	case OptionalFieldProperty:
		return "optional"
	default:
		return ""
	}
}

type FormFieldKey = string

type FormFieldValue = string

type FormField struct {
	Key          string
	Label        string
	Kind         FieldKind
	Example      string
	Property     FieldProperty
	DefaultValue string
	Options      []string // only used for SelectFieldKind
	ValidateFunc func(string) error
}

type FormSpec struct {
	Title     string
	Fields    []FormField
	BuildFunc func(map[FormFieldKey]FormFieldValue) (config.ConnectionConfig, error)
}

var formSpecs = map[config.ConnType]FormSpec{
	config.PostgresConnType: postgresFormSpec,
}

const (
	postgresExampleHost        = "localhost"
	postgresExamplePort        = "5432"
	postgresExampleUsername    = "postgres"
	postgresExamplePassword    = "4^@CP^S8\\le9"
	postgresExampleDatabase    = "master"
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
		Key:     strings.ToLower(config.PostgresFormFieldPassword),
		Label:   config.PostgresFormFieldPassword,
		Kind:    HiddenFieldKind,
		Example: postgresExamplePassword,
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

var postgresFormSpec = FormSpec{
	Title:  "Add a new Postgres connection",
	Fields: PostgresFormFields,
	BuildFunc: func(values map[FormFieldKey]FormFieldValue) (config.ConnectionConfig, error) {
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
			SSLMode:    values[strings.ToLower(config.PostgresFormFieldSSLMode)],
		}, nil
	},
}
