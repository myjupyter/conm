package spec

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/myjupyter/conm/internal/config"
)

const (
	mongodbExampleHost        = "localhost"
	mongodbExamplePort        = "27017"
	mongodbExampleUsername    = "root"
	mongodbExampleDatabase    = "app"
	mongodbExampleAuthSource  = "admin"
	mongodbExampleName        = "orders-primary"
	mongodbExampleDescription = "Orders document store"
	mongodbExampleTags        = "prod,eu-west,documents"
)

const mongodbDefaultPort = 27017

const (
	mongodbTLSEnabled  FormFieldValue = "true"
	mongodbTLSDisabled FormFieldValue = "false"
)

type mongodbFormField = string

const (
	mongodbFormFieldName        mongodbFormField = "Name"
	mongodbFormFieldDescription mongodbFormField = "Description"
	mongodbFormFieldTags        mongodbFormField = "Tags"
	mongodbFormFieldHost        mongodbFormField = "Host"
	mongodbFormFieldPort        mongodbFormField = "Port"
	mongodbFormFieldUsername    mongodbFormField = "Username"
	mongodbFormFieldDatabase    mongodbFormField = "Database"
	mongodbFormFieldAuthSource  mongodbFormField = "Auth source"
	mongodbFormFieldTLS         mongodbFormField = "TLS"
)

var MongoDBFormFields = []FormField{
	{
		Key:          strings.ToLower(mongodbFormFieldHost),
		Label:        mongodbFormFieldHost,
		Example:      mongodbExampleHost,
		ValidateFunc: config.ValidateMongoDBHost,
	},
	{
		Key:          strings.ToLower(mongodbFormFieldPort),
		Label:        mongodbFormFieldPort,
		Example:      mongodbExamplePort,
		Kind:         IntFieldKind,
		Property:     OptionalFieldProperty,
		DefaultValue: strconv.Itoa(mongodbDefaultPort),
		ValidateFunc: func(value string) error {
			if value == "" {
				return nil
			}
			port, err := strconv.Atoi(value)
			if err != nil {
				return errors.New("port must be a number")
			}
			return config.ValidateMongoDBPort(port)
		},
	},
	{
		Key:          strings.ToLower(mongodbFormFieldUsername),
		Label:        mongodbFormFieldUsername,
		Example:      mongodbExampleUsername,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateMongoDBUsername,
	},
	secretProviderField,
	passwordField,
	{
		Key:          strings.ToLower(mongodbFormFieldDatabase),
		Label:        mongodbFormFieldDatabase,
		Example:      mongodbExampleDatabase,
		ValidateFunc: config.ValidateMongoDBDatabase,
	},
	{
		Key:          strings.ToLower(mongodbFormFieldAuthSource),
		Label:        mongodbFormFieldAuthSource,
		Example:      mongodbExampleAuthSource,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateMongoDBAuthSource,
	},
	{
		Key:          strings.ToLower(mongodbFormFieldTLS),
		Label:        mongodbFormFieldTLS,
		Kind:         SelectFieldKind,
		DefaultValue: mongodbTLSDisabled,
		Options:      []string{mongodbTLSDisabled, mongodbTLSEnabled},
	},
	{
		Key:          strings.ToLower(mongodbFormFieldName),
		Label:        mongodbFormFieldName,
		Example:      mongodbExampleName,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateMongoDBName,
	},
	{
		Key:          strings.ToLower(mongodbFormFieldDescription),
		Label:        mongodbFormFieldDescription,
		Example:      mongodbExampleDescription,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateMongoDBDescription,
	},
	{
		Key:      strings.ToLower(mongodbFormFieldTags),
		Label:    mongodbFormFieldTags,
		Example:  mongodbExampleTags,
		Property: OptionalFieldProperty,
		ValidateFunc: func(value string) error {
			if len(value) > tagsMaxLength {
				return fmt.Errorf("tags must be at most %d characters long", tagsMaxLength)
			}
			return nil
		},
	},
}

var MongoDBFormSpec = FormSpec[config.Connection]{
	AddTitle:  "Add a new MongoDB connection",
	EditTitle: "Edit a MongoDB connection",
	Fields:    MongoDBFormFields,
	Sections: []FormSection{
		{
			Title: connectionSectionTitle,
			Note:  connectionSectionNote,
			Fields: []FormFieldKey{
				strings.ToLower(mongodbFormFieldHost),
				strings.ToLower(mongodbFormFieldPort),
				strings.ToLower(mongodbFormFieldUsername),
				SecretProviderKey,
				SecretValueKey,
				strings.ToLower(mongodbFormFieldDatabase),
				strings.ToLower(mongodbFormFieldAuthSource),
				strings.ToLower(mongodbFormFieldTLS),
			},
		},
		metadataSection(
			strings.ToLower(mongodbFormFieldName),
			strings.ToLower(mongodbFormFieldDescription),
			strings.ToLower(mongodbFormFieldTags),
		),
	},
	SeedFunc: func(c config.Connection) map[FormFieldKey]FormFieldValue {
		mg, ok := c.(config.MongoDB)
		if !ok {
			return nil
		}
		mode, value := SplitSecret(mg.Password)
		return map[FormFieldKey]FormFieldValue{
			strings.ToLower(mongodbFormFieldHost):        mg.Hostname,
			strings.ToLower(mongodbFormFieldPort):        strconv.Itoa(mg.PortNumber),
			strings.ToLower(mongodbFormFieldUsername):    mg.User,
			SecretProviderKey:                            mode,
			SecretValueKey:                               value,
			strings.ToLower(mongodbFormFieldDatabase):    mg.DBName,
			strings.ToLower(mongodbFormFieldAuthSource):  mg.AuthSource,
			strings.ToLower(mongodbFormFieldTLS):         strconv.FormatBool(mg.TLS),
			strings.ToLower(mongodbFormFieldName):        mg.Metadata.Name,
			strings.ToLower(mongodbFormFieldDescription): mg.Metadata.Description,
			strings.ToLower(mongodbFormFieldTags):        strings.Join(mg.Metadata.Tags, ", "),
		}
	},
	BuildFunc: func(values map[FormFieldKey]FormFieldValue) (config.Connection, error) {
		port, err := intValue(values, strings.ToLower(mongodbFormFieldPort), mongodbDefaultPort)
		if err != nil {
			return nil, err
		}

		return config.MongoDB{
			Metadata: buildMeta(
				values,
				strings.ToLower(mongodbFormFieldName),
				strings.ToLower(mongodbFormFieldDescription),
				strings.ToLower(mongodbFormFieldTags),
			),
			Hostname:   values[strings.ToLower(mongodbFormFieldHost)],
			PortNumber: port,
			User:       values[strings.ToLower(mongodbFormFieldUsername)],
			Password: JoinSecret(
				values[SecretProviderKey],
				values[SecretValueKey],
			),
			DBName:     values[strings.ToLower(mongodbFormFieldDatabase)],
			AuthSource: values[strings.ToLower(mongodbFormFieldAuthSource)],
			TLS:        values[strings.ToLower(mongodbFormFieldTLS)] == mongodbTLSEnabled,
		}, nil
	},
}
