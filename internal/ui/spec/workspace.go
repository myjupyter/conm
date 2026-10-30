package spec

import (
	"github.com/myjupyter/conm/internal/config"
)

const WorkspaceNameKey FormFieldKey = "name"

var WorkspaceFormSpec = FormSpec[string]{
	AddTitle:  "Add a workspace",
	EditTitle: "Rename a workspace",
	Fields: []FormField{
		{
			Key:          WorkspaceNameKey,
			Label:        "Name",
			Example:      "corp",
			ValidateFunc: config.ValidateWorkspaceName,
		},
	},
	Sections: []FormSection{
		{Title: "workspace", Note: "a folder of its own connections and secrets", Fields: []FormFieldKey{WorkspaceNameKey}},
	},
	BuildFunc: func(values map[FormFieldKey]FormFieldValue) (string, error) {
		return values[WorkspaceNameKey], nil
	},
	SeedFunc: func(name string) map[FormFieldKey]FormFieldValue {
		return map[FormFieldKey]FormFieldValue{WorkspaceNameKey: name}
	},
}
