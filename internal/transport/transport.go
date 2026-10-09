package transport

import (
	"context"
)

type Transport interface {
	Open(ctx context.Context) (host string, port int, err error)
	Close(ctx context.Context) error
}
