package spec

import (
	"errors"
	"strconv"
	"strings"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/i18n"
)

const (
	mongodbExampleHost       = "localhost"
	mongodbExamplePort       = "27017"
	mongodbExampleUsername   = "root"
	mongodbExampleDatabase   = "app"
	mongodbExampleAuthSource = "admin"
	mongodbExampleName       = "orders-primary"
	mongodbExampleTags       = "prod,eu-west,documents"
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
		Label:        i18n.Sprintf("Host"),
		Example:      mongodbExampleHost,
		ValidateFunc: config.ValidateMongoDBHost,
	},
	{
		Key:          strings.ToLower(mongodbFormFieldPort),
		Label:        i18n.Sprintf("Port"),
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
				return errors.New(i18n.Sprintf("port must be a number"))
			}
			return config.ValidateMongoDBPort(port)
		},
	},
	{
		Key:          strings.ToLower(mongodbFormFieldUsername),
		Label:        i18n.Sprintf("Username"),
		Example:      mongodbExampleUsername,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateMongoDBUsername,
	},
	secretProviderField,
	passwordField,
	{
		Key:          strings.ToLower(mongodbFormFieldDatabase),
		Label:        i18n.Sprintf("Database"),
		Example:      mongodbExampleDatabase,
		ValidateFunc: config.ValidateMongoDBDatabase,
	},
	{
		Key:          strings.ToLower(mongodbFormFieldAuthSource),
		Label:        i18n.Sprintf("Auth source"),
		Example:      mongodbExampleAuthSource,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateMongoDBAuthSource,
	},
	{
		Key:          strings.ToLower(mongodbFormFieldTLS),
		Label:        i18n.Sprintf("TLS"),
		Kind:         SelectFieldKind,
		DefaultValue: mongodbTLSDisabled,
		Options:      []string{mongodbTLSDisabled, mongodbTLSEnabled},
	},
	{
		Key:          strings.ToLower(mongodbFormFieldName),
		Label:        i18n.Sprintf("Name"),
		Example:      mongodbExampleName,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateMongoDBName,
	},
	{
		Key:          strings.ToLower(mongodbFormFieldDescription),
		Label:        i18n.Sprintf("Description"),
		Example:      i18n.Sprintf("Orders document store"),
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateMongoDBDescription,
	},
	{
		Key:      strings.ToLower(mongodbFormFieldTags),
		Label:    i18n.Sprintf("Tags"),
		Example:  mongodbExampleTags,
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

var MongoDBFormSpec = FormSpec[config.Connection]{
	AddTitle:  i18n.Sprintf("Add a new MongoDB connection"),
	EditTitle: i18n.Sprintf("Edit a MongoDB connection"),
	Fields:    MongoDBFormFields,
	Sections: []FormSection{
		{
			Title: connectionSectionTitle,
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
		transportSection,
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
			TransportTypeKey:                             transportTypeOf(mg.Transport),
			TransportKey:                                 transportNameOf(mg.Transport),
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
			Transport: JoinTransport(values[TransportTypeKey], values[TransportKey]),
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
