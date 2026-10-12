package transport

import (
	"strings"

	"github.com/myjupyter/conm/internal/config"
)

type Reference interface {
	TransportRef() string
}

func Ref(kind config.TransportKind, name string) string {
	return kind + ":" + name
}

func ParseRef(raw string) (config.TransportKind, string, bool) {
	kind, name, found := strings.Cut(raw, ":")
	if !found || kind == "" || name == "" {
		return "", "", false
	}

	return kind, name, true
}
