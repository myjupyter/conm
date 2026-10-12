package spec

import (
	"github.com/myjupyter/conm/internal/config"
)

const tagsMaxLength = 255

// The section a connection form is split into, shared by every entity spec so
// the copy reads the same whatever database the form is for.
const (
	connectionSectionTitle = "connection"
	MetadataSectionTitle   = "meta"
)

const (
	// TextFieldKind is a default kind if isn't set.
	TextFieldKind FieldKind = iota
	HiddenFieldKind
	IntFieldKind
	SelectFieldKind
	SourceFieldKind
)

const (
	// RequiredFieldProperty is a default property if isn't set.
	RequiredFieldProperty FieldProperty = iota
	OptionalFieldProperty
)

var FormSpecs = map[config.ConnType]FormSpec[config.Connection]{
	config.PostgresConnType:   PostgresFormSpec,
	config.MySQLConnType:      MySQLFormSpec,
	config.MSSQLConnType:      MSSQLFormSpec,
	config.ClickHouseConnType: ClickHouseFormSpec,
	config.RedisConnType:      RedisFormSpec,
	config.MongoDBConnType:    MongoDBFormSpec,
	config.SSHConnType:        SSHFormSpec,
}

type FieldProperty uint8

type FieldKind uint8

type FormFieldKey = string

type FormFieldValue = string

type Sources map[string][]string

type FieldSource struct {
	Name        string
	VisibleFunc func(Sources) bool
}

func SourceAvailable(source string) func(Sources) bool {
	return func(sources Sources) bool {
		return len(sources[source]) > 0
	}
}

type FormField struct {
	Key          string
	Label        string
	Kind         FieldKind
	Property     FieldProperty
	Example      string
	DefaultValue string
	Options      []string // only used for SelectFieldKind
	OptionsFunc  func(map[FormFieldKey]FormFieldValue) []string
	Source       *FieldSource
	TextFunc     func(map[FormFieldKey]FormFieldValue) FieldText
	ValidateFunc func(string) error
}

type FieldText struct {
	Label   string
	Example string
}

type FormSection struct {
	Title  string
	Note   string
	Fields []FormFieldKey
}

type FormSpec[T any] struct {
	AddTitle  string
	EditTitle string
	Fields    []FormField
	Sections  []FormSection
	BuildFunc func(map[FormFieldKey]FormFieldValue) (T, error)
	SeedFunc  func(T) map[FormFieldKey]FormFieldValue
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
