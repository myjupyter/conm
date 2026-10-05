package ui

import (
	"maps"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/myjupyter/conm/internal/ui/spec"
)

func selectCombinations(fields []spec.FormField) []map[spec.FormFieldKey]spec.FormFieldValue {
	combos := []map[spec.FormFieldKey]spec.FormFieldValue{{}}
	for _, f := range fields {
		if f.Kind != spec.SelectFieldKind || len(f.Options) == 0 {
			continue
		}
		var next []map[spec.FormFieldKey]spec.FormFieldValue
		for _, combo := range combos {
			for _, opt := range f.Options {
				extended := map[spec.FormFieldKey]spec.FormFieldValue{f.Key: opt}
				maps.Copy(extended, combo)
				next = append(next, extended)
			}
		}
		combos = next
	}
	return combos
}

func labelsOf(fields []spec.FormField) []string {
	var labels []string
	for _, f := range fields {
		labels = append(labels, f.Label)
		if f.LabelFunc == nil {
			continue
		}
		for _, values := range selectCombinations(fields) {
			labels = append(labels, f.LabelFunc(values))
		}
	}
	return labels
}

func TestFormLabelsFit(t *testing.T) {
	check := func(t *testing.T, fields []spec.FormField) {
		t.Helper()
		for _, label := range labelsOf(fields) {
			assert.LessOrEqualf(t, len([]rune(label+" *")), formLabelW, "label %q is cut", strings.ToLower(label))
		}
	}

	for kind, sp := range spec.FormSpecs {
		t.Run(kind.String(), func(t *testing.T) { check(t, sp.Fields) })
	}
	for scheme, sp := range secretFormSpecs {
		t.Run("secret "+scheme, func(t *testing.T) { check(t, sp.Fields) })
	}
}
