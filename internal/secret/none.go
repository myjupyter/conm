package secret

import "context"

var _ Provider = (*NoneProvider)(nil)

type NoneProvider struct{}

func (*NoneProvider) Scheme() Scheme { return None }

func (*NoneProvider) Resolve(_ context.Context, _ Reference) (string, error) { return "", nil }

func (*NoneProvider) Store(_ context.Context, _ Reference, _ string) error { return nil }

func (*NoneProvider) Remove(_ context.Context, _ Reference) error { return nil }
