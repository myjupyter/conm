package secret

import "context"

type Provider interface {
	Scheme() (Scheme, SchemeMethods)
	Resolve(ctx context.Context, spec Secretable) (string, error)
	Store(ctx context.Context, spec Secretable, password string) error
	Remove(ctx context.Context, spec Secretable) error
	Track(spec Secretable)
	Untrack(spec Secretable)
	Usages(Scheme) []Usage
}
