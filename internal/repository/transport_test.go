package repository

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/network"
)

type linkedWorkspace struct {
	*Workspace
	postgres *ConnectionRepository
	ssh      *ConnectionRepository
}

func testWorkspace(t *testing.T) linkedWorkspace {
	t.Helper()

	dir := t.TempDir()
	conm := config.Conm{Connections: []config.ConnectionSettings{
		{Type: config.PostgresConnType, CLI: "psql", Enabled: true},
		{Type: config.SSHConnType, CLI: "ssh", Enabled: true},
	}}

	file, err := config.OpenConfig[*config.TransportConfigWrapper](filepath.Join(dir, "transport.toml"))
	require.NoError(t, err)
	t.Cleanup(func() { file.Close() })

	w := &Workspace{byKind: map[config.ConnType]Connections{}}
	w.Transports = newTransportRepository(file, UsageLookupFunc(w.transportUsages), w)
	connectionUsage := NewConnectionIndex(w.Transports)

	open := func(open func() (*ConnectionRepository, error)) *ConnectionRepository {
		repo, err := open()
		require.NoError(t, err)
		t.Cleanup(func() { repo.Close() })
		w.conns = append(w.conns, repo)
		w.byKind[repo.Kind()] = repo
		return repo
	}
	postgres := open(func() (*ConnectionRepository, error) {
		return openConnections[*config.PostgresConfigWrapper](
			conm, config.PostgresConnType, filepath.Join(dir, "postgres.toml"), w.Transports, connectionUsage,
		)
	})
	ssh := open(func() (*ConnectionRepository, error) {
		return openConnections[*config.SSHConfigWrapper](
			conm, config.SSHConnType, filepath.Join(dir, "ssh.toml"), w.Transports, connectionUsage,
		)
	})

	return linkedWorkspace{Workspace: w, postgres: postgres, ssh: ssh}
}

func bastion(name, host string) config.SSH {
	return config.SSH{
		Metadata: config.ConnMeta{Name: name}, Hostname: host, PortNumber: 22, User: "me", Auth: config.SSHAuthAgent,
	}
}

func prodTransport(via string) config.Transport {
	return config.Transport{Name: "prod", Kind: config.SSHTransportKind, Via: via}
}

func behind(transportRef string) config.Postgres {
	db := pg("orders", "db.internal", "orders")
	db.Transport = transportRef
	return db
}

func withBastion(t *testing.T, w linkedWorkspace) {
	t.Helper()

	require.NoError(t, w.ssh.Add(bastion("bastion", "bastion.local")))
}

func withTransport(t *testing.T, w linkedWorkspace) {
	t.Helper()

	withBastion(t, w)
	require.NoError(t, w.Transports.Add(prodTransport("bastion")))
}

func withDatabaseBehind(t *testing.T, w linkedWorkspace) {
	t.Helper()

	withTransport(t, w)
	require.NoError(t, w.postgres.Add(behind("ssh:prod")))
}

func assertErrorOfType(t *testing.T, want, err error) {
	t.Helper()

	for e := err; e != nil; e = errors.Unwrap(e) {
		if reflect.TypeOf(e) == reflect.TypeOf(want) {
			return
		}
	}
	assert.Failf(t, "wrong error", "error = %v, want %T", err, want)
}

func TestTransportRepository_Add(t *testing.T) {
	type args struct {
		t config.Transport
	}
	tests := []struct {
		name    string
		prepare func(*testing.T, linkedWorkspace)
		args    args
		wantErr error
	}{
		{
			name:    "a transport through an existing bastion",
			prepare: withBastion,
			args:    args{t: prodTransport("bastion")},
		},
		{
			name:    "a transport through a missing bastion",
			args:    args{t: prodTransport("bastion")},
			wantErr: &ConnectionNotFoundError{},
		},
		{
			name:    "a second transport with the same name",
			prepare: withTransport,
			args:    args{t: prodTransport("bastion")},
			wantErr: &DuplicateTransportError{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := testWorkspace(t)
			if tt.prepare != nil {
				tt.prepare(t, w)
			}

			err := w.Transports.Add(tt.args.t)

			if tt.wantErr != nil {
				assertErrorOfType(t, tt.wantErr, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestTransportRepository_Edit(t *testing.T) {
	type args struct {
		i int
		t config.Transport
	}
	renamed := prodTransport("bastion")
	renamed.Name = "staging"
	tuned := prodTransport("bastion")
	tuned.KeepAlive = 30

	tests := []struct {
		name    string
		prepare func(*testing.T, linkedWorkspace)
		args    args
		wantErr error
	}{
		{
			name:    "an unused transport can be renamed",
			prepare: withTransport,
			args:    args{i: 0, t: renamed},
		},
		{
			name:    "a transport in use cannot be renamed",
			prepare: withDatabaseBehind,
			args:    args{i: 0, t: renamed},
			wantErr: &TransportInUseError{},
		},
		{
			name:    "a transport in use can change its settings",
			prepare: withDatabaseBehind,
			args:    args{i: 0, t: tuned},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := testWorkspace(t)
			tt.prepare(t, w)

			err := w.Transports.Edit(tt.args.i, tt.args.t)

			if tt.wantErr != nil {
				assertErrorOfType(t, tt.wantErr, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestTransportRepository_Remove(t *testing.T) {
	type args struct {
		i int
	}
	tests := []struct {
		name     string
		prepare  func(*testing.T, linkedWorkspace)
		args     args
		wantLen  int
		wantUsed []Usage
	}{
		{
			name:    "an unused transport goes",
			prepare: withTransport,
			args:    args{i: 0},
			wantLen: 0,
		},
		{
			name:     "a transport in use stays",
			prepare:  withDatabaseBehind,
			args:     args{i: 0},
			wantLen:  1,
			wantUsed: []Usage{{Kind: "postgres", Owner: "orders"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := testWorkspace(t)
			tt.prepare(t, w)

			err := w.Transports.Remove(tt.args.i)

			assert.Equal(t, tt.wantLen, w.Transports.Len())
			if tt.wantUsed != nil {
				inUse, ok := errors.AsType[*TransportInUseError](err)
				require.Truef(t, ok, "error = %v, want *TransportInUseError", err)
				assert.Equal(t, tt.wantUsed, inUse.By)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestConnectionRepository_Add_transport(t *testing.T) {
	type args struct {
		cfg config.Connection
	}
	tests := []struct {
		name     string
		prepare  func(*testing.T, linkedWorkspace)
		args     args
		wantConn network.Connection
		wantErr  error
	}{
		{
			name:     "a database behind an existing transport goes through it",
			prepare:  withTransport,
			args:     args{cfg: behind("ssh:prod")},
			wantConn: &network.Transported{},
		},
		{
			name:    "a database behind a missing transport",
			args:    args{cfg: behind("ssh:prod")},
			wantErr: &TransportNotFoundError{},
		},
		{
			name:    "a reference whose kind does not match the record",
			prepare: withTransport,
			args:    args{cfg: behind("vpn:prod")},
			wantErr: &TransportNotFoundError{},
		},
		{
			name:    "a reference without a kind",
			prepare: withTransport,
			args:    args{cfg: behind("prod")},
			wantErr: &TransportNotFoundError{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := testWorkspace(t)
			if tt.prepare != nil {
				tt.prepare(t, w)
			}

			err := w.postgres.Add(tt.args.cfg)

			if tt.wantErr != nil {
				assertErrorOfType(t, tt.wantErr, err)
				return
			}
			require.NoError(t, err)
			got, ok := w.postgres.ConnectionAt(0)
			require.True(t, ok)
			assert.IsType(t, tt.wantConn, got)
		})
	}
}

func TestConnectionRepository_Edit_bastion(t *testing.T) {
	type args struct {
		i   int
		cfg config.Connection
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name:    "a bastion in use cannot be renamed",
			args:    args{i: 0, cfg: bastion("jump", "bastion.local")},
			wantErr: &ConnectionInUseError{},
		},
		{
			name: "a bastion in use can move to another host",
			args: args{i: 0, cfg: bastion("bastion", "bastion2.local")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := testWorkspace(t)
			withDatabaseBehind(t, w)

			err := w.ssh.Edit(tt.args.i, tt.args.cfg)

			if tt.wantErr != nil {
				assertErrorOfType(t, tt.wantErr, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestConnectionRepository_Remove_bastion(t *testing.T) {
	type args struct {
		i int
	}
	tests := []struct {
		name     string
		prepare  func(*testing.T, linkedWorkspace)
		args     args
		wantLen  int
		wantUsed []Usage
	}{
		{
			name:    "an unused bastion goes",
			prepare: withBastion,
			args:    args{i: 0},
			wantLen: 0,
		},
		{
			name:     "a bastion in use stays",
			prepare:  withTransport,
			args:     args{i: 0},
			wantLen:  1,
			wantUsed: []Usage{{Kind: "transport", Owner: "prod"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := testWorkspace(t)
			tt.prepare(t, w)

			err := w.ssh.Remove(tt.args.i)

			assert.Equal(t, tt.wantLen, w.ssh.Len())
			if tt.wantUsed != nil {
				inUse, ok := errors.AsType[*ConnectionInUseError](err)
				require.Truef(t, ok, "error = %v, want *ConnectionInUseError", err)
				assert.Equal(t, tt.wantUsed, inUse.By)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestTransportRepository_Resolve(t *testing.T) {
	type args struct {
		target config.Connection
	}
	tests := []struct {
		name     string
		prepare  func(*testing.T, linkedWorkspace)
		between  func(*testing.T, linkedWorkspace)
		args     args
		wantSame bool
		wantErr  error
	}{
		{
			name:     "the same database gets the same transport",
			prepare:  withTransport,
			args:     args{target: behind("ssh:prod")},
			wantSame: true,
		},
		{
			name:    "a moved bastion gets a fresh transport",
			prepare: withTransport,
			between: func(t *testing.T, w linkedWorkspace) {
				t.Helper()

				require.NoError(t, w.ssh.Edit(0, bastion("bastion", "bastion2.local")))
			},
			args: args{target: behind("ssh:prod")},
		},
		{
			name:    "retuned transport settings get a fresh transport",
			prepare: withTransport,
			between: func(t *testing.T, w linkedWorkspace) {
				t.Helper()

				tuned := prodTransport("bastion")
				tuned.KeepAlive = 30
				require.NoError(t, w.Transports.Edit(0, tuned))
			},
			args: args{target: behind("ssh:prod")},
		},
		{
			name: "a transport that vanished after the database was saved",
			prepare: func(t *testing.T, w linkedWorkspace) {
				t.Helper()

				withDatabaseBehind(t, w)
				w.Transports.file.Remove(0)
			},
			args:    args{target: behind("ssh:prod")},
			wantErr: &TransportNotFoundError{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := testWorkspace(t)
			tt.prepare(t, w)

			got, err := w.Transports.Resolve(tt.args.target)
			if tt.wantErr != nil {
				assertErrorOfType(t, tt.wantErr, err)
				return
			}
			require.NoError(t, err)

			if tt.between != nil {
				tt.between(t, w)
			}

			got1, err := w.Transports.Resolve(tt.args.target)
			require.NoError(t, err)
			if tt.wantSame {
				assert.Same(t, got, got1)
				return
			}
			assert.NotSame(t, got, got1)
		})
	}
}

func TestConnectionRepository_ConnectionAt_transport(t *testing.T) {
	tests := []struct {
		name     string
		prepare  func(*testing.T, linkedWorkspace)
		wantConn network.Connection
		wantErr  string
	}{
		{
			name:     "a database behind a transport goes through it",
			prepare:  withDatabaseBehind,
			wantConn: &network.Transported{},
		},
		{
			name: "a database whose transport vanished is never dialled direct",
			prepare: func(t *testing.T, w linkedWorkspace) {
				t.Helper()

				withDatabaseBehind(t, w)
				w.Transports.file.Remove(0)
			},
			wantConn: &network.Transported{},
			wantErr:  `transport "ssh:prod" does not exist`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := testWorkspace(t)
			tt.prepare(t, w)

			got, ok := w.postgres.ConnectionAt(0)
			require.True(t, ok)
			assert.IsType(t, tt.wantConn, got)
			if tt.wantErr == "" {
				return
			}

			_, err := got.Ping(t.Context())
			var opErr *network.OpError
			require.ErrorAs(t, err, &opErr)
			assert.Equal(t, network.TransportErrorCode, opErr.Code)
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}
