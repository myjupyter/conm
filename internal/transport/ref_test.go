package transport

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/myjupyter/conm/internal/config"
)

func TestRef(t *testing.T) {
	t.Parallel()

	type args struct {
		kind config.TransportKind
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "ssh transport",
			args: args{kind: config.SSHTransportKind, name: "prod-bastion"},
			want: "ssh:prod-bastion",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, Ref(tt.args.kind, tt.args.name))
		})
	}
}

func TestParseRef(t *testing.T) {
	t.Parallel()

	type args struct {
		raw string
	}
	tests := []struct {
		name  string
		args  args
		want  config.TransportKind
		want1 string
		want2 bool
	}{
		{
			name:  "ssh transport",
			args:  args{raw: "ssh:prod-bastion"},
			want:  config.SSHTransportKind,
			want1: "prod-bastion",
			want2: true,
		},
		{
			name: "bare name",
			args: args{raw: "prod-bastion"},
		},
		{
			name: "no name",
			args: args{raw: "ssh:"},
		},
		{
			name: "no kind",
			args: args{raw: ":prod"},
		},
		{
			name: "empty",
			args: args{raw: ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, got1, got2 := ParseRef(tt.args.raw)

			is := assert.New(t)
			is.Equal(tt.want, got)
			is.Equal(tt.want1, got1)
			is.Equal(tt.want2, got2)
		})
	}
}
