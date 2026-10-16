package ui

import (
	"maps"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui/spec"
)

const (
	testSourceHost      = "host"
	testSourceVia       = "via"
	testSourceType      = "type"
	testSourceTransport = "transport"
	testSourceDirect    = "direct"
)

type formVals = map[spec.FormFieldKey]spec.FormFieldValue

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
				Key:          testSourceType,
				Label:        "Type",
				Kind:         spec.SourceFieldKind,
				DefaultValue: testSourceDirect,
				Source:       &spec.FieldSource{Name: "types"},
			},
			{
				Key:     testSourceTransport,
				Label:   "Transport",
				Kind:    spec.SourceFieldKind,
				Example: "connects directly",
				Source: &spec.FieldSource{NameFunc: func(values formVals) string {
					if values[testSourceType] == testSourceDirect {
						return ""
					}
					return "transports " + values[testSourceType]
				}},
			},
		},
		Sections: []spec.FormSection{
			{Title: "connection", Fields: []spec.FormFieldKey{testSourceHost, testSourceVia}},
			{Title: "transport", Fields: []spec.FormFieldKey{testSourceType, testSourceTransport}},
		},
	}
}

func testSources() spec.Sources {
	return spec.Sources{
		"ssh":            {"bastion", "jump"},
		"types":          {testSourceDirect, "ssh"},
		"transports ssh": {"prod", "staging"},
	}
}

func newSourceForm(sources spec.Sources, initial formVals) formModel {
	return newFormModel(config.PostgresConnType, sourceFormSpec(), "edit", initial, initial != nil, formContext{sources: sources})
}

func fieldIndex(t *testing.T, m formModel, key spec.FormFieldKey) int {
	t.Helper()

	for i := range m.fields {
		if m.fields[i].Key == key {
			return i
		}
	}
	require.Failf(t, "no field", "field %q is not in the form", key)
	return -1
}

func TestFormModel_options_source(t *testing.T) {
	t.Parallel()

	type args struct {
		key  spec.FormFieldKey
		kind string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "a source offers its values",
			args: args{key: testSourceVia, kind: testSourceDirect},
			want: []string{"bastion", "jump"},
		},
		{
			name: "a source named after another field follows it",
			args: args{key: testSourceTransport, kind: "ssh"},
			want: []string{"prod", "staging"},
		},
		{
			name: "a source that does not apply offers nothing",
			args: args{key: testSourceTransport, kind: testSourceDirect},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newSourceForm(testSources(), nil)
			m.vals[testSourceType] = tt.args.kind

			assert.Equal(t, tt.want, m.options(m.fields[fieldIndex(t, m, tt.args.key)]))
		})
	}
}

func TestNewFormModel_source(t *testing.T) {
	t.Parallel()

	type args struct {
		initial formVals
		sources spec.Sources
	}
	tests := []struct {
		name          string
		args          args
		wantVia       string
		wantType      string
		wantTransport string
		wantSections  []string
	}{
		{
			name:         "a new connection goes direct and the transport section is still there",
			args:         args{sources: testSources()},
			wantVia:      "bastion",
			wantType:     testSourceDirect,
			wantSections: []string{"connection", "transport"},
		},
		{
			name: "an edited connection keeps its transport",
			args: args{
				initial: formVals{testSourceType: "ssh", testSourceTransport: "staging"},
				sources: testSources(),
			},
			wantVia:       "bastion",
			wantType:      "ssh",
			wantTransport: "staging",
			wantSections:  []string{"connection", "transport"},
		},
		{
			name:         "with no transports at all the section only offers direct",
			args:         args{sources: spec.Sources{"ssh": {"bastion"}, "types": {testSourceDirect}}},
			wantVia:      "bastion",
			wantType:     testSourceDirect,
			wantSections: []string{"connection", "transport"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := newSourceForm(tt.args.sources, tt.args.initial)

			titles := make([]string, 0, len(got.sections))
			for _, s := range got.sections {
				titles = append(titles, s.Title)
			}
			is := assert.New(t)
			is.Equal(tt.wantVia, got.vals[testSourceVia])
			is.Equal(tt.wantType, got.vals[testSourceType])
			is.Equal(tt.wantTransport, got.vals[testSourceTransport])
			is.Equal(tt.wantSections, titles)
		})
	}
}

func TestFormModel_fieldError_source(t *testing.T) {
	t.Parallel()

	type args struct {
		values formVals
		key    spec.FormFieldKey
	}
	tests := []struct {
		name        string
		args        args
		wantErr     string
		wantMissing bool
	}{
		{
			name: "a value from the source",
			args: args{values: formVals{testSourceVia: "jump"}, key: testSourceVia},
		},
		{
			name:        "a value the source no longer has",
			args:        args{values: formVals{testSourceVia: "old-bastion"}, key: testSourceVia},
			wantErr:     "old-bastion is no longer available",
			wantMissing: true,
		},
		{
			name: "a transport is not asked for when the connection goes direct",
			args: args{values: formVals{testSourceType: testSourceDirect, testSourceTransport: ""}, key: testSourceTransport},
		},
		{
			name:    "a transport type needs a transport",
			args:    args{values: formVals{testSourceType: "ssh", testSourceTransport: ""}, key: testSourceTransport},
			wantErr: "transport is required",
		},
		{
			name:        "a transport that is gone",
			args:        args{values: formVals{testSourceType: "ssh", testSourceTransport: "old"}, key: testSourceTransport},
			wantErr:     "old is no longer available",
			wantMissing: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newSourceForm(testSources(), nil)
			maps.Copy(m.vals, tt.args.values)
			i := fieldIndex(t, m, tt.args.key)

			got := m.fieldError(i)

			assert.Equal(t, tt.wantMissing, m.isMissing(m.fields[i]))
			assert.NotEmpty(t, m.render())
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
		key  spec.FormFieldKey
		from formVals
	}
	tests := []struct {
		name          string
		args          args
		wantVia       string
		wantType      string
		wantTransport string
	}{
		{
			name:     "steps through the source values",
			args:     args{key: testSourceVia, from: formVals{testSourceVia: "bastion"}},
			wantVia:  "jump",
			wantType: testSourceDirect,
		},
		{
			name:     "a missing value moves onto a real one",
			args:     args{key: testSourceVia, from: formVals{testSourceVia: "old-bastion"}},
			wantVia:  "jump",
			wantType: testSourceDirect,
		},
		{
			name:          "picking a transport type offers its first transport",
			args:          args{key: testSourceType, from: formVals{testSourceType: testSourceDirect}},
			wantVia:       "bastion",
			wantType:      "ssh",
			wantTransport: "prod",
		},
		{
			name:     "going back to direct drops the transport",
			args:     args{key: testSourceType, from: formVals{testSourceType: "ssh", testSourceTransport: "staging"}},
			wantVia:  "bastion",
			wantType: testSourceDirect,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := newSourceForm(testSources(), nil)
			maps.Copy(m.vals, tt.args.from)

			m.navCycle(fieldIndex(t, m, tt.args.key), 1)

			is := assert.New(t)
			is.Equal(tt.wantVia, m.vals[testSourceVia])
			is.Equal(tt.wantType, m.vals[testSourceType])
			is.Equal(tt.wantTransport, m.vals[testSourceTransport])
		})
	}
}
