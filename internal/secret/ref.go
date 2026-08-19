package secret

import "strings"

// Reference is a config pointing at a secret: a bare password, or a
// "<scheme>:<location>" spec. Value receivers have to satisfy it — connection
// configs are stored and passed around by value, so SetSecretRef (pointer
// receiver) can't be part of it.
type Reference interface {
	SecretRef() string
}

func Ref(scheme Scheme, location string) string {
	return scheme + ":" + location
}

type ref struct {
	scheme Scheme
	spec   string
	raw    string
}

func ParseRef(raw string) (Scheme, string, bool) {
	r, ok := parseRef(raw)
	return r.scheme, r.spec, ok
}

func parseRef(raw string) (ref, bool) {
	scheme, spec, found := strings.Cut(raw, ":")
	if !found {
		return ref{}, false
	}

	return ref{scheme: Scheme(scheme), spec: spec, raw: raw}, true
}
