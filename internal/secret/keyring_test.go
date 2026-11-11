package secret

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

func TestKeyringProvider_isolatesWorkspaces(t *testing.T) {
	type args struct {
		location string
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "plain location", args: args{location: "orders"}},
		{name: "path location", args: args{location: "conm/orders"}},
		{name: "nested path location", args: args{location: "prod/eu/orders"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyring.MockInit()
			ctx := context.Background()
			ref := rawRef(Ref(Keyring, tt.args.location))
			work, personal := &KeyringProvider{Workspace: "conm.work"}, &KeyringProvider{Workspace: "conm.personal"}
			require.NoError(t, work.Store(ctx, ref, "work-pw"))
			require.NoError(t, personal.Store(ctx, ref, "personal-pw"))

			require.NoError(t, work.Remove(ctx, ref))

			got, err := personal.Resolve(ctx, ref)
			require.NoError(t, err)
			assert.Equal(t, "personal-pw", got)
			stored, err := keyring.Get("conm.personal", tt.args.location)
			require.NoError(t, err)
			assert.Equal(t, "personal-pw", stored)
			_, err = work.Resolve(ctx, ref)
			assert.Error(t, err)
		})
	}
}
