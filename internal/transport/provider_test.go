package transport

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

func TestCache_Get(t *testing.T) {
	t.Parallel()

	type args struct {
		rec    config.Transport
		via    config.Connection
		target config.Connection
	}
	rec := config.Transport{Name: "prod", Kind: config.SSHTransportKind, Via: "bastion"}
	bastion := config.SSH{Hostname: "bastion.local", User: "me", Auth: config.SSHAuthAgent}
	orders := config.Postgres{Hostname: "orders.internal", PortNumber: 5432}
	first := args{rec: rec, via: bastion, target: orders}

	tests := []struct {
		name     string
		args     args
		wantSame bool
		wantErr  string
	}{
		{
			name:     "the same transport to the same database is shared",
			args:     first,
			wantSame: true,
		},
		{
			name: "an ssh transport to another database is its own tunnel",
			args: args{rec: rec, via: bastion, target: config.Postgres{Hostname: "billing.internal", PortNumber: 5432}},
		},
		{
			name: "a changed bastion is a new transport",
			args: args{rec: rec, via: config.SSH{Hostname: "bastion2.local", User: "me", Auth: config.SSHAuthAgent}, target: orders},
		},
		{
			name: "changed transport settings are a new transport",
			args: args{
				rec:    config.Transport{Name: "prod", Kind: config.SSHTransportKind, Via: "bastion", KeepAlive: 30},
				via:    bastion,
				target: orders,
			},
		},
		{
			name:    "an unknown kind",
			args:    args{rec: config.Transport{Name: "prod", Kind: "vpn"}, via: bastion, target: orders},
			wantErr: `unknown kind "vpn"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := NewPool(&fakeTransportLauncher{}, secret.Default())
			before, err := p.Get(first.rec, first.via, first.target)
			require.NoError(t, err)

			got, err := p.Get(tt.args.rec, tt.args.via, tt.args.target)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)

			if tt.wantSame {
				assert.Same(t, before, got)
				return
			}
			assert.NotSame(t, before, got)
		})
	}
}
