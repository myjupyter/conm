package secret

import "strings"

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
