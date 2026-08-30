package spec

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/myjupyter/conm/internal/config"
)

const (
	mssqlExampleHost        = "localhost"
	mssqlExamplePort        = "1433"
	mssqlExampleUsername    = "sa"
	mssqlExampleDatabase    = "master"
	mssqlExampleName        = "prod-primary"
	mssqlExampleDescription = "Production primary database"
	mssqlExampleTags        = "prod,eu-west,primary"
)

const mssqlDefaultPort = 1433

const (
	mssqlTrustEnabled  FormFieldValue = "true"
	mssqlTrustDisabled FormFieldValue = "false"
)

type mssqlFormField = string

const (
	mssqlFormFieldName        mssqlFormField = "Name"
	mssqlFormFieldDescription mssqlFormField = "Description"
	mssqlFormFieldTags        mssqlFormField = "Tags"
	mssqlFormFieldHost        mssqlFormField = "Host"
	mssqlFormFieldPort        mssqlFormField = "Port"
	mssqlFormFieldUsername    mssqlFormField = "Username"
	mssqlFormFieldDatabase    mssqlFormField = "Database"
	mssqlFormFieldEncrypt     mssqlFormField = "Encrypt"
	mssqlFormFieldTrustCert   mssqlFormField = "TrustCert"
)

var MSSQLEncryptModesOrder = []config.MSSQLEncryptMode{
	config.MSSQLEncryptDisable,
	config.MSSQLEncryptLogin,
	config.MSSQLEncryptRequire,
	config.MSSQLEncryptStrict,
}

var MSSQLFormFields = []FormField{
	{
		Key:          strings.ToLower(mssqlFormFieldHost),
		Label:        mssqlFormFieldHost,
		Example:      mssqlExampleHost,
		ValidateFunc: config.ValidateMSSQLHost,
	},
	{
		Key:          strings.ToLower(mssqlFormFieldPort),
		Label:        mssqlFormFieldPort,
		Example:      mssqlExamplePort,
		Kind:         IntFieldKind,
		Property:     OptionalFieldProperty,
		DefaultValue: strconv.Itoa(mssqlDefaultPort),
		ValidateFunc: func(value string) error {
			if value == "" {
				return nil
			}
			port, err := strconv.Atoi(value)
			if err != nil {
				return errors.New("port must be a number")
			}
			return config.ValidateMSSQLPort(port)
		},
	},
	{
		Key:          strings.ToLower(mssqlFormFieldUsername),
		Label:        mssqlFormFieldUsername,
		Example:      mssqlExampleUsername,
		ValidateFunc: config.ValidateMSSQLUsername,
	},
	secretProviderField,
	passwordField,
	{
		Key:          strings.ToLower(mssqlFormFieldDatabase),
		Label:        mssqlFormFieldDatabase,
		Example:      mssqlExampleDatabase,
		ValidateFunc: config.ValidateMSSQLDatabase,
	},
	{
		Key:          strings.ToLower(mssqlFormFieldEncrypt),
		Label:        mssqlFormFieldEncrypt,
		Kind:         SelectFieldKind,
		DefaultValue: config.MSSQLEncryptLogin,
		Options:      MSSQLEncryptModesOrder,
		ValidateFunc: config.ValidateMSSQLEncryptMode,
	},
	{
		Key:          strings.ToLower(mssqlFormFieldTrustCert),
		Label:        mssqlFormFieldTrustCert,
		Kind:         SelectFieldKind,
		DefaultValue: mssqlTrustDisabled,
		Options:      []string{mssqlTrustDisabled, mssqlTrustEnabled},
	},
	{
		Key:          strings.ToLower(mssqlFormFieldName),
		Label:        mssqlFormFieldName,
		Example:      mssqlExampleName,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateMSSQLName,
	},
	{
		Key:          strings.ToLower(mssqlFormFieldDescription),
		Label:        mssqlFormFieldDescription,
		Example:      mssqlExampleDescription,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateMSSQLDescription,
	},
	{
		Key:      strings.ToLower(mssqlFormFieldTags),
		Label:    mssqlFormFieldTags,
		Example:  mssqlExampleTags,
		Property: OptionalFieldProperty,
		ValidateFunc: func(value string) error {
			if len(value) > tagsMaxLength {
				return fmt.Errorf("tags must be at most %d characters long", tagsMaxLength)
			}
			return nil
		},
	},
}

var MSSQLFormSpec = FormSpec[config.Connection]{
	AddTitle:  "Add a new SQL Server connection",
	EditTitle: "Edit a SQL Server connection",
	Fields:    MSSQLFormFields,
	Sections: []FormSection{
		{
			Title: connectionSectionTitle,
			Note:  connectionSectionNote,
			Fields: []FormFieldKey{
				strings.ToLower(mssqlFormFieldHost),
				strings.ToLower(mssqlFormFieldPort),
				strings.ToLower(mssqlFormFieldUsername),
				SecretProviderKey,
				SecretValueKey,
				strings.ToLower(mssqlFormFieldDatabase),
				strings.ToLower(mssqlFormFieldEncrypt),
				strings.ToLower(mssqlFormFieldTrustCert),
			},
		},
		{
			Title: metadataSectionTitle,
			Note:  metadataSectionNote,
			Fields: []FormFieldKey{
				strings.ToLower(mssqlFormFieldName),
				strings.ToLower(mssqlFormFieldDescription),
				strings.ToLower(mssqlFormFieldTags),
			},
		},
	},
	SeedFunc: func(c config.Connection) map[FormFieldKey]FormFieldValue {
		ms, ok := c.(config.MSSQL)
		if !ok {
			return nil
		}
		mode, value := SplitSecret(ms.Password)
		return map[FormFieldKey]FormFieldValue{
			strings.ToLower(mssqlFormFieldHost):        ms.Hostname,
			strings.ToLower(mssqlFormFieldPort):        strconv.Itoa(ms.PortNumber),
			strings.ToLower(mssqlFormFieldUsername):    ms.User,
			SecretProviderKey:                          mode,
			SecretValueKey:                             value,
			strings.ToLower(mssqlFormFieldDatabase):    ms.DBName,
			strings.ToLower(mssqlFormFieldEncrypt):     ms.EncryptMode,
			strings.ToLower(mssqlFormFieldTrustCert):   strconv.FormatBool(ms.TrustCert),
			strings.ToLower(mssqlFormFieldName):        ms.Meta.Name,
			strings.ToLower(mssqlFormFieldDescription): ms.Meta.Description,
			strings.ToLower(mssqlFormFieldTags):        strings.Join(ms.Meta.Tags, ", "),
		}
	},
	BuildFunc: func(values map[FormFieldKey]FormFieldValue) (config.Connection, error) {
		port := mssqlDefaultPort
		if rawPort, ok := values[strings.ToLower(mssqlFormFieldPort)]; ok && rawPort != "" {
			p, err := strconv.Atoi(rawPort)
			if err != nil {
				return nil, fmt.Errorf("invalid port value: %w", err)
			}
			port = p
		}

		return config.MSSQL{
			Meta: config.ConnMeta{
				Name:        values[strings.ToLower(mssqlFormFieldName)],
				Description: values[strings.ToLower(mssqlFormFieldDescription)],
				Tags:        parseTags(values[strings.ToLower(mssqlFormFieldTags)]),
			},
			Hostname:   values[strings.ToLower(mssqlFormFieldHost)],
			PortNumber: port,
			User:       values[strings.ToLower(mssqlFormFieldUsername)],
			Password: JoinSecret(
				values[SecretProviderKey],
				values[SecretValueKey],
			),
			DBName:      values[strings.ToLower(mssqlFormFieldDatabase)],
			EncryptMode: values[strings.ToLower(mssqlFormFieldEncrypt)],
			TrustCert:   values[strings.ToLower(mssqlFormFieldTrustCert)] == mssqlTrustEnabled,
		}, nil
	},
}
