package importer

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadINI(t *testing.T) {
	type args struct {
		content string
	}
	tests := []struct {
		name string
		args args
		want []Section
	}{
		{
			name: "empty file",
		},
		{
			name: "only comments and blank lines",
			args: args{content: "# a comment\n\n; another\n   \n"},
		},
		{
			name: "sections keep their order and their pairs",
			args: args{content: "[orders]\nhost=db.internal\nport=5433\n\n[billing]\ndbname=billing\n"},
			want: []Section{
				{Name: "orders", Params: []Param{{Key: "host", Value: "db.internal"}, {Key: "port", Value: "5433"}}},
				{Name: "billing", Params: []Param{{Key: "dbname", Value: "billing"}}},
			},
		},
		{
			name: "a section without pairs is still a section",
			args: args{content: "[orders]\n[billing]\n"},
			want: []Section{{Name: "orders"}, {Name: "billing"}},
		},
		{
			name: "the last line needs no newline",
			args: args{content: "[orders]\nhost=db.internal"},
			want: []Section{
				{Name: "orders", Params: []Param{{Key: "host", Value: "db.internal"}}},
			},
		},
		{
			name: "comments between pairs are dropped",
			args: args{content: "[orders]\n# the host\nhost=db.internal\n; the port\nport=5433\n"},
			want: []Section{
				{Name: "orders", Params: []Param{{Key: "host", Value: "db.internal"}, {Key: "port", Value: "5433"}}},
			},
		},
		{
			name: "spaces around a section name, a key and a value are dropped",
			args: args{content: "  [ orders ]  \n\thost = db.internal  \n"},
			want: []Section{
				{Name: "orders", Params: []Param{{Key: "host", Value: "db.internal"}}},
			},
		},
		{
			name: "windows line endings",
			args: args{content: "[orders]\r\nhost=db.internal\r\n"},
			want: []Section{
				{Name: "orders", Params: []Param{{Key: "host", Value: "db.internal"}}},
			},
		},
		{
			name: "pairs before the first section belong to nothing",
			args: args{content: "host=stray\n[orders]\n"},
			want: []Section{{Name: "orders"}},
		},
		{
			name: "a value keeps its own equals signs and inner spaces",
			args: args{content: "[orders]\noptions=-c search_path=app\n"},
			want: []Section{
				{Name: "orders", Params: []Param{{Key: "options", Value: "-c search_path=app"}}},
			},
		},
		{
			name: "a bare key and a key with nothing after the equals sign have no value",
			args: args{content: "[orders]\nssl\npassword=\n"},
			want: []Section{
				{Name: "orders", Params: []Param{{Key: "ssl"}, {Key: "password"}}},
			},
		},
		{
			name: "a repeated key is kept every time, in order",
			args: args{content: "[orders]\nhost=first\nhost=second\n"},
			want: []Section{
				{Name: "orders", Params: []Param{{Key: "host", Value: "first"}, {Key: "host", Value: "second"}}},
			},
		},
		{
			name: "a repeated section is kept every time, not merged",
			args: args{content: "[orders]\nhost=first\n[orders]\nport=5433\n"},
			want: []Section{
				{Name: "orders", Params: []Param{{Key: "host", Value: "first"}}},
				{Name: "orders", Params: []Param{{Key: "port", Value: "5433"}}},
			},
		},
		{
			name: "a comment mark inside a value is part of the value",
			args: args{content: "[orders]\npassword=a#b;c\n"},
			want: []Section{
				{Name: "orders", Params: []Param{{Key: "password", Value: "a#b;c"}}},
			},
		},
		{
			name: "quotes are left to the caller",
			args: args{content: "[orders]\npassword=\"s3cret\"\n"},
			want: []Section{
				{Name: "orders", Params: []Param{{Key: "password", Value: `"s3cret"`}}},
			},
		},
		{
			name: "a bracket that is not closed is a key, not a section",
			args: args{content: "[orders]\n[billing\n"},
			want: []Section{
				{Name: "orders", Params: []Param{{Key: "[billing"}}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadINI(writeTemp(t, tt.args.content))
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestReadINI_missingFile(t *testing.T) {
	got, err := ReadINI(filepath.Join(t.TempDir(), "absent"))
	require.ErrorIs(t, err, fs.ErrNotExist)
	assert.Nil(t, got)
}

func writeTemp(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "file")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	return path
}
