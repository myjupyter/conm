package secret

import (
	"context"
	"strings"
)

var _ Provider = (*LiteralProvider)(nil)

type LiteralProvider struct{}

func (*LiteralProvider) Scheme() (Scheme, SchemeMethods) { return Literal, Resolve }

func (*LiteralProvider) Resolve(_ context.Context, secret Secretable) (string, error) {
	return strings.TrimPrefix(secret.Secret(), "literal:"), nil
}

func (*LiteralProvider) Store(_ context.Context, _ Secretable, _ string) error { return nil }

func (*LiteralProvider) Remove(_ context.Context, _ Secretable) error { return nil }

func (p *LiteralProvider) Track(_ Secretable) {}

func (p *LiteralProvider) Untrack(_ Secretable) {}

func (p *LiteralProvider) Usages(_ Scheme) []Usage { return nil }
