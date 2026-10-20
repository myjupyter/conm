package spec

import (
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/i18n"
)

const tagsMaxLength = 255

// The section a connection form is split into, shared by every entity spec so
// the copy reads the same whatever database the form is for.
var (
	connectionSectionTitle = i18n.Sprintf("connection")
	MetadataSectionTitle   = i18n.Sprintf("meta")
)

const (
	// TextFieldKind is a default kind if isn't set.
	TextFieldKind FieldKind = iota
	HiddenFieldKind
	IntFieldKind
	SelectFieldKind
	SourceFieldKind
	StringArrayFieldKind
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
	Name     string
	NameFunc func(map[FormFieldKey]FormFieldValue) string
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
		return i18n.Sprintf("required")
	case OptionalFieldProperty:
		return i18n.Sprintf("optional")
	default:
		return ""
	}
}
