package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui/spec"
)

const (
	testSourceVia       = "via"
	testSourceTransport = "transport"
	testSourceHost      = "host"
)

func sourceFormSpec() spec.FormSpec[config.Connection] {
	return spec.FormSpec[config.Connection]{
		Fields: []spec.FormField{
			{Key: testSourceHost, Label: "Host"},
			{
				Key:    testSourceVia,
				Label:  "Via",
				Kind:   spec.SourceFieldKind,
				Source: &spec.FieldSource{Name: config.SSHConnType.String()},
			},
			{
				Key:      testSourceTransport,
				Label:    "Transport",
				Kind:     spec.SourceFieldKind,
				Property: spec.OptionalFieldProperty,
				Source: &spec.FieldSource{
					Name:        "transport",
					VisibleFunc: spec.SourceAvailable("transport"),
				},
			},
		},
		Sections: []spec.FormSection{
			{Title: "connection", Fields: []spec.FormFieldKey{testSourceHost, testSourceVia}},
			{Title: "route", Fields: []spec.FormFieldKey{testSourceTransport}},
		},
	}
}

func fieldIndex(t *testing.T, m formModel, key spec.FormFieldKey) int {
	t.Helper()

	for i := range m.fields {
		f := &m.fields[i]
		if f.Key == key {
			return i
		}
	}
	require.Failf(t, "no field", "field %q is not in the form", key)
	return -1
}

func TestFormModel_options_source(t *testing.T) {
	t.Parallel()

	type args struct {
		key spec.FormFieldKey
	}
	sources := spec.Sources{
		"ssh":       {"bastion", "jump"},
		"transport": {"ssh:prod"},
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "a required source offers its values",
			args: args{key: testSourceVia},
			want: []string{"bastion", "jump"},
		},
		{
			name: "an optional source offers an empty choice first",
			args: args{key: testSourceTransport},
			want: []string{"", "ssh:prod"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newFormModel(config.PostgresConnType, sourceFormSpec(), "add", nil, false, nil, sources)
			f := m.fields[fieldIndex(t, m, tt.args.key)]

			assert.Equal(t, tt.want, m.options(f))
		})
	}
}

func TestNewFormModel_source(t *testing.T) {
	t.Parallel()

	type args struct {
		initial map[spec.FormFieldKey]spec.FormFieldValue
		sources spec.Sources
	}
	tests := []struct {
		name          string
		args          args
		wantVia       string
		wantTransport string
		wantSections  []string
	}{
		{
			name:          "a required source starts on its first value, an optional one on none",
			args:          args{sources: spec.Sources{"ssh": {"bastion", "jump"}, "transport": {"ssh:prod"}}},
			wantVia:       "bastion",
			wantTransport: "",
			wantSections:  []string{"connection", "route"},
		},
		{
			name:         "a field whose source is empty is not shown, and an emptied section goes with it",
			args:         args{sources: spec.Sources{"ssh": {"bastion"}}},
			wantVia:      "bastion",
			wantSections: []string{"connection"},
		},
		{
			name: "a saved value keeps its field on screen even when the source is empty",
			args: args{
				initial: map[spec.FormFieldKey]spec.FormFieldValue{testSourceVia: "bastion", testSourceTransport: "ssh:prod"},
				sources: spec.Sources{"ssh": {"bastion"}},
			},
			wantVia:       "bastion",
			wantTransport: "ssh:prod",
			wantSections:  []string{"connection", "route"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := newFormModel(config.PostgresConnType, sourceFormSpec(), "add", tt.args.initial, false, nil, tt.args.sources)

			titles := make([]string, 0, len(got.sections))
			for _, s := range got.sections {
				titles = append(titles, s.Title)
			}
			is := assert.New(t)
			is.Equal(tt.wantVia, got.vals[testSourceVia])
			is.Equal(tt.wantTransport, got.vals[testSourceTransport])
			is.Equal(tt.wantSections, titles)
		})
	}
}

func TestFormModel_fieldError_source(t *testing.T) {
	t.Parallel()

	type args struct {
		key   spec.FormFieldKey
		value string
	}
	sources := spec.Sources{"ssh": {"bastion", "jump"}, "transport": {"ssh:prod"}}

	tests := []struct {
		name        string
		args        args
		wantErr     string
		wantMissing bool
	}{
		{
			name: "a value from the source",
			args: args{key: testSourceVia, value: "jump"},
		},
		{
			name:        "a value the source no longer has",
			args:        args{key: testSourceVia, value: "old-bastion"},
			wantErr:     "old-bastion is no longer available",
			wantMissing: true,
		},
		{
			name:    "a required source left empty",
			args:    args{key: testSourceVia, value: ""},
			wantErr: "via is required",
		},
		{
			name: "an optional source left empty",
			args: args{key: testSourceTransport, value: ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newFormModel(config.PostgresConnType, sourceFormSpec(), "edit", nil, true, nil, sources)
			m.vals[tt.args.key] = tt.args.value
			i := fieldIndex(t, m, tt.args.key)

			got := m.fieldError(i)

			assert.Equal(t, tt.wantMissing, m.isMissing(m.fields[i]))
			if tt.wantErr == "" {
				assert.Empty(t, got)
				return
			}
			assert.Contains(t, got, tt.wantErr)
		})
	}
}

func TestFormModel_navCycle_source(t *testing.T) {
	t.Parallel()

	type args struct {
		key   spec.FormFieldKey
		from  string
		steps int
	}
	sources := spec.Sources{"ssh": {"bastion", "jump"}, "transport": {"ssh:prod"}}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "steps through the source values",
			args: args{key: testSourceVia, from: "bastion", steps: 1},
			want: "jump",
		},
		{
			name: "wraps around",
			args: args{key: testSourceVia, from: "jump", steps: 1},
			want: "bastion",
		},
		{
			name: "an optional source cycles back to none",
			args: args{key: testSourceTransport, from: "ssh:prod", steps: 1},
			want: "",
		},
		{
			name: "a missing value moves onto a real one",
			args: args{key: testSourceVia, from: "old-bastion", steps: 1},
			want: "jump",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newFormModel(config.PostgresConnType, sourceFormSpec(), "edit", nil, true, nil, sources)
			m.vals[tt.args.key] = tt.args.from
			i := fieldIndex(t, m, tt.args.key)

			for range tt.args.steps {
				m.navCycle(i, 1)
			}

			assert.Equal(t, tt.want, m.vals[tt.args.key])
		})
	}
}
