package secret

import "context"

type Provider interface {
	Scheme() Scheme
	Resolve(ctx context.Context, ref Reference) (string, error)
	Store(ctx context.Context, ref Reference, password string) error
	Remove(ctx context.Context, ref Reference) error
}
