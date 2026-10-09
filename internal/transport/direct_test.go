package transport

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
)

func TestDirectHandsBackTheTarget(t *testing.T) {
	d := NewDirect(config.Postgres{Hostname: "db.internal", PortNumber: 5432})

	host, port, err := d.Open(t.Context())

	require.NoError(t, err)
	assert.Equal(t, "db.internal", host)
	assert.Equal(t, 5432, port)
	require.NoError(t, d.Close(t.Context()))
}
