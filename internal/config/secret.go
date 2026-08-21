package config

type Secret interface {
	ID() string
	Provider() string
	Location() string
	Description() string
	Params() []SecretParam

	IsValid() bool
	Validate() []error
}

type SecretParam struct {
	Name  string
	Value string
}

type SecretType int

const (
	LiteralSecretType SecretType = iota + 1
	KeyringSecretType
)

func (t SecretType) String() string {
	switch t {
	case LiteralSecretType:
		return "literal"
	case KeyringSecretType:
		return "keyring"
	default:
		return "unknown"
	}
}
