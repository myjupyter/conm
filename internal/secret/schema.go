package secret

type Scheme = string

const (
	Literal Scheme = "literal"
	Keyring Scheme = "keyring"
)
