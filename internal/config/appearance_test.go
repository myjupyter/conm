package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadConm_appearance(t *testing.T) {
	t.Parallel()

	type args struct {
		toml string
	}
	tests := []struct {
		name string
		args args
		want AppearanceSettings
	}{
		{
			name: "no section takes the defaults",
			args: args{toml: ""},
			want: AppearanceSettings{Border: "single", Accent: "green", Glyphs: "unicode"},
		},
		{
			name: "known values are kept",
			args: args{toml: "[conm.appearance]\nborder = \"rounded\"\naccent = \"pink\"\nglyphs = \"ascii\"\n"},
			want: AppearanceSettings{Border: "rounded", Accent: "pink", Glyphs: "ascii"},
		},
		{
			name: "unknown values fall back to the default",
			args: args{toml: "[conm.appearance]\nborder = \"dotted\"\naccent = \"blue\"\nglyphs = \"emoji\"\n"},
			want: AppearanceSettings{Border: "single", Accent: "blue", Glyphs: "unicode"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ReadConm(writeConm(t, tt.args.toml))
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Appearance)
		})
	}
}
