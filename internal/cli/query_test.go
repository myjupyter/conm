package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryArgs(t *testing.T) {
	t.Parallel()

	type args struct {
		client string
		query  []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "psql joins the query behind -c",
			args: args{client: Psql, query: []string{"SELECT", "*", "FROM", "t;"}},
			want: []string{"-c", "SELECT * FROM t;"},
		},
		{
			name: "mysql takes the query behind -e",
			args: args{client: MySQL, query: []string{"SELECT 1"}},
			want: []string{"-e", "SELECT 1"},
		},
		{
			name: "sqlcmd takes the query behind -Q",
			args: args{client: SQLCmd, query: []string{"SELECT 1"}},
			want: []string{"-Q", "SELECT 1"},
		},
		{
			name: "clickhouse takes the query behind --query",
			args: args{client: ClickHouse, query: []string{"SELECT 1"}},
			want: []string{"--query", "SELECT 1"},
		},
		{
			name: "mongosh evaluates the query",
			args: args{client: Mongosh, query: []string{"db.users.countDocuments()"}},
			want: []string{"--eval", "db.users.countDocuments()"},
		},
		{
			name: "redis-cli keeps the command words apart",
			args: args{client: RedisCLI, query: []string{"SET", "k", "a b"}},
			want: []string{"SET", "k", "a b"},
		},
		{
			name: "ssh runs the remote command as given",
			args: args{client: SSH, query: []string{"uptime"}},
			want: []string{"uptime"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			queryArgsOf, ok := queryArgs[tt.args.client]
			require.True(t, ok)

			got := queryArgsOf(tt.args.query)

			assert.Equal(t, tt.want, got)
		})
	}
}
