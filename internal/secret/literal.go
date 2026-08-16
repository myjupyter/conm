package secret

import (
	"context"
	"strings"
)

var _ Provider = (*LiteralProvider)(nil)

type LiteralProvider struct{}

func (*LiteralProvider) Scheme() (Scheme, SchemeMethods) { return Literal, Resolve }

func (*LiteralProvider) Resolve(_ context.Context, ref Reference) (string, error) {
	return strings.TrimPrefix(ref.SecretRef(), "literal:"), nil
}

func (*LiteralProvider) Store(_ context.Context, _ Reference, _ string) error { return nil }

func (*LiteralProvider) Remove(_ context.Context, _ Reference) error { return nil }
