package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func defaultAppearance(override func(*AppearanceSettings)) AppearanceSettings {
	a := AppearanceSettings{
		Border: "single", Accent: "green", Glyphs: "unicode", Text: "default",
		BorderColor: "default", GlyphColor: "default", TextColor: "default",
	}
	if override != nil {
		override(&a)
	}
	return a
}

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
			want: defaultAppearance(nil),
		},
		{
			name: "known values are kept",
			args: args{toml: "[conm.appearance]\nborder = \"block\"\naccent = \"pink\"\nglyphs = \"emoji\"\ntext = \"soft\"\n"},
			want: defaultAppearance(func(a *AppearanceSettings) {
				a.Border, a.Accent, a.Glyphs, a.Text = "block", "pink", "emoji", "soft"
			}),
		},
		{
			name: "a hex accent is kept",
			args: args{toml: "[conm.appearance]\naccent = \"#A1b2C3\"\n"},
			want: defaultAppearance(func(a *AppearanceSettings) { a.Accent = "#A1b2C3" }),
		},
		{
			name: "a malformed hex accent falls back to the default",
			args: args{toml: "[conm.appearance]\naccent = \"#abc\"\n"},
			want: defaultAppearance(nil),
		},
		{
			name: "overrides are kept, bad ones dropped",
			args: args{toml: "[conm.appearance.border_runes]\ntop_left = \"╭\"\nwide = \"x\"\nleft = \"||\"\nright = \"好\"\n" +
				"[conm.appearance.glyph_runes]\ncaret = \"👉\"\nstatus_ok = \"✅✅\"\n" +
				"[conm.appearance.text_style]\nnote = \"italic+faint\"\ntitle = \"shiny\"\n"},
			want: defaultAppearance(func(a *AppearanceSettings) {
				a.BorderRunes = map[string]string{"top_left": "╭"}
				a.GlyphRunes = map[string]string{"caret": "👉"}
				a.TextStyle = map[string]string{"note": "italic+faint"}
			}),
		},
		{
			name: "colour palettes and overrides",
			args: args{toml: "[conm.appearance]\nborder_color = \"duotone\"\nglyph_color = \"muted\"\n" +
				"[conm.appearance.border_colors]\ntop = \"#ff0000\"\nleft = \"teal\"\nmiddle = \"red\"\n" +
				"[conm.appearance.glyph_colors]\ncaret = \"accent\"\nstatus_ok = \"neon\"\nellipsis = \"red\"\n" +
				"[conm.appearance.text_colors]\nnote = \"dim\"\nshadow = \"red\"\ntitle = \"teal>pink\"\n" +
				"[conm.appearance.text_backgrounds]\ntab = \"pink\"\nheading = \"#202020>#303030\"\nnote = \"red\"\n"},
			want: defaultAppearance(func(a *AppearanceSettings) {
				a.BorderColor, a.GlyphColor = "duotone", "muted"
				a.BorderColors = map[string]string{"top": "#ff0000", "left": "teal"}
				a.GlyphColors = map[string]string{"caret": "accent"}
				a.TextColors = map[string]string{"note": "dim", "title": "teal>pink"}
				a.TextBackgrounds = map[string]string{"tab": "pink", "heading": "#202020>#303030"}
			}),
		},
		{
			name: "unknown values fall back to the default",
			args: args{toml: "[conm.appearance]\nborder = \"wavy\"\naccent = \"blue\"\nglyphs = \"runes\"\ntext = \"loud\"\n"},
			want: defaultAppearance(func(a *AppearanceSettings) { a.Accent = "blue" }),
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
		name     string
		override func(*AppearanceSettings)
		wantErr  bool
	}{
		{name: "defaults"},
		{name: "presets", override: func(a *AppearanceSettings) { a.Border, a.Glyphs, a.Text = "shade", "emoji", "emphasis" }},
		{name: "hex accent", override: func(a *AppearanceSettings) { a.Accent = "#ff8800" }},
		{name: "short hex", override: func(a *AppearanceSettings) { a.Accent = "#f80" }, wantErr: true},
		{name: "hex without hash", override: func(a *AppearanceSettings) { a.Accent = "ff8800" }, wantErr: true},
		{name: "unknown border", override: func(a *AppearanceSettings) { a.Border = "wavy" }, wantErr: true},
		{name: "unknown glyphs", override: func(a *AppearanceSettings) { a.Glyphs = "runes" }, wantErr: true},
		{name: "unknown text", override: func(a *AppearanceSettings) { a.Text = "loud" }, wantErr: true},
		{name: "overrides", override: func(a *AppearanceSettings) {
			a.BorderRunes = map[string]string{"top_left": "╭", "rule": "▀"}
			a.GlyphRunes = map[string]string{"caret": "»", "status_ok": "✅"}
			a.TextStyle = map[string]string{"note": "italic", "title": "plain", "key": "bold+underline"}
		}},
		{name: "colours", override: func(a *AppearanceSettings) {
			a.BorderColor, a.GlyphColor = "accent", "bright"
			a.BorderColors = map[string]string{"top": "#123456", "rule": "dim"}
			a.GlyphColors = map[string]string{"caret": "auto", "status_ok": "pink"}
		}},
		{name: "gradients", override: func(a *AppearanceSettings) {
			a.BorderColor, a.TextColor = "sunset", "ocean"
			a.BorderColors = map[string]string{"top": "teal>pink", "left": "#ff0000>#00ff00>blue"}
			a.TextColors = map[string]string{"title": "accent>pink"}
		}},
		{name: "accent gradient", override: func(a *AppearanceSettings) { a.Accent = "teal>#ff00aa>pink" }},
		{name: "accent gradient through a name", override: func(a *AppearanceSettings) { a.Accent = "accent>pink" }, wantErr: true},
		{name: "gradient on a glyph", override: func(a *AppearanceSettings) { a.GlyphColors = map[string]string{"caret": "teal>pink"} }, wantErr: true},
		{name: "one-stop gradient", override: func(a *AppearanceSettings) { a.BorderColors = map[string]string{"top": "teal>"} }, wantErr: true},
		{name: "auto in a gradient", override: func(a *AppearanceSettings) { a.BorderColors = map[string]string{"top": "auto>pink"} }, wantErr: true},
		{name: "text colours", override: func(a *AppearanceSettings) {
			a.TextColor = "accent"
			a.TextColors = map[string]string{"title": "#abcdef", "note": "soft"}
		}},
		{name: "text backgrounds", override: func(a *AppearanceSettings) {
			a.TextBackgrounds = map[string]string{"tab": "teal>pink", "heading": "#111111"}
		}},
		{name: "background for a role without one", override: func(a *AppearanceSettings) { a.TextBackgrounds = map[string]string{"note": "red"} }, wantErr: true},
		{name: "unknown text palette", override: func(a *AppearanceSettings) { a.TextColor = "neon" }, wantErr: true},
		{name: "bad text colour", override: func(a *AppearanceSettings) { a.TextColors = map[string]string{"note": "dark"} }, wantErr: true},
		{name: "unknown border palette", override: func(a *AppearanceSettings) { a.BorderColor = "neon" }, wantErr: true},
		{name: "unknown glyph palette", override: func(a *AppearanceSettings) { a.GlyphColor = "neon" }, wantErr: true},
		{name: "bad border colour", override: func(a *AppearanceSettings) { a.BorderColors = map[string]string{"top": "#12"} }, wantErr: true},
		{name: "glyph colour for text-bound glyph", override: func(a *AppearanceSettings) { a.GlyphColors = map[string]string{"ellipsis": "red"} }, wantErr: true},
		{name: "unknown border element", override: func(a *AppearanceSettings) { a.BorderRunes = map[string]string{"middle": "+"} }, wantErr: true},
		{name: "wide border rune", override: func(a *AppearanceSettings) { a.BorderRunes = map[string]string{"left": "█好"} }, wantErr: true},
		{name: "two-column border rune", override: func(a *AppearanceSettings) { a.BorderRunes = map[string]string{"left": "好"} }, wantErr: true},
		{name: "two glyphs", override: func(a *AppearanceSettings) { a.GlyphRunes = map[string]string{"caret": "->"} }, wantErr: true},
		{name: "empty glyph", override: func(a *AppearanceSettings) { a.GlyphRunes = map[string]string{"caret": ""} }, wantErr: true},
		{name: "unknown attribute", override: func(a *AppearanceSettings) { a.TextStyle = map[string]string{"note": "blink"} }, wantErr: true},
		{name: "repeated attribute", override: func(a *AppearanceSettings) { a.TextStyle = map[string]string{"note": "bold+bold"} }, wantErr: true},
		{name: "plain with attribute", override: func(a *AppearanceSettings) { a.TextStyle = map[string]string{"note": "plain+bold"} }, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := defaultAppearance(tt.override).Validate()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestIsGradient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value string
		want  bool
	}{
		{value: "teal>pink", want: true},
		{value: "#ff0000>accent>#0000ff", want: true},
		{value: "teal"},
		{value: "teal>"},
		{value: ">pink"},
		{value: "auto>pink"},
		{value: "teal>neon"},
		{value: "teal > pink"},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, IsGradient(tt.value))
		})
	}
}

func TestIsGlyph(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "narrow symbol", value: "❯", want: true},
		{name: "emoji", value: "✅", want: true},
		{name: "emoji with skin tone", value: "👍🏽", want: true},
		{name: "two characters", value: "ab"},
		{name: "two emoji", value: "✅✅"},
		{name: "empty", value: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, IsGlyph(tt.value))
		})
	}
}
