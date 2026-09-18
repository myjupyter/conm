package spec

import (
	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
)

const (
	DatabaseEnabled  FormFieldValue = "enabled"
	DatabaseDisabled FormFieldValue = "disabled"
)

type databaseFormField = string

const (
	databaseFormFieldState  databaseFormField = "state"
	databaseFormFieldClient databaseFormField = "client"
)

var DatabaseFormSpecs = map[config.ConnType]FormSpec[config.ConnectionSettings]{
	config.PostgresConnType:   databaseFormSpec(config.PostgresConnType),
	config.MySQLConnType:      databaseFormSpec(config.MySQLConnType),
	config.MSSQLConnType:      databaseFormSpec(config.MSSQLConnType),
	config.ClickHouseConnType: databaseFormSpec(config.ClickHouseConnType),
	config.RedisConnType:      databaseFormSpec(config.RedisConnType),
	config.MongoDBConnType:    databaseFormSpec(config.MongoDBConnType),
}

func DatabaseState(enabled bool) FormFieldValue {
	if enabled {
		return DatabaseEnabled
	}
	return DatabaseDisabled
}

func databaseFormSpec(t config.ConnType) FormSpec[config.ConnectionSettings] {
	return FormSpec[config.ConnectionSettings]{
		EditTitle: "availability",
		Fields: []FormField{
			{
				Key:          databaseFormFieldState,
				Label:        databaseFormFieldState,
				Kind:         SelectFieldKind,
				Options:      []string{DatabaseEnabled, DatabaseDisabled},
				DefaultValue: DatabaseDisabled,
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
				databaseFormFieldState:  DatabaseState(d.Enabled),
				databaseFormFieldClient: d.CLI,
			}
		},
		BuildFunc: func(values map[FormFieldKey]FormFieldValue) (config.ConnectionSettings, error) {
			return config.ConnectionSettings{
				Kind:    config.DatabaseConnKind,
				Type:    t,
				CLI:     values[databaseFormFieldClient],
				Enabled: values[databaseFormFieldState] == DatabaseEnabled,
			}, nil
		},
	}
}
