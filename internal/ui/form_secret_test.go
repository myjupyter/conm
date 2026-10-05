package ui

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
	"github.com/myjupyter/conm/internal/ui/spec"
)

func cycleSSHAuthTo(t *testing.T, m *formModel, auth config.SSHAuth) {
	t.Helper()

	i := slices.IndexFunc(m.fields, func(f spec.FormField) bool { return f.Key == "auth" })
	require.GreaterOrEqual(t, i, 0)
	from, to := slices.Index(spec.SSHAuthOrder, m.vals["auth"]), slices.Index(spec.SSHAuthOrder, auth)
	require.GreaterOrEqual(t, to, 0)
	dir := 1
	if to < from {
		dir = -1
	}
	for range max(to-from, from-to) {
		m.navCycle(i, dir)
	}
	require.Equal(t, auth, m.vals["auth"])
}

func TestFormSSHAuthChangeResetsSecret(t *testing.T) {
	tests := map[string]struct {
		from         config.SSH
		via          []config.SSHAuth
		wantProvider string
	}{
		"password literal to key": {
			from:         config.SSH{Auth: config.SSHAuthPassword, Password: "hunter2"},
			via:          []config.SSHAuth{config.SSHAuthKey},
			wantProvider: secret.Filepath,
		},
		"key filepath to password": {
			from:         config.SSH{Auth: config.SSHAuthKey, Password: "filepath:~/.ssh/id"},
			via:          []config.SSHAuth{config.SSHAuthPassword},
			wantProvider: secret.Literal,
		},
		"password literal to key and back": {
			from:         config.SSH{Auth: config.SSHAuthPassword, Password: "hunter2"},
			via:          []config.SSHAuth{config.SSHAuthKey, config.SSHAuthPassword},
			wantProvider: secret.Literal,
		},
		"key filepath to agent": {
			from:         config.SSH{Auth: config.SSHAuthKey, Password: "filepath:~/.ssh/id"},
			via:          []config.SSHAuth{config.SSHAuthAgent},
			wantProvider: secret.None,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.from.Hostname, tt.from.PortNumber, tt.from.User = "10.0.0.5", 22, "deploy"
			initial := spec.SSHFormSpec.SeedFunc(tt.from)
			require.NotNil(t, initial)
			m := newFormModel(config.SSHConnType, spec.SSHFormSpec, "edit", initial, true, nil)
			require.NotEmpty(t, m.vals[spec.SecretValueKey])

			for _, auth := range tt.via {
				cycleSSHAuthTo(t, &m, auth)
			}

			assert.Equal(t, tt.wantProvider, m.vals[spec.SecretProviderKey])
			assert.Empty(t, m.vals[spec.SecretValueKey])
		})
	}
}

func TestFormSSHSecretLabel(t *testing.T) {
	tests := map[string]struct {
		conn config.SSH
		want string
	}{
		"key from file":         {conn: config.SSH{Auth: config.SSHAuthKey, Password: "filepath:~/.ssh/id"}, want: "private key path"},
		"literal key":           {conn: config.SSH{Auth: config.SSHAuthKey, Password: "-----BEGIN OPENSSH PRIVATE KEY-----"}, want: "private key"},
		"key from keyring":      {conn: config.SSH{Auth: config.SSHAuthKey, Password: "keyring:deploy-key"}, want: "private key"},
		"password from keyring": {conn: config.SSH{Auth: config.SSHAuthPassword, Password: "keyring:deploy-pw"}, want: "password"},
		"literal password":      {conn: config.SSH{Auth: config.SSHAuthPassword, Password: "hunter2"}, want: "password"},
		"agent with no secret":  {conn: config.SSH{Auth: config.SSHAuthAgent}, want: "password"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			tt.conn.Hostname, tt.conn.PortNumber, tt.conn.User = "10.0.0.5", 22, "deploy"
			m := newFormModel(config.SSHConnType, spec.SSHFormSpec, "edit", spec.SSHFormSpec.SeedFunc(tt.conn), true, nil)

			i := slices.IndexFunc(m.fields, func(f spec.FormField) bool { return f.Key == spec.SecretValueKey })
			require.GreaterOrEqual(t, i, 0)
			assert.Equal(t, tt.want, m.label(m.fields[i]))
		})
	}
}
