package spec

import (
	"github.com/myjupyter/conm/internal/config"
)

const tagsMaxLength = 255

// The section a connection form is split into, shared by every entity spec so
// the copy reads the same whatever database the form is for.
const (
	connectionSectionTitle = "connection"
	connectionSectionNote  = "how conm reaches the server"
	MetadataSectionTitle   = "meta"
	metadataSectionNote    = "meta information about the connection"
)

const (
	// TextFieldKind is a default kind if isn't set.
	TextFieldKind FieldKind = iota
	HiddenFieldKind
	IntFieldKind
	SelectFieldKind
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
