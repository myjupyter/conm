package spec

import (
	"strings"

	"github.com/myjupyter/conm/internal/config"
)

const tagsMaxLength = 255

const (
	// TextFieldKind is a default kind if isn't set
	TextFieldKind FieldKind = iota
	HiddenFieldKind
	IntFieldKind
	SelectFieldKind
)

const (
	// RequiredFieldProperty is a default property if isn't set
	RequiredFieldProperty FieldProperty = iota
	OptionalFieldProperty
)

var FormSpecs = map[config.ConnType]FormSpec{
	config.PostgresConnType: PostgresFormSpec,
}

type FieldProperty int

type FieldKind int

type FormFieldKey = string

type FormFieldValue = string

type FormField struct {
	Key          string
	Label        string
	Kind         FieldKind
	Example      string
	Property     FieldProperty
	DefaultValue string
	Options      []string // only used for SelectFieldKind
	ValidateFunc func(string) error
}

type FormSpec struct {
	AddTitle  string
	EditTitle string
	Fields    []FormField
	BuildFunc func(map[FormFieldKey]FormFieldValue) (config.ConnectionConfig, error)
	SeedFunc  func(config.ConnectionConfig) map[FormFieldKey]FormFieldValue
}

func (f FieldProperty) String() string {
	switch f {
	case RequiredFieldProperty:
		return "required"
	case OptionalFieldProperty:
		return "optional"
	default:
		return ""
	}
}

func parseTags(raw string) []string {
	var tags []string
	for t := range strings.SplitSeq(raw, ",") {
		if t = strings.TrimSpace(t); t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}
