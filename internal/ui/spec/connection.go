package spec

import (
	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
)

const (
	ConnectionEnabled  FormFieldValue = "enabled"
	ConnectionDisabled FormFieldValue = "disabled"
)

type databaseFormField = string

const (
	databaseFormFieldState  databaseFormField = "state"
	databaseFormFieldClient databaseFormField = "client"
)

var ConnectionFormSpecs = map[config.ConnType]FormSpec[config.ConnectionSettings]{
	config.PostgresConnType:   connectionFormSpec(config.PostgresConnType),
	config.MySQLConnType:      connectionFormSpec(config.MySQLConnType),
	config.MSSQLConnType:      connectionFormSpec(config.MSSQLConnType),
	config.ClickHouseConnType: connectionFormSpec(config.ClickHouseConnType),
	config.RedisConnType:      connectionFormSpec(config.RedisConnType),
	config.MongoDBConnType:    connectionFormSpec(config.MongoDBConnType),
	config.SSHConnType:        connectionFormSpec(config.SSHConnType),
}

func ConnectionState(enabled bool) FormFieldValue {
	if enabled {
		return ConnectionEnabled
	}
	return ConnectionDisabled
}

func connectionFormSpec(t config.ConnType) FormSpec[config.ConnectionSettings] {
	return FormSpec[config.ConnectionSettings]{
		EditTitle: "availability",
		Fields: []FormField{
			{
				Key:          databaseFormFieldState,
				Label:        databaseFormFieldState,
				Kind:         SelectFieldKind,
				Options:      []string{ConnectionEnabled, ConnectionDisabled},
				DefaultValue: ConnectionDisabled,
			},
			{
				Key:          databaseFormFieldClient,
				Label:        databaseFormFieldClient,
				Kind:         SelectFieldKind,
				Options:      cli.Clients(t),
				ValidateFunc: func(name string) error { return cli.Validate(t, name) },
			},
		},
		Sections: []FormSection{
			{
				Title:  "availability",
				Note:   "a disabled database keeps its connections but leaves the table",
				Fields: []FormFieldKey{databaseFormFieldState, databaseFormFieldClient},
			},
		},
		SeedFunc: func(d config.ConnectionSettings) map[FormFieldKey]FormFieldValue {
			if d.Type != t {
				return nil
			}
			return map[FormFieldKey]FormFieldValue{
				databaseFormFieldState:  ConnectionState(d.Enabled),
				databaseFormFieldClient: d.CLI,
			}
		},
		BuildFunc: func(values map[FormFieldKey]FormFieldValue) (config.ConnectionSettings, error) {
			return config.ConnectionSettings{
				Type:    t,
				CLI:     values[databaseFormFieldClient],
				Enabled: values[databaseFormFieldState] == ConnectionEnabled,
			}, nil
		},
	}
}
