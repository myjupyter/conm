package params

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

func TestParseSSH(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		syntax   Syntax
		client   string
		conn     config.SSH
		warnings []string
	}{
		{
			name:   "flags and a destination",
			input:  `ssh -p 2222 -i ~/.ssh/id_ci -J deploy@bastion.eu -A -o ServerAliveInterval=30 -L 5432:localhost:5432 ci@runner-03.ci`,
			syntax: FlagSyntax,
			client: "ssh",
			conn: config.SSH{
				Hostname:     "runner-03.ci",
				PortNumber:   2222,
				User:         "ci",
				Auth:         config.SSHAuthKey,
				Password:     secret.Ref(secret.Filepath, "~/.ssh/id_ci"),
				Jump:         "deploy@bastion.eu",
				ForwardAgent: true,
				KeepAlive:    30,
				LocalForward: "5432:localhost:5432",
			},
		},
		{
			name:   "a bare url needs no client",
			input:  `ssh://deploy@bastion.eu:2200`,
			syntax: URISyntax,
			conn: config.SSH{
				Hostname:   "bastion.eu",
				PortNumber: 2200,
				User:       "deploy",
				Auth:       config.SSHAuthAgent,
			},
		},
		{
			name:   "user from a flag, unknown flags and the remote command warned",
			input:  `ssh -l root -vC -F ~/.ssh/config host.internal uptime`,
			syntax: FlagSyntax,
			client: "ssh",
			conn: config.SSH{
				Hostname:   "host.internal",
				PortNumber: 22,
				User:       "root",
				Auth:       config.SSHAuthAgent,
			},
			warnings: []string{"ignored flag -v", "ignored flag -C", "ignored flag -F", `dropped remote command "uptime"`},
		},
		{
			name:     "dashed arguments of the remote command are not flags",
			input:    `ssh host.internal ls -la`,
			syntax:   FlagSyntax,
			client:   "ssh",
			conn:     config.SSH{Hostname: "host.internal", PortNumber: 22, Auth: config.SSHAuthAgent},
			warnings: []string{`dropped remote command "ls -la"`},
		},
		{
			name:     "flags right after the destination still count",
			input:    `ssh host.internal -p 2222 -- ls -la`,
			syntax:   FlagSyntax,
			client:   "ssh",
			conn:     config.SSH{Hostname: "host.internal", PortNumber: 2222, Auth: config.SSHAuthAgent},
			warnings: []string{`dropped remote command "ls -la"`},
		},
		{
			name:   "-l wins over the user in the destination",
			input:  `ssh -l alice bob@host.internal`,
			syntax: FlagSyntax,
			client: "ssh",
			conn:   config.SSH{Hostname: "host.internal", PortNumber: 22, User: "alice", Auth: config.SSHAuthAgent},
		},
		{
			name:   "HostName wins over the destination alias",
			input:  `ssh -o HostName=10.0.0.5 prod`,
			syntax: FlagSyntax,
			client: "ssh",
			conn:   config.SSH{Hostname: "10.0.0.5", PortNumber: 22, Auth: config.SSHAuthAgent},
		},
		{
			name:   "-p wins over the url port and an earlier -o Port",
			input:  `ssh -o Port=1 -p 2222 ssh://host.internal:2200`,
			syntax: URISyntax,
			client: "ssh",
			conn:   config.SSH{Hostname: "host.internal", PortNumber: 2222, Auth: config.SSHAuthAgent},
		},
		{
			name:   "an option takes a space as its separator",
			input:  `ssh -o "User alice" -o "Port = 2222" host.internal`,
			syntax: FlagSyntax,
			client: "ssh",
			conn:   config.SSH{Hostname: "host.internal", PortNumber: 2222, User: "alice", Auth: config.SSHAuthAgent},
		},
		{
			name:   "the first value of an option wins",
			input:  `ssh -o User=alice -o User=bob host.internal`,
			syntax: FlagSyntax,
			client: "ssh",
			conn:   config.SSH{Hostname: "host.internal", PortNumber: 22, User: "alice", Auth: config.SSHAuthAgent},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := Parse(config.SSHConnType, tt.input)
			require.NoError(t, err)

			is := assert.New(t)
			is.Equal(tt.syntax, res.Syntax)
			is.Equal(tt.client, res.Client)
			is.Equal(tt.conn, res.Conn)
			is.Equal(tt.warnings, res.Warnings)
		})
	}
}

func TestParseSSHRefusesForeignScheme(t *testing.T) {
	_, err := Parse(config.SSHConnType, "postgres://me@db.local:5432/app")
	require.ErrorIs(t, err, ErrTypeMismatch)
}
