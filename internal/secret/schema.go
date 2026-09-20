package secret

type Scheme = string

const (
	Literal  Scheme = "literal"
	Keyring  Scheme = "keyring"
	None     Scheme = "none"
	Filepath Scheme = "filepath"
)

func IsStore(scheme Scheme) bool {
	return scheme != "" && scheme != Literal && scheme != None && scheme != Filepath
}

func IsMaterial(scheme Scheme) bool {
	return scheme == "" || scheme == Literal
}
