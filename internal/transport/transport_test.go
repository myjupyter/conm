package transport

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

func TestThrough(t *testing.T) {
	t.Parallel()

	type args struct {
		kind config.TransportKind
	}
	tests := []struct {
		name  string
		args  args
		want  config.ConnType
		want1 bool
	}{
		{
			name:  "ssh goes through an ssh connection",
			args:  args{kind: config.SSHTransportKind},
			want:  config.SSHConnType,
			want1: true,
		},
		{
			name: "an unknown kind goes through nothing",
			args: args{kind: "vpn"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, got1 := Through(tt.args.kind)

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want1, got1)
		})
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	type args struct {
		rec config.Transport
		via config.Connection
	}
	bastion := config.SSH{Hostname: "bastion.local", User: "me", Auth: config.SSHAuthAgent}
	target := config.Postgres{Hostname: "db.internal", PortNumber: 5432}

	tests := []struct {
		name    string
		args    args
		want    Transport
		wantErr string
	}{
		{
			name: "ssh through an ssh connection",
			args: args{rec: config.Transport{Name: "prod", Kind: config.SSHTransportKind}, via: bastion},
			want: &SSH{},
		},
		{
			name:    "ssh through a database",
			args:    args{rec: config.Transport{Name: "prod", Kind: config.SSHTransportKind}, via: target},
			wantErr: "goes through ssh, not postgres",
		},
		{
			name:    "an unknown kind",
			args:    args{rec: config.Transport{Name: "prod", Kind: "vpn"}, via: bastion},
			wantErr: `unknown kind "vpn"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := New(&fakeTransportLauncher{}, tt.args.rec, tt.args.via, target, secret.Default())
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.IsType(t, tt.want, got)
		})
	}
}
