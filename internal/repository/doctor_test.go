package repository

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
)

func TestClientCheck_Ready(t *testing.T) {
	type fields struct {
		Client  cli.Info
		Invalid error
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "client installed",
			fields: fields{Client: cli.Info{Name: cli.Psql, Installed: true}},
			want:   true,
		},
		{
			name:   "client not installed",
			fields: fields{Client: cli.Info{Name: cli.Psql}},
			want:   false,
		},
		{
			name:   "no client resolved",
			fields: fields{},
			want:   false,
		},
		{
			name:   "configured client not valid for the type",
			fields: fields{Client: cli.Info{Name: cli.Psql, Installed: true}, Invalid: errors.New("invalid")},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := ClientCheck{Client: tt.fields.Client, Invalid: tt.fields.Invalid}
			assert.Equal(t, tt.want, c.Ready())
		})
	}
}

func Test_checkClient(t *testing.T) {
	type args struct {
		configured string
		installed  []string
	}
	tests := []struct {
		name        string
		args        args
		want        cli.Info
		wantInvalid bool
	}{
		{
			name: "nothing installed",
			args: args{},
			want: cli.Info{},
		},
		{
			name: "first installed client with its version",
			args: args{installed: []string{cli.Pgcli, cli.USQL}},
			want: cli.Info{Name: cli.Pgcli, Installed: true, Version: "17.2"},
		},
		{
			name: "configured client over the first installed",
			args: args{configured: cli.USQL, installed: []string{cli.Pgcli, cli.USQL}},
			want: cli.Info{Name: cli.USQL, Installed: true, Version: "17.2"},
		},
		{
			name: "configured client not installed",
			args: args{configured: cli.Pgcli, installed: []string{cli.Psql}},
			want: cli.Info{Name: cli.Pgcli},
		},
		{
			name:        "configured client not valid for the type",
			args:        args{configured: cli.MySQL, installed: []string{cli.MySQL}},
			want:        cli.Info{Name: cli.MySQL},
			wantInvalid: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, name := range tt.args.installed {
				require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\necho 'client 17.2'\n"), 0o700))
			}
			t.Setenv("PATH", dir)
			if tt.want.Installed {
				tt.want.Path = filepath.Join(dir, tt.want.Name)
			}

			cfg := config.Conm{Connections: []config.ConnectionSettings{{Type: config.PostgresConnType, CLI: tt.args.configured}}}
			got := checkClient(t.Context(), cfg, config.PostgresConnType)

			assert.Equal(t, config.PostgresConnType, got.Type)
			assert.Equal(t, tt.want, got.Client)
			if tt.wantInvalid {
				assert.Error(t, got.Invalid)
			} else {
				assert.NoError(t, got.Invalid)
			}
		})
	}
}

func TestCheckFiles(t *testing.T) {
	validPg := config.Postgres{Metadata: config.ConnMeta{Name: "ok"}, Hostname: "localhost", PortNumber: 5432, User: "u", DBName: "d"}
	badPort := validPg
	badPort.Metadata.Name = "bad-port"
	badPort.PortNumber = 99999
	badTransport := config.Transport{Name: "jump"}

	type args struct {
		transports []config.Transport
		postgres   []config.Connection
		brokenFile func() string
	}
	tests := []struct {
		name       string
		args       args
		wantErr    map[string]bool
		wantRecord map[string][]RecordCheck
	}{
		{
			name: "no files",
			args: args{},
		},
		{
			name: "only invalid records reported",
			args: args{
				transports: []config.Transport{badTransport},
				postgres:   []config.Connection{validPg, badPort},
			},
			wantRecord: map[string][]RecordCheck{
				"transport": {{Name: "jump", Errs: badTransport.Validate()}},
				"postgres":  {{Name: "bad-port", Errs: badPort.Validate()}},
			},
		},
		{
			name:    "unparseable connections file",
			args:    args{brokenFile: config.RedisPath},
			wantErr: map[string]bool{"redis": true},
		},
		{
			name:    "unparseable transport file",
			args:    args{brokenFile: config.TransportPath},
			wantErr: map[string]bool{"transport": true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			t.Setenv("XDG_DATA_HOME", t.TempDir())
			xdg.Reload()
			t.Cleanup(xdg.Reload)
			require.NoError(t, config.UseWorkspace(config.DefaultWorkspace))

			trs, err := config.OpenConfig[*config.TransportConfigWrapper](config.TransportPath())
			require.NoError(t, err)
			trs.Add(tt.args.transports...)
			require.NoError(t, trs.Save())
			pgs, err := config.OpenConfig[*config.PostgresConfigWrapper](config.PostgresPath())
			require.NoError(t, err)
			pgs.Add(tt.args.postgres...)
			require.NoError(t, pgs.Save())
			if tt.args.brokenFile != nil {
				require.NoError(t, os.WriteFile(tt.args.brokenFile(), []byte("broken = ["), 0o600))
			}

			got := CheckFiles(config.Conm{})

			kinds := make([]string, 0, len(got))
			for _, check := range got {
				kinds = append(kinds, check.Kind)
				assert.Equal(t, tt.wantErr[check.Kind], check.Err != nil, check.Kind)
				assert.Equal(t, tt.wantRecord[check.Kind], check.Records, check.Kind)
			}
			assert.Equal(t, append([]string{"transport"}, config.ConnTypeNames()...), kinds)
		})
	}
}
