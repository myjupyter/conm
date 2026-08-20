package secret

import (
	"context"
	"fmt"
)

var _ Provider = (*Resolver)(nil)

type Resolver struct {
	providers map[Scheme]Provider
}

func Default() *Resolver {
	return new(
		&KeyringProvider{},
		&LiteralProvider{},
	)
}

func new(providers ...Provider) *Resolver {
	r := &Resolver{providers: make(map[Scheme]Provider, len(providers))}
	for _, p := range providers {
		r.register(p)
	}
	return r
}

// Scheme is empty: the resolver is not addressable by a scheme of its own, it
// dispatches to the provider a reference names.
func (r *Resolver) Scheme() Scheme { return "" }

func (r *Resolver) Resolve(ctx context.Context, ref Reference) (string, error) {
	raw := ref.SecretRef()

	p, parsed, ok := r.provider(raw)
	if !ok {
		return raw, nil
	}

	password, err := p.Resolve(ctx, ref)
	if err != nil {
		return "", fmt.Errorf("resolve %q secret: %w", parsed.scheme, err)
	}
	return password, nil
}

func (r *Resolver) Store(ctx context.Context, ref Reference, password string) error {
	p, parsed, ok := r.provider(ref.SecretRef())
	if !ok {
		return nil
	}

	if err := p.Store(ctx, ref, password); err != nil {
		return fmt.Errorf("store %q secret: %w", parsed.scheme, err)
	}
	return nil
}

func (r *Resolver) Remove(ctx context.Context, ref Reference) error {
	p, parsed, ok := r.provider(ref.SecretRef())
	if !ok {
		return nil
	}

	if err := p.Remove(ctx, ref); err != nil {
		return fmt.Errorf("remove %q secret: %w", parsed.scheme, err)
	}
	return nil
}

func (r *Resolver) provider(raw string) (Provider, ref, bool) {
	if parsed, ok := parseRef(raw); ok {
		if p, ok := r.providers[parsed.scheme]; ok {
			return p, parsed, true
		}
	}

	p, ok := r.providers[Literal]
	if !ok {
		return nil, ref{}, false
	}

	return p, ref{scheme: Literal, spec: raw, raw: raw}, true
}

func (r *Resolver) register(p Provider) {
	scheme := p.Scheme()
	if scheme == "" {
		// A provider without a scheme of its own (another Resolver) is not
		// reachable through a reference; registering it would shadow nothing.
		return
	}

	r.providers[scheme] = p
}
