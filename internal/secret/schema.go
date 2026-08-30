package secret

type Scheme = string

const (
	Literal Scheme = "literal"
	Keyring Scheme = "keyring"
	None    Scheme = "none"
)

func IsStore(scheme Scheme) bool {
	return scheme != "" && scheme != Literal && scheme != None
}
