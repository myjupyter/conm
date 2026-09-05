package repository

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/myjupyter/conm/internal/config"
)

func testRepo(t *testing.T) *ConnectionRepository {
	t.Helper()

	conm := config.Conm{Databases: []config.Database{
		{Type: config.PostgresConnType, CLI: "psql", Enabled: true},
	}}

	repo, err := openConnections[*config.PostgresConfigWrapper](
		conm,
		config.PostgresConnType,
		filepath.Join(t.TempDir(), "postgres.toml"),
	)
	if err != nil {
		t.Fatalf("openConnections failed: %v", err)
	}
	t.Cleanup(func() { repo.Close() })

	return repo
}

func pg(name, host, db string) config.Postgres {
	return config.Postgres{
		Meta:       config.ConnMeta{Name: name},
		Hostname:   host,
		PortNumber: 5432,
		User:       "me",
		DBName:     db,
		SSLMode:    config.PostgresSSLModePrefer,
	}
}

func TestAddRejectsTheSameEndpoint(t *testing.T) {
	repo := testRepo(t)

	if err := repo.Add(pg("first", "db.example.com", "metrics")); err != nil {
		t.Fatalf("first add failed: %v", err)
	}

	err := repo.Add(pg("second", "db.example.com", "metrics"))

	dup, ok := errors.AsType[*DuplicateConnectionError](err)
	if !ok {
		t.Fatalf("error = %v, want *DuplicateConnectionError", err)
	}
	if dup.Name != "first" || dup.Target != "me@db.example.com:5432/metrics" {
		t.Errorf("duplicate = %+v, want the first connection", dup)
	}
	if repo.Len() != 1 {
		t.Errorf("repository holds %d connections, want 1", repo.Len())
	}
}

func TestAddRejectsTheSameName(t *testing.T) {
	repo := testRepo(t)

	if err := repo.Add(pg("prod", "db1.example.com", "metrics")); err != nil {
		t.Fatalf("first add failed: %v", err)
	}

	err := repo.Add(pg("prod", "db2.example.com", "other"))

	if _, ok := errors.AsType[*DuplicateNameError](err); !ok {
		t.Fatalf("error = %v, want *DuplicateNameError", err)
	}
	if repo.Len() != 1 {
		t.Errorf("repository holds %d connections, want 1", repo.Len())
	}
}

func TestAddAcceptsADifferentEndpoint(t *testing.T) {
	repo := testRepo(t)

	for _, conn := range []config.Postgres{
		pg("", "db.example.com", "metrics"),
		pg("", "db.example.com", "billing"),
		pg("", "replica.example.com", "metrics"),
	} {
		if err := repo.Add(conn); err != nil {
			t.Fatalf("add %s failed: %v", conn.DBName, err)
		}
	}

	if repo.Len() != 3 {
		t.Errorf("repository holds %d connections, want 3", repo.Len())
	}
}

func TestEditDoesNotCollideWithItself(t *testing.T) {
	repo := testRepo(t)

	if err := repo.Add(pg("prod", "db.example.com", "metrics")); err != nil {
		t.Fatalf("add failed: %v", err)
	}

	edited := pg("prod", "db.example.com", "metrics")
	edited.SSLMode = config.PostgresSSLModeRequire
	if err := repo.Edit(0, edited); err != nil {
		t.Fatalf("editing a connection into itself failed: %v", err)
	}
}

func TestEditRejectsAnotherRowsEndpoint(t *testing.T) {
	repo := testRepo(t)

	if err := repo.Add(pg("first", "db1.example.com", "metrics")); err != nil {
		t.Fatalf("first add failed: %v", err)
	}
	if err := repo.Add(pg("second", "db2.example.com", "metrics")); err != nil {
		t.Fatalf("second add failed: %v", err)
	}

	err := repo.Edit(1, pg("second", "db1.example.com", "metrics"))

	if _, ok := errors.AsType[*DuplicateConnectionError](err); !ok {
		t.Fatalf("error = %v, want *DuplicateConnectionError", err)
	}
}
