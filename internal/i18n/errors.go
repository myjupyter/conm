package i18n

import "errors"

var (
	errTranslationType   = errors.New("a translation must be a string or a table of plural forms")
	errMissingOtherForm  = errors.New(`a plural table must have the "other" form`)
	errUnknownPluralForm = errors.New("a plural form must be one of zero, one, two, few, many, other")
)
