package importer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
)

func TestRead_ssh(t *testing.T) {
	type args struct {
		content string
	}
	type want struct {
		name     string
		conn     config.Connection
		warnings int
	}
	tests := []struct {
		name string
		args args
		want []want
	}{
		{
			name: "a host block with the options conm keeps",
			args: args{content: "Host bastion\n  HostName bastion.example.com\n  User ops\n  Port 2222\n" +
				"  IdentityFile ~/.ssh/id_ed25519\n  ProxyJump gate\n  ForwardAgent yes\n" +
				"  ServerAliveInterval 30\n  LocalForward 15432 localhost:5432\n"},
			want: []want{{
				name: "bastion",
				conn: config.SSH{
					Metadata:     config.ConnMeta{Name: "bastion"},
					Hostname:     "bastion.example.com",
					PortNumber:   2222,
					User:         "ops",
					Auth:         config.SSHAuthKey,
					Password:     "filepath:~/.ssh/id_ed25519",
					Jump:         "gate",
					ForwardAgent: true,
					KeepAlive:    30,
					LocalForward: "15432:localhost:5432",
				},
			}},
		},
		{
			name: "the alias is the host when there is no HostName, and every alias of a block is a connection",
			args: args{content: "Host db web\n  User=deploy\n"},
			want: []want{
				{
					name: "db",
					conn: config.SSH{Metadata: config.ConnMeta{Name: "db"}, Hostname: "db", PortNumber: 22, User: "deploy", Auth: config.SSHAuthAgent},
				},
				{
					name: "web",
					conn: config.SSH{Metadata: config.ConnMeta{Name: "web"}, Hostname: "web", PortNumber: 22, User: "deploy", Auth: config.SSHAuthAgent},
				},
			},
		},
		{
			name: "a pattern is no connection but its options reach the hosts it matches, the first value winning",
			args: args{content: "User early\n\nHost build.internal\n  Port 2200\n\nHost *.internal\n  User internal\n  Port 2222\n\n" +
				"Host other\n  User own\n\nHost * !other\n  ForwardAgent yes\n"},
			want: []want{
				{
					name: "build.internal",
					conn: config.SSH{
						Metadata:     config.ConnMeta{Name: "build.internal"},
						Hostname:     "build.internal",
						PortNumber:   2200,
						User:         "early",
						Auth:         config.SSHAuthAgent,
						ForwardAgent: true,
					},
				},
				{
					name: "other",
					conn: config.SSH{Metadata: config.ConnMeta{Name: "other"}, Hostname: "other", PortNumber: 22, User: "early", Auth: config.SSHAuthAgent},
				},
			},
		},
		{
			name: "a Match block is skipped whole, an unknown option is a warning, a quoted value loses its quotes",
			args: args{content: "Host one\n  User \"quoted\"\n  StrictHostKeyChecking no\n\nMatch host one\n  User matched\n  Port 9\n\nHost two\n  User plain\n"},
			want: []want{
				{
					name:     "one",
					conn:     config.SSH{Metadata: config.ConnMeta{Name: "one"}, Hostname: "one", PortNumber: 22, User: "quoted", Auth: config.SSHAuthAgent},
					warnings: 1,
				},
				{
					name: "two",
					conn: config.SSH{Metadata: config.ConnMeta{Name: "two"}, Hostname: "two", PortNumber: 22, User: "plain", Auth: config.SSHAuthAgent},
				},
			},
		},
		{
			name: "a config with patterns only holds no connection",
			args: args{content: "Host *\n  ForwardAgent yes\n"},
			want: []want{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config")
			require.NoError(t, os.WriteFile(path, []byte(tt.args.content), 0o600))

			entries, err := Read(config.SSHConnType, path)
			require.NoError(t, err)

			got := make([]want, 0, len(entries))
			for _, e := range entries {
				require.NoError(t, e.Err)
				assert.Empty(t, e.Password)
				got = append(got, want{name: e.Name, conn: e.Conn, warnings: len(e.Warnings)})
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
