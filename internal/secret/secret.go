package secret

type Secretable interface {
	Secret() string
}
