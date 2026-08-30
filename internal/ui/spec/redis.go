package spec

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/myjupyter/conm/internal/config"
)

const (
	redisExampleHost        = "localhost"
	redisExamplePort        = "6379"
	redisExampleUsername    = "default"
	redisExampleDatabase    = "0"
	redisExampleName        = "cache-primary"
	redisExampleDescription = "Session cache"
	redisExampleTags        = "prod,eu-west,cache"
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
		Label:        redisFormFieldHost,
		Example:      redisExampleHost,
		ValidateFunc: config.ValidateRedisHost,
	},
	{
		Key:          strings.ToLower(redisFormFieldPort),
		Label:        redisFormFieldPort,
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
				return errors.New("port must be a number")
			}
			return config.ValidateRedisPort(port)
		},
	},
	{
		Key:          strings.ToLower(redisFormFieldUsername),
		Label:        redisFormFieldUsername,
		Example:      redisExampleUsername,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateRedisUsername,
	},
	secretProviderField,
	passwordField,
	{
		Key:          strings.ToLower(redisFormFieldDatabase),
		Label:        redisFormFieldDatabase,
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
				return errors.New("database must be a number")
			}
			return config.ValidateRedisDatabase(db)
		},
	},
	{
		Key:          strings.ToLower(redisFormFieldTLSMode),
		Label:        redisFormFieldTLSMode,
		Kind:         SelectFieldKind,
		DefaultValue: config.RedisTLSModeDisable,
		Options:      TLSModesOrder,
		ValidateFunc: config.ValidateRedisTLSMode,
	},
	{
		Key:          strings.ToLower(redisFormFieldName),
		Label:        redisFormFieldName,
		Example:      redisExampleName,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateRedisName,
	},
	{
		Key:          strings.ToLower(redisFormFieldDescription),
		Label:        redisFormFieldDescription,
		Example:      redisExampleDescription,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateRedisDescription,
	},
	{
		Key:      strings.ToLower(redisFormFieldTags),
		Label:    redisFormFieldTags,
		Example:  redisExampleTags,
		Property: OptionalFieldProperty,
		ValidateFunc: func(value string) error {
			if len(value) > tagsMaxLength {
				return fmt.Errorf("tags must be at most %d characters long", tagsMaxLength)
			}
			return nil
		},
	},
}

var RedisFormSpec = FormSpec[config.Connection]{
	AddTitle:  "Add a new Redis connection",
	EditTitle: "Edit a Redis connection",
	Fields:    RedisFormFields,
	Sections: []FormSection{
		{
			Title: "connection",
			Note:  "how conm reaches the server",
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
		{
			Title: "metadata",
			Note:  "yours — never sent to the server",
			Fields: []FormFieldKey{
				strings.ToLower(redisFormFieldName),
				strings.ToLower(redisFormFieldDescription),
				strings.ToLower(redisFormFieldTags),
			},
		},
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
			strings.ToLower(redisFormFieldDatabase):    strconv.Itoa(rd.DBIndex),
			strings.ToLower(redisFormFieldTLSMode):     rd.TLSMode,
			strings.ToLower(redisFormFieldName):        rd.Meta.Name,
			strings.ToLower(redisFormFieldDescription): rd.Meta.Description,
			strings.ToLower(redisFormFieldTags):        strings.Join(rd.Meta.Tags, ", "),
		}
	},
	BuildFunc: func(values map[FormFieldKey]FormFieldValue) (config.Connection, error) {
		port := redisDefaultPort
		if rawPort, ok := values[strings.ToLower(redisFormFieldPort)]; ok && rawPort != "" {
			p, err := strconv.Atoi(rawPort)
			if err != nil {
				return nil, fmt.Errorf("invalid port value: %w", err)
			}
			port = p
		}

		db := redisDefaultDatabase
		if rawDB, ok := values[strings.ToLower(redisFormFieldDatabase)]; ok && rawDB != "" {
			d, err := strconv.Atoi(rawDB)
			if err != nil {
				return nil, fmt.Errorf("invalid database value: %w", err)
			}
			db = d
		}

		return config.Redis{
			Meta: config.ConnMeta{
				Name:        values[strings.ToLower(redisFormFieldName)],
				Description: values[strings.ToLower(redisFormFieldDescription)],
				Tags:        parseTags(values[strings.ToLower(redisFormFieldTags)]),
			},
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
