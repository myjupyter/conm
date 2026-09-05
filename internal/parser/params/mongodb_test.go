package params

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

func TestParseMongoDB(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		syntax   Syntax
		client   string
		conn     config.MongoDB
		warnings []string
	}{
		{
			name:   "a uri is the positional the shell takes",
			input:  `mongosh "mongodb://app:s3cr3t@mongo.internal:27017/records?authSource=admin"`,
			syntax: URISyntax,
			client: "mongosh",
			conn: config.MongoDB{
				Hostname:   "mongo.internal",
				PortNumber: 27017,
				User:       "app",
				Password:   secret.Ref(secret.Literal, "s3cr3t"),
				DBName:     "records",
				AuthSource: "admin",
			},
		},
		{
			name:   "the srv scheme is tls and a lookup conm cannot do",
			input:  `mongosh "mongodb+srv://app@cluster.example.net/records"`,
			syntax: URISyntax,
			client: "mongosh",
			conn: config.MongoDB{
				Hostname:   "cluster.example.net",
				PortNumber: 27017,
				User:       "app",
				DBName:     "records",
				TLS:        true,
			},
			warnings: []string{mongodbSRVWarning},
		},
		{
			name:   "a bare positional is a database",
			input:  `mongosh records`,
			syntax: FlagSyntax,
			client: "mongosh",
			conn: config.MongoDB{
				PortNumber: 27017,
				DBName:     "records",
			},
		},
		{
			name:   "an address positional carries host, port and database",
			input:  `mongosh mongo.internal:27017/records`,
			syntax: FlagSyntax,
			client: "mongosh",
			conn: config.MongoDB{
				Hostname:   "mongo.internal",
				PortNumber: 27017,
				DBName:     "records",
			},
		},
		{
			name:   "an address positional may be host and port only",
			input:  `mongosh mongo.internal:27018`,
			syntax: FlagSyntax,
			client: "mongosh",
			conn: config.MongoDB{
				Hostname:   "mongo.internal",
				PortNumber: 27018,
			},
		},
		{
			name:   "flags carry the whole connection",
			input:  `mongosh --host mongo.internal --port 27018 -u app -p s3cr3t --authenticationDatabase admin`,
			syntax: FlagSyntax,
			client: "mongosh",
			conn: config.MongoDB{
				Hostname:   "mongo.internal",
				PortNumber: 27018,
				User:       "app",
				Password:   secret.Ref(secret.Literal, "s3cr3t"),
				AuthSource: "admin",
			},
		},
		{
			name:   "the legacy shell takes -p only attached",
			input:  `mongo --host mongo.internal -p records`,
			syntax: FlagSyntax,
			client: "mongo",
			conn: config.MongoDB{
				Hostname:   "mongo.internal",
				PortNumber: 27017,
				DBName:     "records",
			},
		},
		{
			name:   "mongosh reads -p as a value",
			input:  `mongosh --host mongo.internal -p s3cr3t records`,
			syntax: FlagSyntax,
			client: "mongosh",
			conn: config.MongoDB{
				Hostname:   "mongo.internal",
				PortNumber: 27017,
				Password:   secret.Ref(secret.Literal, "s3cr3t"),
				DBName:     "records",
			},
		},
		{
			name:   "the tls flag takes no value",
			input:  `mongosh --host mongo.internal --tls records`,
			syntax: FlagSyntax,
			client: "mongosh",
			conn: config.MongoDB{
				Hostname:   "mongo.internal",
				PortNumber: 27017,
				DBName:     "records",
				TLS:        true,
			},
		},
		{
			name:   "the legacy ssl flag is the same field",
			input:  `mongo --host mongo.internal --ssl records`,
			syntax: FlagSyntax,
			client: "mongo",
			conn: config.MongoDB{
				Hostname:   "mongo.internal",
				PortNumber: 27017,
				DBName:     "records",
				TLS:        true,
			},
		},
		{
			name:   "flags win over the uri they follow",
			input:  `mongosh "mongodb://app@mongo.internal:27017/records" --host replica.internal --port 27018`,
			syntax: URISyntax,
			client: "mongosh",
			conn: config.MongoDB{
				Hostname:   "replica.internal",
				PortNumber: 27018,
				User:       "app",
				DBName:     "records",
			},
		},
		{
			name:   "a script is not a connection parameter",
			input:  `mongosh --host mongo.internal records --eval "db.stats()" --quiet`,
			syntax: FlagSyntax,
			client: "mongosh",
			conn: config.MongoDB{
				Hostname:   "mongo.internal",
				PortNumber: 27017,
				DBName:     "records",
			},
		},
		{
			name:   "a replica set in the host flag has no field to keep it in",
			input:  `mongosh --host "rs0/node1:27017,node2:27017" records`,
			syntax: FlagSyntax,
			client: "mongosh",
			conn: config.MongoDB{
				Hostname:   "node1",
				PortNumber: 27017,
				DBName:     "records",
			},
			warnings: []string{
				`replica set "rs0" has no field to keep it in`,
				"kept only the first of 2 hosts: node1:27017",
			},
		},
		{
			name:   "a replica set parameter has no field to keep it in",
			input:  `mongosh "mongodb://node1:27017,node2:27017/records?replicaSet=rs0"`,
			syntax: URISyntax,
			client: "mongosh",
			conn: config.MongoDB{
				Hostname:   "node1",
				PortNumber: 27017,
				DBName:     "records",
			},
			warnings: []string{
				"kept only the first of 2 hosts: node1:27017",
				`replica set "rs0" has no field to keep it in`,
			},
		},
		{
			name:   "a parameter with no field is warned about",
			input:  `mongosh "mongodb://mongo.internal/records?authSource=admin&readPreference=secondary"`,
			syntax: URISyntax,
			client: "mongosh",
			conn: config.MongoDB{
				Hostname:   "mongo.internal",
				PortNumber: 27017,
				DBName:     "records",
				AuthSource: "admin",
			},
			warnings: []string{"ignored parameter readPreference=secondary"},
		},
		{
			name:   "an unreadable port is warned about",
			input:  `mongosh --host mongo.internal --port twenty`,
			syntax: FlagSyntax,
			client: "mongosh",
			conn: config.MongoDB{
				Hostname:   "mongo.internal",
				PortNumber: 27017,
			},
			warnings: []string{`port "twenty" is not a number, kept 27017`},
		},
		{
			name:   "an unreadable tls parameter is warned about",
			input:  `mongosh "mongodb://mongo.internal/records?tls=sure"`,
			syntax: URISyntax,
			client: "mongosh",
			conn: config.MongoDB{
				Hostname:   "mongo.internal",
				PortNumber: 27017,
				DBName:     "records",
			},
			warnings: []string{`tls "sure" is not a boolean, kept false`},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			is, must := assert.New(t), require.New(t)

			res, err := Parse(config.MongoDBConnType, test.input)
			must.NoError(err)

			is.Equal(test.syntax, res.Syntax)
			is.Equal(test.client, res.Client)
			is.Equal(test.warnings, res.Warnings)

			conn, ok := res.Conn.(config.MongoDB)
			must.Truef(ok, "connection is a %T, want config.MongoDB", res.Conn)
			is.Equal(test.conn, conn)
		})
	}
}

func TestParseMongoDBRoundTripsConnectionString(t *testing.T) {
	want := config.MongoDB{
		Hostname:   "mongo.example.com",
		PortNumber: 27017,
		User:       "app",
		DBName:     "records",
		AuthSource: "admin",
		TLS:        true,
	}

	res, err := Parse(config.MongoDBConnType, want.ConnectionString(""))
	require.NoError(t, err)

	got, ok := res.Conn.(config.MongoDB)
	require.Truef(t, ok, "connection is a %T, want config.MongoDB", res.Conn)
	assert.Equal(t, want, got)
}

func TestParseMongoDBErrors(t *testing.T) {
	tests := []struct {
		name  string
		kind  config.ConnType
		input string
		want  error
	}{
		{
			name:  "a client of another database",
			kind:  config.MongoDBConnType,
			input: `psql -h localhost mydb`,
			want:  ErrTypeMismatch,
		},
		{
			name:  "a connection string of another database",
			kind:  config.MongoDBConnType,
			input: `redis://cache.internal:6379/0`,
			want:  ErrTypeMismatch,
		},
		{
			name:  "a mongodb url handed to another database",
			kind:  config.RedisConnType,
			input: `mongodb://app@mongo.internal:27017/records`,
			want:  ErrTypeMismatch,
		},
		{
			name:  "a ca has no field to keep it in",
			kind:  config.MongoDBConnType,
			input: `mongosh --host mongo.internal --tlsCAFile /etc/ca.pem records`,
			want:  ErrUnsupportedTLS,
		},
		{
			name:  "an insecure parameter is refused whatever its case",
			kind:  config.MongoDBConnType,
			input: `mongosh "mongodb://mongo.internal/records?tlsInsecure=true"`,
			want:  ErrUnsupportedTLS,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Parse(test.kind, test.input)
			assert.ErrorIs(t, err, test.want)
		})
	}
}

func TestParseArgsMongoDB(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want config.MongoDB
	}{
		{
			name: "bare flags need no client name",
			args: []string{"--host", "mongo.internal", "--port", "27018", "-u", "app"},
			want: config.MongoDB{
				Hostname:   "mongo.internal",
				PortNumber: 27018,
				User:       "app",
			},
		},
		{
			name: "a client name is still stripped",
			args: []string{"mongosh", "mongodb://mongo.internal/records"},
			want: config.MongoDB{
				Hostname:   "mongo.internal",
				PortNumber: 27017,
				DBName:     "records",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := ParseArgs(config.MongoDBConnType, test.args)
			require.NoError(t, err)

			conn, ok := res.Conn.(config.MongoDB)
			require.Truef(t, ok, "connection is a %T, want config.MongoDB", res.Conn)
			assert.Equal(t, test.want, conn)
		})
	}
}
