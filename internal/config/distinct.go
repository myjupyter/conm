package config

import "fmt"

type Distinct interface {
	Identity() string
}

func identity(parts ...string) string {
	return fmt.Sprintf("%q", parts)
}
