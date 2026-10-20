package spec

import (
	"errors"
	"strconv"
	"strings"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/i18n"
)

const (
	redisExampleHost     = "localhost"
	redisExamplePort     = "6379"
	redisExampleUsername = "default"
	redisExampleDatabase = "0"
	redisExampleName     = "cache-primary"
	redisExampleTags     = "prod,eu-west,cache"
)

const redisDefaultPort = 6379

const redisDefaultDatabase = 0

type redisFormField = string

const (
	redisFormFieldName        redisFormField = "Name"
	redisFormFieldDescription redisFormField = "Description"
	redisFormFieldTags        redisFormField = "Tags"
	redisFormFieldHost        redisFormField = "Host"
	redisFormFieldPort        redisFormField = "Port"
	redisFormFieldUsername    redisFormField = "Username"
	redisFormFieldDatabase    redisFormField = "Database"
	redisFormFieldTLSMode     redisFormField = "TLS"
)

var TLSModesOrder = []config.RedisTLSMode{
	config.RedisTLSModeDisable,
	config.RedisTLSModeRequire,
}

var RedisFormFields = []FormField{
	{
		Key:          strings.ToLower(redisFormFieldHost),
		Label:        i18n.Sprintf("Host"),
		Example:      redisExampleHost,
		ValidateFunc: config.ValidateRedisHost,
	},
	{
		Key:          strings.ToLower(redisFormFieldPort),
		Label:        i18n.Sprintf("Port"),
		Example:      redisExamplePort,
		Kind:         IntFieldKind,
		Property:     OptionalFieldProperty,
		DefaultValue: strconv.Itoa(redisDefaultPort),
		ValidateFunc: func(value string) error {
			if value == "" {
				return nil
			}
			port, err := strconv.Atoi(value)
			if err != nil {
				return errors.New(i18n.Sprintf("port must be a number"))
			}
			return config.ValidateRedisPort(port)
		},
	},
	{
		Key:          strings.ToLower(redisFormFieldUsername),
		Label:        i18n.Sprintf("Username"),
		Example:      redisExampleUsername,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateRedisUsername,
	},
	secretProviderField,
	passwordField,
	{
		Key:          strings.ToLower(redisFormFieldDatabase),
		Label:        i18n.Sprintf("Database"),
		Example:      redisExampleDatabase,
		Kind:         IntFieldKind,
		Property:     OptionalFieldProperty,
		DefaultValue: strconv.Itoa(redisDefaultDatabase),
		ValidateFunc: func(value string) error {
			if value == "" {
				return nil
			}
			db, err := strconv.Atoi(value)
			if err != nil {
				return errors.New(i18n.Sprintf("database must be a number"))
			}
			return config.ValidateRedisDatabase(db)
		},
	},
	{
		Key:          strings.ToLower(redisFormFieldTLSMode),
		Label:        i18n.Sprintf("TLS"),
		Kind:         SelectFieldKind,
		DefaultValue: config.RedisTLSModeDisable,
		Options:      TLSModesOrder,
		ValidateFunc: config.ValidateRedisTLSMode,
	},
	{
		Key:          strings.ToLower(redisFormFieldName),
		Label:        i18n.Sprintf("Name"),
		Example:      redisExampleName,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateRedisName,
	},
	{
		Key:          strings.ToLower(redisFormFieldDescription),
		Label:        i18n.Sprintf("Description"),
		Example:      i18n.Sprintf("Session cache"),
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateRedisDescription,
	},
	{
		Key:      strings.ToLower(redisFormFieldTags),
		Label:    i18n.Sprintf("Tags"),
		Example:  redisExampleTags,
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

var RedisFormSpec = FormSpec[config.Connection]{
	AddTitle:  i18n.Sprintf("Add a new Redis connection"),
	EditTitle: i18n.Sprintf("Edit a Redis connection"),
	Fields:    RedisFormFields,
	Sections: []FormSection{
		{
			Title: connectionSectionTitle,
			Fields: []FormFieldKey{
				strings.ToLower(redisFormFieldHost),
				strings.ToLower(redisFormFieldPort),
				strings.ToLower(redisFormFieldUsername),
				SecretProviderKey,
				SecretValueKey,
				strings.ToLower(redisFormFieldDatabase),
				strings.ToLower(redisFormFieldTLSMode),
			},
		},
		transportSection,
		metadataSection(
			strings.ToLower(redisFormFieldName),
			strings.ToLower(redisFormFieldDescription),
			strings.ToLower(redisFormFieldTags),
		),
	},
	SeedFunc: func(c config.Connection) map[FormFieldKey]FormFieldValue {
		rd, ok := c.(config.Redis)
		if !ok {
			return nil
		}
		mode, value := SplitSecret(rd.Password)
		return map[FormFieldKey]FormFieldValue{
			strings.ToLower(redisFormFieldHost):        rd.Hostname,
			strings.ToLower(redisFormFieldPort):        strconv.Itoa(rd.PortNumber),
			strings.ToLower(redisFormFieldUsername):    rd.User,
			SecretProviderKey:                          mode,
			SecretValueKey:                             value,
			TransportTypeKey:                           transportTypeOf(rd.Transport),
			TransportKey:                               transportNameOf(rd.Transport),
			strings.ToLower(redisFormFieldDatabase):    strconv.Itoa(rd.DBIndex),
			strings.ToLower(redisFormFieldTLSMode):     rd.TLSMode,
			strings.ToLower(redisFormFieldName):        rd.Metadata.Name,
			strings.ToLower(redisFormFieldDescription): rd.Metadata.Description,
			strings.ToLower(redisFormFieldTags):        strings.Join(rd.Metadata.Tags, ", "),
		}
	},
	BuildFunc: func(values map[FormFieldKey]FormFieldValue) (config.Connection, error) {
		port, err := intValue(values, strings.ToLower(redisFormFieldPort), redisDefaultPort)
		if err != nil {
			return nil, err
		}

		db, err := intValue(values, strings.ToLower(redisFormFieldDatabase), redisDefaultDatabase)
		if err != nil {
			return nil, err
		}

		return config.Redis{
			Transport: JoinTransport(values[TransportTypeKey], values[TransportKey]),
			Metadata: buildMeta(
				values,
				strings.ToLower(redisFormFieldName),
				strings.ToLower(redisFormFieldDescription),
				strings.ToLower(redisFormFieldTags),
			),
			Hostname:   values[strings.ToLower(redisFormFieldHost)],
			PortNumber: port,
			User:       values[strings.ToLower(redisFormFieldUsername)],
			Password: JoinSecret(
				values[SecretProviderKey],
				values[SecretValueKey],
			),
			DBIndex: db,
			TLSMode: values[strings.ToLower(redisFormFieldTLSMode)],
		}, nil
	},
}
