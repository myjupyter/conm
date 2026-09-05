package ui

import (
	"strings"

	"github.com/myjupyter/conm/internal/ui/spec"
)

type formValues = map[spec.FormFieldKey]spec.FormFieldValue

type formStatus struct {
	text string
	kind statusKind
}

func warnStatus(warnings []string) formStatus {
	if len(warnings) == 0 {
		return formStatus{}
	}

	return formStatus{text: strings.Join(warnings, " · "), kind: kindWarn}
}

type formResult[T any] struct {
	value  T
	values formValues
	ok     bool
}

type formLoop[T any] struct {
	open func(formValues, formStatus) (formResult[T], error)
	save func(T) error
}

func (l formLoop[T]) run(initial formValues, status formStatus) (bool, error) {
	for {
		res, err := l.open(initial, status)
		if err != nil || !res.ok {
			return false, err
		}

		saveErr := l.save(res.value)
		if saveErr == nil {
			return true, nil
		}

		initial = res.values
		status = formStatus{text: saveErr.Error(), kind: kindErr}
	}
}
