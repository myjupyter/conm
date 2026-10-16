package repository

import (
	"iter"
	"maps"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/myjupyter/conm/internal/config"
)

type fakeTransportUser struct {
	kind config.ConnType
	refs map[string]string
}

func (f fakeTransportUser) Kind() config.ConnType { return f.kind }

func (f fakeTransportUser) TransportRefs() iter.Seq2[string, string] { return maps.All(f.refs) }

type fakeConnectionUser struct {
	refs map[string]string
}

func (fakeConnectionUser) Kind() string { return "transport" }

func (f fakeConnectionUser) ConnectionRefs() iter.Seq2[string, string] { return maps.All(f.refs) }

func TestTransportIndex_Usages(t *testing.T) {
	t.Parallel()

	type fields struct {
		users []TransportUser
	}
	type args struct {
		ref string
	}
	users := []TransportUser{
		fakeTransportUser{kind: config.PostgresConnType, refs: map[string]string{"orders": "ssh:prod"}},
		fakeTransportUser{kind: config.RedisConnType, refs: map[string]string{"cache": "ssh:staging"}},
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   []Usage
	}{
		{
			name:   "a transport in use",
			fields: fields{users: users},
			args:   args{ref: "ssh:prod"},
			want:   []Usage{{Kind: "postgres", Owner: "orders"}},
		},
		{
			name:   "a transport nobody uses",
			fields: fields{users: users},
			args:   args{ref: "ssh:dev"},
		},
		{
			name:   "no reference",
			fields: fields{users: users},
			args:   args{ref: ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ix := NewTransportIndex(tt.fields.users...)

			assert.Equal(t, tt.want, ix.Usages(tt.args.ref))
		})
	}
}

func TestConnectionIndex_Usages(t *testing.T) {
	t.Parallel()

	type fields struct {
		users []ConnectionUser
	}
	type args struct {
		ref string
	}
	users := []ConnectionUser{
		fakeConnectionUser{refs: map[string]string{"prod": "ssh:bastion"}},
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   []Usage
	}{
		{
			name:   "a connection a transport goes through",
			fields: fields{users: users},
			args:   args{ref: connectionRef(config.SSHConnType, "bastion")},
			want:   []Usage{{Kind: "transport", Owner: "prod"}},
		},
		{
			name:   "a connection nothing goes through",
			fields: fields{users: users},
			args:   args{ref: connectionRef(config.SSHConnType, "jump")},
		},
		{
			name:   "a connection without a name",
			fields: fields{users: users},
			args:   args{ref: connectionRef(config.SSHConnType, "")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ix := NewConnectionIndex(tt.fields.users...)

			assert.Equal(t, tt.want, ix.Usages(tt.args.ref))
		})
	}
}
