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

func (r *Resolver) Scheme() (Scheme, SchemeMethods) {
	return Global, Resolve | Store | Remove | Usages
}

func (r *Resolver) Resolve(ctx context.Context, spec Secretable) (string, error) {
	raw := spec.Secret()

	p, ref, ok := r.provider(raw)
	if !ok {
		return raw, nil
	}

	password, err := p.Resolve(ctx, spec)
	if err != nil {
		return "", fmt.Errorf("resolve %q secret: %w", ref.scheme, err)
	}
	return password, nil
}

func (r *Resolver) Store(ctx context.Context, spec Secretable, password string) error {
	p, ref, ok := r.provider(spec.Secret())
	if !ok {
		return nil
	}

	if err := p.Store(ctx, spec, password); err != nil {
		return fmt.Errorf("store %q secret: %w", ref.scheme, err)
	}
	return nil
}

func (r *Resolver) Remove(ctx context.Context, spec Secretable) error {
	p, ref, ok := r.provider(spec.Secret())
	if !ok {
		return nil
	}

	if err := p.Remove(ctx, spec); err != nil {
		return fmt.Errorf("remove %q secret: %w", ref.scheme, err)
	}
	return nil
}

func (r *Resolver) Track(spec Secretable) {
	if p, _, ok := r.provider(spec.Secret()); ok {
		p.Track(spec)
	}
}

func (r *Resolver) Untrack(s Secretable) {
	if p, _, ok := r.provider(s.Secret()); ok {
		p.Untrack(s)
	}
}

func (r *Resolver) Usages(scheme Scheme) []Usage {
	p, ok := r.providers[scheme]
	if !ok {
		return nil
	}

	return p.Usages(scheme)
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
	scheme, _ := p.Scheme()
	r.providers[scheme] = p
}
