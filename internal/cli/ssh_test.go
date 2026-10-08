package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
)

var sshPasswordFlags = []string{
	"-o", "PreferredAuthentications=password", "-o", "PubkeyAuthentication=no",
	"-o", "StrictHostKeyChecking=accept-new",
}

func Test_sshCommand(t *testing.T) {
	t.Parallel()

	type args struct {
		cfg      config.Connection
		password string
	}
	tests := []struct {
		name        string
		args        args
		wantArgs    []string
		wantAskpass bool
		wantErr     error
	}{
		{
			name: "every field as a flag",
			args: args{
				cfg: config.SSH{
					Hostname: "bastion.local", PortNumber: 2222, User: "me", Auth: config.SSHAuthKey,
					Jump: "me@jump.local", ForwardAgent: true, KeepAlive: 30, LocalForward: "5432:localhost:5432",
				},
				password: "~/.ssh/id_ed25519",
			},
			wantArgs: []string{
				"-p", "2222", "-i", "~/.ssh/id_ed25519", "-J", "me@jump.local", "-A",
				"-o", "ServerAliveInterval=30", "-L", "5432:localhost:5432", "me@bastion.local",
			},
		},
		{
			name: "password through askpass",
			args: args{
				cfg:      config.SSH{Hostname: "bastion.local", PortNumber: 22, User: "me", Auth: config.SSHAuthPassword},
				password: "s3cret",
			},
			wantArgs:    append(append([]string{"-p", "22"}, sshPasswordFlags...), "me@bastion.local"),
			wantAskpass: true,
		},
		{
			name: "askpass kept away from the jump hosts",
			args: args{
				cfg: config.SSH{
					Hostname: "db-box", PortNumber: 22, User: "me", Auth: config.SSHAuthPassword,
					Jump: "a@hop1, me@jump.local:2200",
				},
				password: "s3cret",
			},
			wantArgs: append(append([]string{"-p", "22"}, sshPasswordFlags...),
				"-o", "ProxyCommand=env -u SSH_ASKPASS -u SSH_ASKPASS_REQUIRE ssh -J a@hop1 -W %h:%p ssh://me@jump.local:2200",
				"me@db-box",
			),
			wantAskpass: true,
		},
		{
			name: "no password keeps -J",
			args: args{
				cfg: config.SSH{Hostname: "db-box", PortNumber: 22, User: "me", Auth: config.SSHAuthPassword, Jump: "me@jump.local"},
			},
			wantArgs: append(append([]string{"-p", "22"}, sshPasswordFlags...), "-J", "me@jump.local", "me@db-box"),
		},
		{
			name:    "a database connection is refused",
			args:    args{cfg: config.Postgres{Hostname: "db.local", PortNumber: 5432, User: "me", DBName: "app"}},
			wantErr: ErrConnType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := sshCommand(tt.args.cfg, tt.args.password)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			if got.cleanup != nil {
				t.Cleanup(got.cleanup)
			}

			is := assert.New(t)
			is.Equal(tt.wantArgs, got.args)
			if !tt.wantAskpass {
				is.Nil(got.env)
				return
			}
			is.Contains(got.env, "SSH_ASKPASS_REQUIRE=force")
			for _, v := range got.env {
				is.NotContains(v, tt.args.password, "the password must not travel in the environment")
			}
		})
	}
}

func Test_sshIdentity(t *testing.T) {
	t.Parallel()

	const material = "-----BEGIN OPENSSH PRIVATE KEY-----\nAAAA\n-----END OPENSSH PRIVATE KEY-----\n"

	tests := []struct {
		name       string
		secret     string
		wantPath   string
		wantTemp   bool
		wantInFile string
	}{
		{
			name:     "a path is handed over as is",
			secret:   "~/.ssh/id_ed25519",
			wantPath: "~/.ssh/id_ed25519",
		},
		{
			name:       "key material goes to a temp file",
			secret:     material,
			wantTemp:   true,
			wantInFile: material,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotPath, gotRemove, err := sshIdentity(tt.secret)
			require.NoError(t, err)

			if !tt.wantTemp {
				assert.Equal(t, tt.wantPath, gotPath)
				assert.Nil(t, gotRemove)
				return
			}
			require.NotNil(t, gotRemove)
			content, err := os.ReadFile(gotPath)
			require.NoError(t, err)
			assert.Equal(t, tt.wantInFile, string(content))

			gotRemove()
			assert.NoFileExists(t, gotPath)
		})
	}
}

func Test_sshUsesAskpass(t *testing.T) {
	t.Parallel()

	type args struct {
		auth     config.SSHAuth
		password string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "password auth with a password", args: args{auth: config.SSHAuthPassword, password: "s3cret"}, want: true},
		{name: "password auth without a password", args: args{auth: config.SSHAuthPassword}},
		{name: "key auth", args: args{auth: config.SSHAuthKey, password: "~/.ssh/id"}},
		{name: "agent auth", args: args{auth: config.SSHAuthAgent}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, sshUsesAskpass(tt.args.auth, tt.args.password))
		})
	}
}

func Test_sshJumpWithoutAskpass(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		jump string
		want string
	}{
		{
			name: "one hop",
			jump: "me@jump.local",
			want: "env -u SSH_ASKPASS -u SSH_ASKPASS_REQUIRE ssh -W %h:%p ssh://me@jump.local",
		},
		{
			name: "one hop with a port",
			jump: "me@jump.local:2200",
			want: "env -u SSH_ASKPASS -u SSH_ASKPASS_REQUIRE ssh -W %h:%p ssh://me@jump.local:2200",
		},
		{
			name: "earlier hops stay a -J, spaces trimmed",
			jump: " a@hop1 , hop2:23 , me@jump.local",
			want: "env -u SSH_ASKPASS -u SSH_ASKPASS_REQUIRE ssh -J a@hop1,hop2:23 -W %h:%p ssh://me@jump.local",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, sshJumpWithoutAskpass(tt.jump))
		})
	}
}

func Test_sshAskpass(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		password string
	}{
		{name: "plain", password: "s3cret"},
		{name: "shell metacharacters", password: `s3cret with 'quotes' "dq" $HOME ; rm -rf /`},
		{name: "unicode", password: "пароль-🔑"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotPath, gotRemove, err := sshAskpass(tt.password)
			require.NoError(t, err)
			require.NotNil(t, gotRemove)

			out, err := exec.CommandContext(t.Context(), gotPath, "me@db-box's password: ").Output()
			require.NoError(t, err)
			assert.Equal(t, tt.password, string(out))
			assert.NotContains(t, gotPath, tt.password, "the password must not be in the askpass path")

			gotRemove()
			assert.NoDirExists(t, filepath.Dir(gotPath))
		})
	}
}
