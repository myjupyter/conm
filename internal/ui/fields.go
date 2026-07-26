package ui

import "github.com/myjupyter/conm/internal/config"

type FieldProperty int

const (
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

type FormField struct {
	Name     string
	Example  string
	Property FieldProperty
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

// PostgresFormFields is the order in which the fields are displayed.
var PostgresFormFields = []FormField{
	// Connection information
	{
		Name:    config.PostgresFormFieldHost,
		Example: postgresExampleHost,
	},
	{
		Name:    config.PostgresFormFieldPort,
		Example: postgresExamplePort,
	},
	{
		Name:    config.PostgresFormFieldUsername,
		Example: postgresExampleUsername,
	},
	{
		Name:    config.PostgresFormFieldPassword,
		Example: postgresExamplePassword,
	},
	{
		Name:    config.PostgresFormFieldDatabase,
		Example: postgresExampleDatabase,
	},
	{
		Name: config.PostgresFormFieldSSLMode,
	},
	// Meta information (optinal)
	{
		Name:     config.PostgresFormFieldName,
		Example:  postgresExampleName,
		Property: OptionalFieldProperty,
	},
	{
		Name:     config.PostgresFormFieldDescription,
		Example:  postgresExampleDescription,
		Property: OptionalFieldProperty,
	},
	{
		Name:     config.PostgresFormFieldTags,
		Example:  postgresExampleTags,
		Property: OptionalFieldProperty,
	},
}

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
