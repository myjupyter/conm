package ui

import (
	"fmt"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui/spec"
)

func RunEditForm(t config.ConnType, existing config.Connection) (config.Connection, bool, error) {
	spec, ok := spec.FormSpecs[t]
	if !ok {
		return nil, false, fmt.Errorf("edit form is not implemented for connection type %q", t)
	}
	if spec.SeedFunc == nil {
		return nil, false, fmt.Errorf("edit form is not seedable for connection type %q", t)
	}

	return runForm(newFormModel(spec, spec.EditTitle, spec.SeedFunc(existing)))
}
