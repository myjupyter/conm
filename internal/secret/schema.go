package secret

type Scheme = string

const (
	Global  Scheme = "global"
	Literal Scheme = "literal"
	Keyring Scheme = "keyring"
)

type SchemeMethods int

const (
	Resolve SchemeMethods = 1 << iota
	Store
	Remove
	Usages
)
