package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransportValidate(t *testing.T) {
	valid := Transport{Name: "prod", Kind: SSHTransportKind, Via: "bastion", ConnectTimeout: 5, KeepAlive: 30}

	tests := map[string]struct {
		edit    func(*Transport)
		wantErr string
	}{
		"valid":            {edit: func(*Transport) {}},
		"no name":          {edit: func(t *Transport) { t.Name = "" }, wantErr: "name can't be empty"},
		"unknown kind":     {edit: func(t *Transport) { t.Kind = "vpn" }, wantErr: `unknown transport kind "vpn"`},
		"nothing to go by": {edit: func(t *Transport) { t.Via = "" }, wantErr: "connection to go through"},
		"auto local port":  {edit: func(t *Transport) { t.LocalPort = 0 }},
		"fixed local port": {edit: func(t *Transport) { t.LocalPort = 15432 }},
		"local port out of range": {
			edit:    func(t *Transport) { t.LocalPort = 70000 },
			wantErr: "local port must be between",
		},
		"negative timeout": {edit: func(t *Transport) { t.ConnectTimeout = -1 }, wantErr: "timeout can't be negative"},
		"negative keepalive": {
			edit:    func(t *Transport) { t.KeepAlive = -1 },
			wantErr: "keepalive can't be negative",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			rec := valid
			tt.edit(&rec)

			errs := rec.Validate()
			if tt.wantErr == "" {
				assert.Empty(t, errs)
				return
			}
			require.Len(t, errs, 1)
			assert.ErrorContains(t, errs[0], tt.wantErr)
		})
	}
}

func TestTransportFileRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "transport.toml")
	rec := Transport{Name: "prod", Kind: SSHTransportKind, Via: "bastion", LocalPort: 15432, ConnectTimeout: 5, KeepAlive: 30}

	file, err := OpenConfig[*TransportConfigWrapper](path)
	require.NoError(t, err)
	file.Add(rec)
	require.NoError(t, file.Save())
	require.NoError(t, file.Close())

	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(raw), "[[transport]]")

	reopened, err := OpenConfig[*TransportConfigWrapper](path)
	require.NoError(t, err)
	t.Cleanup(func() { reopened.Close() })

	require.Equal(t, 1, reopened.Len())
	got := reopened.Get(0)
	assert.True(t, got.IsValid())
	assert.Equal(t, rec.Validate(), got.Validate())
	got.validationErrs = nil
	assert.Equal(t, rec, got)
}
