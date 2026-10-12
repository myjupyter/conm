package transport

import (
	"context"
	"fmt"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

type Transport interface {
	Open(ctx context.Context) (host string, port int, err error)
	Close(ctx context.Context) error
}

func Through(kind config.TransportKind) (config.ConnType, bool) {
	switch kind {
	case config.SSHTransportKind:
		return config.SSHConnType, true
	default:
		return 0, false
	}
}

func shareKey(rec config.Transport, target config.Connection) string {
	switch rec.Kind {
	case config.SSHTransportKind:
		return rec.Identity() + " to " + address(target)
	default:
		return rec.Identity()
	}
}

func New(
	launcher launcher,
	rec config.Transport,
	via config.Connection,
	target config.Connection,
	sec secret.Provider,
) (Transport, error) {
	switch rec.Kind {
	case config.SSHTransportKind:
		bastion, ok := via.(config.SSH)
		if !ok {
			return nil, fmt.Errorf("transport %q goes through ssh, not %s", rec.Name, via.ConnType())
		}
		return NewSSH(launcher, bastion, rec, target, sec), nil
	default:
		return nil, fmt.Errorf("transport %q: unknown kind %q", rec.Name, rec.Kind)
	}
}
