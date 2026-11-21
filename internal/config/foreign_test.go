package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "file")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	return path
}

func TestPGPass_Password(t *testing.T) {
	const file = `# hostname:port:database:username:password
db.internal:5432:orders:app:first
db.internal:5432:orders:app:shadowed
db.internal:*:*:admin:any-db
*:*:*:colon\:user:with\:colon\\slash
too:few:fields
`
	type args struct {
		host     string
		port     string
		database string
		user     string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "the first matching row wins",
			args: args{host: "db.internal", port: "5432", database: "orders", user: "app"},
			want: "first",
		},
		{
			name: "wildcards match any value",
			args: args{host: "db.internal", port: "6543", database: "billing", user: "admin"},
			want: "any-db",
		},
		{
			name: "escaped colons and backslashes are part of the field",
			args: args{host: "elsewhere", port: "5432", database: "x", user: "colon:user"},
			want: `with:colon\slash`,
		},
		{
			name: "no row for the connection",
			args: args{host: "db.internal", port: "5432", database: "orders", user: "nobody"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pass, err := ReadPGPass(writeTemp(t, file))
			require.NoError(t, err)
			assert.Equal(t, tt.want, pass.Password(tt.args.host, tt.args.port, tt.args.database, tt.args.user))
		})
	}
}

func TestReadPGPass_missingFile(t *testing.T) {
	pass, err := ReadPGPass(filepath.Join(t.TempDir(), "absent"))
	require.NoError(t, err)
	assert.Empty(t, pass.Password("db.internal", "5432", "orders", "app"))
}
