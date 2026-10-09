package transport

import (
	"context"

	"github.com/myjupyter/conm/internal/config"
)

var _ Transport = (*Direct)(nil)

type Direct struct {
	target config.Connection
}

func NewDirect(target config.Connection) *Direct {
	return &Direct{target: target}
}

func (d *Direct) Open(context.Context) (string, int, error) {
	return d.target.Host(), d.target.Port(), nil
}

func (*Direct) Close(context.Context) error {
	return nil
}
