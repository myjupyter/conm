package ui

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui/spec"
)

func TestNewFormModel_name(t *testing.T) {
	type args struct {
		kind    config.ConnType
		initial map[spec.FormFieldKey]spec.FormFieldValue
		isEdit  bool
	}
	parsed := spec.PostgresFormSpec.SeedFunc(config.Postgres{Hostname: "localhost", User: "me", DBName: "mydb"})
	stored := spec.PostgresFormSpec.SeedFunc(config.Postgres{Metadata: config.ConnMeta{Name: "orders"}, Hostname: "localhost"})
	tests := []struct {
		name          string
		args          args
		wantGenerated bool
		want          string
	}{
		{name: "empty postgres add form", args: args{kind: config.PostgresConnType}, wantGenerated: true},
		{name: "empty mysql add form", args: args{kind: config.MySQLConnType}, wantGenerated: true},
		{name: "empty mssql add form", args: args{kind: config.MSSQLConnType}, wantGenerated: true},
		{name: "empty clickhouse add form", args: args{kind: config.ClickHouseConnType}, wantGenerated: true},
		{name: "empty redis add form", args: args{kind: config.RedisConnType}, wantGenerated: true},
		{name: "empty mongodb add form", args: args{kind: config.MongoDBConnType}, wantGenerated: true},
		{name: "empty ssh add form", args: args{kind: config.SSHConnType}, wantGenerated: true},
		{name: "add form prefilled without a name", args: args{kind: config.PostgresConnType, initial: parsed}, wantGenerated: true},
		{name: "add form reopened with a typed name", args: args{kind: config.PostgresConnType, initial: stored}, want: "orders"},
		{name: "edit form keeps the stored name", args: args{kind: config.PostgresConnType, initial: stored, isEdit: true}, want: "orders"},
		{name: "edit form of an unnamed record stays unnamed", args: args{kind: config.PostgresConnType, initial: parsed, isEdit: true}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formSpec, ok := spec.FormSpecs[tt.args.kind]
			require.True(t, ok)

			m := newFormModel(tt.args.kind, formSpec, "form", tt.args.initial, tt.args.isEdit, formContext{})
			got := m.vals["name"]

			if tt.wantGenerated {
				assert.Regexp(t, `^[a-z]+-[a-z]+$`, got)
				assert.NoError(t, config.ValidateConnectionName(got))
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestFormModel_fieldError_name(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "empty name is refused", args: args{value: ""}, want: "name is required"},
		{name: "blank name is refused", args: args{value: "  "}, want: "name is required"},
		{name: "typed name is accepted", args: args{value: "orders"}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newFormModel(config.PostgresConnType, spec.PostgresFormSpec, "add", nil, false, formContext{})
			i := slices.IndexFunc(m.fields, func(f spec.FormField) bool { return f.Key == "name" })
			require.GreaterOrEqual(t, i, 0)
			m.vals["name"] = tt.args.value

			got := m.fieldError(i)

			assert.Equal(t, tt.want, got)
		})
	}
}
