package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIdentity(t *testing.T) {
	tests := map[string]struct {
		conn Connection
		want string
	}{
		"postgres": {
			conn: Postgres{
				Hostname: "db.internal", PortNumber: 5432, User: "svc", Password: "hunter2",
				DBName: "orders", SchemaName: "public", SSLMode: PostgresSSLModeRequire,
			},
			want: `["postgres" "db.internal" "5432" "svc" "orders" "public"]`,
		},
		"mysql": {
			conn: MySQL{
				Hostname: "db.internal", PortNumber: 3306, User: "svc", Password: "hunter2",
				DBName: "orders", TLSMode: "true",
			},
			want: `["mysql" "db.internal" "3306" "svc" "orders"]`,
		},
		"mssql": {
			conn: MSSQL{
				Hostname: "db.internal", PortNumber: 1433, User: "sa", Password: "hunter2",
				DBName: "master", EncryptMode: "true", TrustCert: true,
			},
			want: `["mssql" "db.internal" "1433" "sa" "master"]`,
		},
		"clickhouse": {
			conn: ClickHouse{
				Hostname: "ch.internal", PortNumber: 9000, User: "default", Password: "hunter2",
				DBName: "events", Secure: true,
			},
			want: `["clickhouse" "ch.internal" "9000" "default" "events"]`,
		},
		"redis": {
			conn: Redis{
				Hostname: "cache.internal", PortNumber: 6379, User: "app", Password: "hunter2",
				DBIndex: 2, TLSMode: "true",
			},
			want: `["redis" "cache.internal" "6379" "app" "2"]`,
		},
		"mongodb": {
			conn: MongoDB{
				Hostname: "mongo.internal", PortNumber: 27017, User: "svc", Password: "hunter2",
				DBName: "orders", AuthSource: "admin", TLS: true,
			},
			want: `["mongodb" "mongo.internal" "27017" "svc" "orders"]`,
		},
		"ssh": {
			conn: SSH{
				Hostname: "10.0.0.5", PortNumber: 22, User: "deploy", Auth: SSHAuthKey,
				Password: "filepath:~/.ssh/id", Jump: " a@hop1:23 , hop2", ForwardAgent: true,
				KeepAlive: 30, LocalForward: "5432:db:5432",
			},
			want: `["ssh" "10.0.0.5" "22" "deploy" "a@hop1:23,hop2" "5432:db:5432"]`,
		},
		"ssh without jump and forward": {
			conn: SSH{Hostname: "10.0.0.5", PortNumber: 22, User: "deploy", Auth: SSHAuthAgent},
			want: `["ssh" "10.0.0.5" "22" "deploy" "" ""]`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.conn.Identity())
		})
	}
}
