package ui

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/repository"
)

func Test_conmCheck(t *testing.T) {
	valid := config.Conm{Keybindings: config.DefaultKeybindings()}
	require.NoError(t, valid.Keybindings.Validate())
	badKeys := config.Conm{Keybindings: config.DefaultKeybindings()}
	badKeys.Keybindings.General.Confirm = []string{"ctrl+nope"}
	keysErr := badKeys.Keybindings.Validate()
	require.Error(t, keysErr)

	type args struct {
		conm    config.Conm
		conmErr error
	}
	tests := []struct {
		name string
		args args
		want doctorCheck
	}{
		{
			name: "readable with valid keybindings",
			args: args{conm: valid},
			want: doctorCheck{kind: kindOK, title: "Conm internal configs"},
		},
		{
			name: "unreadable file",
			args: args{conmErr: errors.New("couldn't parse\nline 2")},
			want: doctorCheck{kind: kindErr, title: "Conm internal configs", problems: []string{"couldn't parse", "line 2"}},
		},
		{
			name: "invalid keybindings",
			args: args{conm: badKeys},
			want: doctorCheck{kind: kindErr, title: "Conm internal configs", problems: errorLines(keysErr)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, conmCheck(tt.args.conm, tt.args.conmErr))
		})
	}
}

func Test_clientCheck(t *testing.T) {
	tests := []struct {
		name string
		c    repository.ClientCheck
		want doctorCheck
	}{
		{
			name: "installed with version",
			c:    repository.ClientCheck{Type: config.PostgresConnType, Client: cli.Info{Name: cli.Psql, Installed: true, Version: "17.2"}},
			want: doctorCheck{kindOK, "Postgres (psql 17.2)", nil},
		},
		{
			name: "installed without version",
			c:    repository.ClientCheck{Type: config.MSSQLConnType, Client: cli.Info{Name: cli.SQLCmd, Installed: true}},
			want: doctorCheck{kindOK, "Mssql (sqlcmd)", nil},
		},
		{
			name: "no client at all",
			c:    repository.ClientCheck{Type: config.MongoDBConnType},
			want: doctorCheck{kindWarn, "Mongodb (no client installed)", []string{"no client installed"}},
		},
		{
			name: "configured client not installed",
			c:    repository.ClientCheck{Type: config.PostgresConnType, Client: cli.Info{Name: cli.Pgcli}},
			want: doctorCheck{kindErr, "Postgres (pgcli)", []string{"pgcli is not installed"}},
		},
		{
			name: "configured client not valid for the type",
			c:    repository.ClientCheck{Type: config.PostgresConnType, Client: cli.Info{Name: cli.MySQL}, Invalid: errors.New("invalid CLI")},
			want: doctorCheck{kindErr, "Postgres (mysql)", []string{"invalid CLI"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, clientCheck(tt.c))
		})
	}
}

func Test_fileCheck(t *testing.T) {
	tests := []struct {
		name string
		c    repository.FileCheck
		want doctorCheck
	}{
		{
			name: "no problems",
			c:    repository.FileCheck{Kind: "postgres"},
			want: doctorCheck{kind: kindOK, title: "Postgres"},
		},
		{
			name: "unreadable file",
			c:    repository.FileCheck{Kind: "redis", Err: errors.New("couldn't parse")},
			want: doctorCheck{kind: kindErr, title: "Redis", problems: []string{"couldn't parse"}},
		},
		{
			name: "one line per record error",
			c: repository.FileCheck{Kind: "transport", Records: []repository.RecordCheck{
				{Name: "jump", Errs: []error{errors.New("via is required"), errors.New("unknown kind")}},
				{Name: "edge", Errs: []error{errors.New("name is required")}},
			}},
			want: doctorCheck{kind: kindErr, title: "Transport", problems: []string{
				`"jump": via is required`,
				`"jump": unknown kind`,
				`"edge": name is required`,
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, fileCheck(tt.c))
		})
	}
}
