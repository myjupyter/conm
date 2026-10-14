package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
)

func TestTransportFormSpec_roundTrip(t *testing.T) {
	t.Parallel()

	type args struct {
		rec config.Transport
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "every setting",
			args: args{rec: config.Transport{
				Name: "prod", Kind: config.SSHTransportKind, Via: "bastion",
				LocalPort: 15432, ConnectTimeout: 5, KeepAlive: 30, Description: "eu-west",
			}},
		},
		{
			name: "only what is required",
			args: args{rec: config.Transport{Name: "prod", Kind: config.SSHTransportKind, Via: "bastion"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := TransportFormSpec.BuildFunc(TransportFormSpec.SeedFunc(tt.args.rec))

			require.NoError(t, err)
			assert.Equal(t, tt.args.rec, got)
		})
	}
}

func TestFormSpecs_transport(t *testing.T) {
	t.Parallel()

	type args struct {
		cfg config.Connection
	}
	tests := []struct {
		name     string
		args     args
		wantType string
		wantName string
		want     string
	}{
		{name: "postgres", args: args{cfg: config.Postgres{Transport: "ssh:prod"}}, wantType: "ssh", wantName: "prod", want: "ssh:prod"},
		{name: "mysql", args: args{cfg: config.MySQL{Transport: "ssh:prod"}}, wantType: "ssh", wantName: "prod", want: "ssh:prod"},
		{name: "mssql", args: args{cfg: config.MSSQL{Transport: "ssh:prod"}}, wantType: "ssh", wantName: "prod", want: "ssh:prod"},
		{name: "clickhouse", args: args{cfg: config.ClickHouse{Transport: "ssh:prod"}}, wantType: "ssh", wantName: "prod", want: "ssh:prod"},
		{name: "redis", args: args{cfg: config.Redis{Transport: "ssh:prod"}}, wantType: "ssh", wantName: "prod", want: "ssh:prod"},
		{name: "mongodb", args: args{cfg: config.MongoDB{Transport: "ssh:prod"}}, wantType: "ssh", wantName: "prod", want: "ssh:prod"},
		{name: "a direct connection", args: args{cfg: config.Postgres{}}, wantType: TransportDirect},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			formSpec := FormSpecs[tt.args.cfg.ConnType()]
			values := formSpec.SeedFunc(tt.args.cfg)

			built, err := formSpec.BuildFunc(values)
			require.NoError(t, err)

			got, ok := built.(interface{ TransportRef() string })
			require.True(t, ok)
			is := assert.New(t)
			is.Equal(tt.wantType, values[TransportTypeKey])
			is.Equal(tt.wantName, values[TransportKey])
			is.Equal(tt.want, got.TransportRef())
			is.Contains(sectionTitles(formSpec.Sections), transportSectionTitle)
		})
	}
}

func sectionTitles(sections []FormSection) []string {
	titles := make([]string, 0, len(sections))
	for _, s := range sections {
		titles = append(titles, s.Title)
	}
	return titles
}
