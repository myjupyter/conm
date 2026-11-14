package repository

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/myjupyter/conm/internal/cli"
)

func TestClientCheck_Ready(t *testing.T) {
	type fields struct {
		Client  cli.Info
		Invalid error
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "client installed",
			fields: fields{Client: cli.Info{Name: cli.Psql, Installed: true}},
			want:   true,
		},
		{
			name:   "client not installed",
			fields: fields{Client: cli.Info{Name: cli.Psql}},
			want:   false,
		},
		{
			name:   "no client resolved",
			fields: fields{},
			want:   false,
		},
		{
			name:   "configured client not valid for the type",
			fields: fields{Client: cli.Info{Name: cli.Psql, Installed: true}, Invalid: errors.New("invalid")},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := ClientCheck{Client: tt.fields.Client, Invalid: tt.fields.Invalid}
			assert.Equal(t, tt.want, c.Ready())
		})
	}
}
