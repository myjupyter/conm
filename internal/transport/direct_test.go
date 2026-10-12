package transport

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
)

func TestDirect_Open(t *testing.T) {
	t.Parallel()

	type fields struct {
		target config.Connection
	}
	tests := []struct {
		name   string
		fields fields
		want   string
		want1  int
	}{
		{
			name:   "hands back the address of the connection",
			fields: fields{target: config.Postgres{Hostname: "db.internal", PortNumber: 5432}},
			want:   "db.internal",
			want1:  5432,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			d := NewDirect(tt.fields.target)

			got, got1, err := d.Open(t.Context())
			require.NoError(t, err)

			is := assert.New(t)
			is.Equal(tt.want, got)
			is.Equal(tt.want1, got1)
			is.NoError(d.Close(t.Context()))
		})
	}
}
