package ui

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/transport"
)

var errResolved = errors.New("resolved through the provider")

type recordingProvider struct {
	resolved *config.Connection
}

func (p recordingProvider) Resolve(target config.Connection) (transport.Transport, error) {
	*p.resolved = target
	return nil, errResolved
}

func TestPingFormCmd(t *testing.T) {
	t.Parallel()

	type args struct {
		cfg config.Connection
	}
	tests := []struct {
		name         string
		args         args
		wantResolved bool
		wantErr      error
	}{
		{
			name:         "a connection behind a transport goes through the provider",
			args:         args{cfg: config.Postgres{Hostname: "db.internal", PortNumber: 5432, Transport: "ssh:prod"}},
			wantResolved: true,
			wantErr:      errResolved,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var resolved config.Connection
			msg, ok := pingFormCmd(tt.args.cfg, recordingProvider{resolved: &resolved})().(formPingMsg)

			require.True(t, ok)
			assert.Equal(t, tt.wantResolved, resolved != nil)
			assert.ErrorIs(t, msg.err, tt.wantErr)
		})
	}
}
