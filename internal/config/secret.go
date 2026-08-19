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
