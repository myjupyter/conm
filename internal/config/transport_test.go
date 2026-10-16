package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransport_Validate(t *testing.T) {
	t.Parallel()

	type fields struct {
		Name           string
		Kind           TransportKind
		Via            string
		LocalPort      int
		ConnectTimeout int
		KeepAlive      int
	}
	valid := fields{Name: "prod", Kind: SSHTransportKind, Via: "bastion", ConnectTimeout: 5, KeepAlive: 30}
	with := func(edit func(*fields)) fields {
		f := valid
		edit(&f)
		return f
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr string
	}{
		{
			name:   "valid",
			fields: valid,
		},
		{
			name:   "fixed local port",
			fields: with(func(f *fields) { f.LocalPort = 15432 }),
		},
		{
			name:    "no name",
			fields:  with(func(f *fields) { f.Name = "" }),
			wantErr: "name can't be empty",
		},
		{
			name:    "unknown kind",
			fields:  with(func(f *fields) { f.Kind = "vpn" }),
			wantErr: `unknown transport kind "vpn"`,
		},
		{
			name:    "nothing to go through",
			fields:  with(func(f *fields) { f.Via = "" }),
			wantErr: "connection to go through",
		},
		{
			name:    "local port out of range",
			fields:  with(func(f *fields) { f.LocalPort = 70000 }),
			wantErr: "local port must be between",
		},
		{
			name:    "negative connect timeout",
			fields:  with(func(f *fields) { f.ConnectTimeout = -1 }),
			wantErr: "timeout can't be negative",
		},
		{
			name:    "negative keepalive",
			fields:  with(func(f *fields) { f.KeepAlive = -1 }),
			wantErr: "keepalive can't be negative",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tr := Transport{
				Name:           tt.fields.Name,
				Kind:           tt.fields.Kind,
				Via:            tt.fields.Via,
				LocalPort:      tt.fields.LocalPort,
				ConnectTimeout: tt.fields.ConnectTimeout,
				KeepAlive:      tt.fields.KeepAlive,
			}

			got := tr.Validate()
			if tt.wantErr == "" {
				assert.Empty(t, got)
				return
			}
			require.Len(t, got, 1)
			assert.ErrorContains(t, got[0], tt.wantErr)
		})
	}
}

func TestTransportConfigWrapper_Marshal(t *testing.T) {
	t.Parallel()

	type args struct {
		rec Transport
	}
	tests := []struct {
		name     string
		args     args
		wantPart string
	}{
		{
			name: "a transport with a fixed local port",
			args: args{rec: Transport{
				Name: "prod", Kind: SSHTransportKind, Via: "bastion", LocalPort: 15432, ConnectTimeout: 5, KeepAlive: 30,
			}},
			wantPart: "local_port = 15432",
		},
		{
			name:     "a transport with an automatic local port",
			args:     args{rec: Transport{Name: "prod", Kind: SSHTransportKind, Via: "bastion"}},
			wantPart: "[[transport]]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "transport.toml")
			file, err := OpenConfig[*TransportConfigWrapper](path)
			require.NoError(t, err)
			file.Add(tt.args.rec)
			require.NoError(t, file.Save())

			raw, err := os.ReadFile(path)
			require.NoError(t, err)
			assert.Contains(t, string(raw), tt.wantPart)

			reopened, err := OpenConfig[*TransportConfigWrapper](path)
			require.NoError(t, err)

			require.Equal(t, 1, reopened.Len())
			got := reopened.Get(0)
			assert.True(t, got.IsValid())
			got.validationErrs = nil
			assert.Equal(t, tt.args.rec, got)
		})
	}
}

func TestDBConnection_SetHost_SetPort(t *testing.T) {
	t.Parallel()

	type args struct {
		host string
		port int
	}
	tests := []struct {
		name string
		conn DBConnection
		args args
	}{
		{
			name: "postgres",
			conn: Postgres{Hostname: "db.internal", PortNumber: 5432, User: "svc", DBName: "orders", Transport: "ssh:prod"},
			args: args{host: "127.0.0.1", port: 40123},
		},
		{
			name: "mysql",
			conn: MySQL{Hostname: "db.internal", PortNumber: 3306, User: "svc", DBName: "orders", Transport: "ssh:prod"},
			args: args{host: "127.0.0.1", port: 40123},
		},
		{
			name: "mssql",
			conn: MSSQL{Hostname: "db.internal", PortNumber: 1433, User: "sa", DBName: "master", Transport: "ssh:prod"},
			args: args{host: "127.0.0.1", port: 40123},
		},
		{
			name: "clickhouse",
			conn: ClickHouse{Hostname: "ch.internal", PortNumber: 9000, User: "default", DBName: "events", Transport: "ssh:prod"},
			args: args{host: "127.0.0.1", port: 40123},
		},
		{
			name: "redis",
			conn: Redis{Hostname: "cache.internal", PortNumber: 6379, DBIndex: 2, Transport: "ssh:prod"},
			args: args{host: "127.0.0.1", port: 40123},
		},
		{
			name: "mongodb",
			conn: MongoDB{Hostname: "mongo.internal", PortNumber: 27017, User: "svc", DBName: "orders", Transport: "ssh:prod"},
			args: args{host: "127.0.0.1", port: 40123},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.conn.SetHost(tt.args.host).SetPort(tt.args.port)

			is := assert.New(t)
			is.Equal(tt.args.host, got.Host())
			is.Equal(tt.args.port, got.Port())
			is.Equal(tt.conn, got.SetHost(tt.conn.Host()).SetPort(tt.conn.Port()), "only the address moves")
			is.NotEqual(tt.args.host, tt.conn.Host(), "the original is a value, not touched")
		})
	}
}

func TestTransport_Identity(t *testing.T) {
	t.Parallel()

	type fields struct {
		Name      string
		Kind      TransportKind
		Via       string
		LocalPort int
		KeepAlive int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "kind and name make a transport",
			fields: fields{Name: "prod", Kind: SSHTransportKind, Via: "bastion"},
			want:   `["ssh" "prod"]`,
		},
		{
			name:   "settings do not",
			fields: fields{Name: "prod", Kind: SSHTransportKind, Via: "jump", LocalPort: 15432, KeepAlive: 30},
			want:   `["ssh" "prod"]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tr := Transport{
				Name:      tt.fields.Name,
				Kind:      tt.fields.Kind,
				Via:       tt.fields.Via,
				LocalPort: tt.fields.LocalPort,
				KeepAlive: tt.fields.KeepAlive,
			}

			assert.Equal(t, tt.want, tr.Identity())
		})
	}
}
