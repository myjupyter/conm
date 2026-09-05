package cli

import (
	"context"
	"fmt"
	"slices"

	"github.com/myjupyter/conm/internal/config"
)

type Launcher struct {
	name string
	kind config.ConnType
}

func For(conm config.Conm, t config.ConnType) Launcher {
	return Launcher{name: conm.CLI(t), kind: t}
}

func (l Launcher) Run(ctx context.Context, cfg config.Connection, password string) error {
	if l.name == "" {
		return fmt.Errorf("%w for %s, run conm init", ErrNoClient, l.kind)
	}

	if !slices.Contains(Clients(l.kind), l.name) {
		return fmt.Errorf("%w: %s is not a %s client", ErrUnknownClient, l.name, l.kind)
	}

	build, ok := commands[l.name]
	if !ok {
		return fmt.Errorf("%w: %s has no command", ErrUnknownClient, l.name)
	}

	cmd, err := build(cfg, password)
	if err != nil {
		return err
	}

	return cmd.run(ctx, l.name)
}
