package repository

import (
	"errors"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
)

func testRepo(t *testing.T) *ConnectionRepository {
	t.Helper()

	conm := config.Conm{Connections: []config.ConnectionSettings{
		{Type: config.PostgresConnType, CLI: "psql", Enabled: true},
	}}

	repo, err := openConnections[*config.PostgresConfigWrapper](
		conm,
		config.PostgresConnType,
		filepath.Join(t.TempDir(), "postgres.toml"),
	)
	require.NoError(t, err)
	t.Cleanup(func() { repo.Close() })

	return repo
}

func pg(name, host, db string) config.Postgres {
	return config.Postgres{
		Metadata:   config.ConnMeta{Name: name},
		Hostname:   host,
		PortNumber: 5432,
		User:       "me",
		DBName:     db,
		SSLMode:    config.PostgresSSLModePrefer,
	}
}

func TestAddRejectsTheSameEndpoint(t *testing.T) {
	repo := testRepo(t)

	require.NoError(t, repo.Add(pg("first", "db.example.com", "metrics")))

	err := repo.Add(pg("second", "db.example.com", "metrics"))

	dup, ok := errors.AsType[*DuplicateConnectionError](err)
	require.Truef(t, ok, "error = %v, want *DuplicateConnectionError", err)
	assert.Equal(t, "first", dup.Name)
	assert.Equal(t, "postgresql://me@db.example.com:5432/metrics?sslmode=prefer", dup.Target)
	assert.Equal(t, 1, repo.Len())
}

func TestAddRejectsTheSameName(t *testing.T) {
	repo := testRepo(t)

	require.NoError(t, repo.Add(pg("prod", "db1.example.com", "metrics")))

	err := repo.Add(pg("prod", "db2.example.com", "other"))

	_, ok := errors.AsType[*DuplicateNameError](err)
	require.Truef(t, ok, "error = %v, want *DuplicateNameError", err)
	assert.Equal(t, 1, repo.Len())
}

func TestAddAcceptsADifferentEndpoint(t *testing.T) {
	repo := testRepo(t)

	for _, conn := range []config.Postgres{
		pg("", "db.example.com", "metrics"),
		pg("", "db.example.com", "billing"),
		pg("", "replica.example.com", "metrics"),
	} {
		require.NoErrorf(t, repo.Add(conn), "adding %s failed", conn.DBName)
	}

	assert.Equal(t, 3, repo.Len())
}

func TestEditDoesNotCollideWithItself(t *testing.T) {
	repo := testRepo(t)

	require.NoError(t, repo.Add(pg("prod", "db.example.com", "metrics")))

	edited := pg("prod", "db.example.com", "metrics")
	edited.SSLMode = config.PostgresSSLModeRequire

	assert.NoError(t, repo.Edit(0, edited), "editing a connection into itself must be allowed")
}

func TestEditRejectsAnotherRowsEndpoint(t *testing.T) {
	repo := testRepo(t)

	require.NoError(t, repo.Add(pg("first", "db1.example.com", "metrics")))
	require.NoError(t, repo.Add(pg("second", "db2.example.com", "metrics")))

	err := repo.Edit(1, pg("second", "db1.example.com", "metrics"))

	_, ok := errors.AsType[*DuplicateConnectionError](err)
	assert.Truef(t, ok, "error = %v, want *DuplicateConnectionError", err)
}

func TestSSHUniqueness(t *testing.T) {
	target := func(jump, forward string) config.SSH {
		return config.SSH{
			Hostname:     "10.0.0.5",
			PortNumber:   22,
			User:         "deploy",
			Auth:         config.SSHAuthAgent,
			Jump:         jump,
			LocalForward: forward,
		}
	}

	tests := map[string]struct {
		existing, added config.SSH
		duplicate       bool
	}{
		"same jump":                {existing: target("bastion-eu", ""), added: target("bastion-eu", ""), duplicate: true},
		"different jump":           {existing: target("bastion-eu", ""), added: target("bastion-us", "")},
		"jump against direct":      {existing: target("", ""), added: target("bastion-eu", "")},
		"jump differing by spaces": {existing: target("a@hop1,hop2", ""), added: target(" a@hop1 , hop2", ""), duplicate: true},
		"same local forward":       {existing: target("", "5432:db:5432"), added: target("", "5432:db:5432"), duplicate: true},
		"different local forward":  {existing: target("", "5432:db:5432"), added: target("", "6379:cache:6379")},
		"forward against plain":    {existing: target("", ""), added: target("", "5432:db:5432")},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := validateUnique(slices.Values([]config.Connection{tt.existing}), tt.added)

			if !tt.duplicate {
				assert.NoError(t, err)
				return
			}
			_, ok := errors.AsType[*DuplicateConnectionError](err)
			assert.Truef(t, ok, "error = %v, want *DuplicateConnectionError", err)
		})
	}
}
