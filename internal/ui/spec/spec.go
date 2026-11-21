package spec

import (
	"github.com/myjupyter/conm/internal/config"
)

const (
	connectionSectionTitle = "connection"
	MetadataSectionTitle   = "meta"
)

const (
	TextFieldKind FieldKind = iota
	HiddenFieldKind
	IntFieldKind
	SelectFieldKind
	SourceFieldKind
	StringArrayFieldKind
)

const (
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
	DefaultFunc  func() string
	Options      []string
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
