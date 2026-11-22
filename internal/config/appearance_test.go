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
			name: "a hex accent is kept",
			args: args{toml: "[conm.appearance]\naccent = \"#A1b2C3\"\n"},
			want: AppearanceSettings{Border: "single", Accent: "#A1b2C3", Glyphs: "unicode"},
		},
		{
			name: "a malformed hex accent falls back to the default",
			args: args{toml: "[conm.appearance]\naccent = \"#abc\"\n"},
			want: AppearanceSettings{Border: "single", Accent: "green", Glyphs: "unicode"},
		},
		{
			name: "rune overrides are kept, bad ones dropped",
			args: args{toml: "[conm.appearance.border_runes]\ntop_left = \"╭\"\nwide = \"x\"\nvertical = \"||\"\n[conm.appearance.glyph_runes]\ncaret = \"→\"\nstatus_ok = \"好\"\n"},
			want: AppearanceSettings{
				Border: "single", Accent: "green", Glyphs: "unicode",
				BorderRunes: map[string]string{"top_left": "╭"},
				GlyphRunes:  map[string]string{"caret": "→"},
			},
		},
		{
			name: "unknown values fall back to the default",
			args: args{toml: "[conm.appearance]\nborder = \"wavy\"\naccent = \"blue\"\nglyphs = \"emoji\"\n"},
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

func TestAppearanceSettings_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		a       AppearanceSettings
		wantErr bool
	}{
		{name: "presets", a: AppearanceSettings{Border: "heavy", Accent: "teal", Glyphs: "nerd"}},
		{name: "hex accent", a: AppearanceSettings{Border: "single", Accent: "#ff8800", Glyphs: "ascii"}},
		{name: "short hex", a: AppearanceSettings{Border: "single", Accent: "#f80", Glyphs: "ascii"}, wantErr: true},
		{name: "hex without hash", a: AppearanceSettings{Border: "single", Accent: "ff8800", Glyphs: "ascii"}, wantErr: true},
		{name: "unknown border", a: AppearanceSettings{Border: "wavy", Accent: "green", Glyphs: "unicode"}, wantErr: true},
		{name: "rune overrides", a: AppearanceSettings{Border: "single", Accent: "green", Glyphs: "unicode",
			BorderRunes: map[string]string{"top_left": "╭"}, GlyphRunes: map[string]string{"caret": "»"}}},
		{name: "unknown border element", a: AppearanceSettings{Border: "single", Accent: "green", Glyphs: "unicode",
			BorderRunes: map[string]string{"middle": "+"}}, wantErr: true},
		{name: "two-cell glyph", a: AppearanceSettings{Border: "single", Accent: "green", Glyphs: "unicode",
			GlyphRunes: map[string]string{"caret": "好"}}, wantErr: true},
		{name: "two-rune glyph", a: AppearanceSettings{Border: "single", Accent: "green", Glyphs: "unicode",
			GlyphRunes: map[string]string{"caret": "->"}}, wantErr: true},
		{name: "unknown glyphs", a: AppearanceSettings{Border: "single", Accent: "green", Glyphs: "emoji"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.a.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
