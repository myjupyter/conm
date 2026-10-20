package spec

import (
	"errors"
	"strconv"
	"strings"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/i18n"
)

const (
	mssqlExampleHost     = "localhost"
	mssqlExamplePort     = "1433"
	mssqlExampleUsername = "sa"
	mssqlExampleDatabase = "master"
	mssqlExampleName     = "prod-primary"
	mssqlExampleTags     = "prod,eu-west,primary"
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
		Label:        i18n.Sprintf("Host"),
		Example:      mssqlExampleHost,
		ValidateFunc: config.ValidateMSSQLHost,
	},
	{
		Key:          strings.ToLower(mssqlFormFieldPort),
		Label:        i18n.Sprintf("Port"),
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
				return errors.New(i18n.Sprintf("port must be a number"))
			}
			return config.ValidateMSSQLPort(port)
		},
	},
	{
		Key:          strings.ToLower(mssqlFormFieldUsername),
		Label:        i18n.Sprintf("Username"),
		Example:      mssqlExampleUsername,
		ValidateFunc: config.ValidateMSSQLUsername,
	},
	secretProviderField,
	passwordField,
	{
		Key:          strings.ToLower(mssqlFormFieldDatabase),
		Label:        i18n.Sprintf("Database"),
		Example:      mssqlExampleDatabase,
		ValidateFunc: config.ValidateMSSQLDatabase,
	},
	{
		Key:          strings.ToLower(mssqlFormFieldEncrypt),
		Label:        i18n.Sprintf("Encrypt"),
		Kind:         SelectFieldKind,
		DefaultValue: config.MSSQLEncryptLogin,
		Options:      MSSQLEncryptModesOrder,
		ValidateFunc: config.ValidateMSSQLEncryptMode,
	},
	{
		Key:          strings.ToLower(mssqlFormFieldTrustCert),
		Label:        i18n.Sprintf("TrustCert"),
		Kind:         SelectFieldKind,
		DefaultValue: mssqlTrustDisabled,
		Options:      []string{mssqlTrustDisabled, mssqlTrustEnabled},
	},
	{
		Key:          strings.ToLower(mssqlFormFieldName),
		Label:        i18n.Sprintf("Name"),
		Example:      mssqlExampleName,
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateMSSQLName,
	},
	{
		Key:          strings.ToLower(mssqlFormFieldDescription),
		Label:        i18n.Sprintf("Description"),
		Example:      i18n.Sprintf("Production primary database"),
		Property:     OptionalFieldProperty,
		ValidateFunc: config.ValidateMSSQLDescription,
	},
	{
		Key:      strings.ToLower(mssqlFormFieldTags),
		Label:    i18n.Sprintf("Tags"),
		Example:  mssqlExampleTags,
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

var MSSQLFormSpec = FormSpec[config.Connection]{
	AddTitle:  i18n.Sprintf("Add a new SQL Server connection"),
	EditTitle: i18n.Sprintf("Edit a SQL Server connection"),
	Fields:    MSSQLFormFields,
	Sections: []FormSection{
		{
			Title: connectionSectionTitle,
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
		transportSection,
		metadataSection(
			strings.ToLower(mssqlFormFieldName),
			strings.ToLower(mssqlFormFieldDescription),
			strings.ToLower(mssqlFormFieldTags),
		),
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
			TransportTypeKey:                           transportTypeOf(ms.Transport),
			TransportKey:                               transportNameOf(ms.Transport),
			strings.ToLower(mssqlFormFieldDatabase):    ms.DBName,
			strings.ToLower(mssqlFormFieldEncrypt):     ms.EncryptMode,
			strings.ToLower(mssqlFormFieldTrustCert):   strconv.FormatBool(ms.TrustCert),
			strings.ToLower(mssqlFormFieldName):        ms.Metadata.Name,
			strings.ToLower(mssqlFormFieldDescription): ms.Metadata.Description,
			strings.ToLower(mssqlFormFieldTags):        strings.Join(ms.Metadata.Tags, ", "),
		}
	},
	BuildFunc: func(values map[FormFieldKey]FormFieldValue) (config.Connection, error) {
		port, err := intValue(values, strings.ToLower(mssqlFormFieldPort), mssqlDefaultPort)
		if err != nil {
			return nil, err
		}

		return config.MSSQL{
			Transport: JoinTransport(values[TransportTypeKey], values[TransportKey]),
			Metadata: buildMeta(
				values,
				strings.ToLower(mssqlFormFieldName),
				strings.ToLower(mssqlFormFieldDescription),
				strings.ToLower(mssqlFormFieldTags),
			),
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
