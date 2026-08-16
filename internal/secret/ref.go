package secret

import "strings"

// Reference is a config pointing at a secret: a bare password, or a
// "<scheme>:<location>" spec. Value receivers have to satisfy it — connection
// configs are stored and passed around by value, so SetSecretRef (pointer
// receiver) can't be part of it.
type Reference interface {
	SecretRef() string
}

type ref struct {
	scheme Scheme
	spec   string
	raw    string
}

func parseRef(raw string) (ref, bool) {
	scheme, spec, found := strings.Cut(raw, ":")
	if !found {
		return ref{}, false
	}

	return ref{scheme: Scheme(scheme), spec: spec, raw: raw}, true
}

// Key is how a stored secret is addressed: "<scheme>:<location>". It is both
// what a connection carries in its secret field and the identity of one entry
// in the secret store.
func Ref(scheme Scheme, location string) string {
	return scheme + ":" + location
}

// StoredKey returns the Key a raw spec points at, and whether it points at the
// store at all. Bare passwords and literal: values are kept inside the
// connection itself, so they have no entry to point at.
func RawRef(raw string) (string, bool) {
	r, ok := parseRef(raw)
	if !ok || r.scheme == Literal {
		return "", false
	}

	return Ref(r.scheme, r.spec), true
}
