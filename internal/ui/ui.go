package ui

import (
	"context"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/network"
	"github.com/myjupyter/conm/internal/repository"
	"github.com/myjupyter/conm/internal/ui/view"
)

func Run(cfg config.Conm) error {
	var anyConnection config.ConnType

	return RunOn(cfg, anyConnection)
}

func RunOn(cfg config.Conm, active config.ConnType) error {
	for {
		if enabledConnections(cfg) == 0 {
			return RunInit()
		}

		m, err := runConnections(cfg, active)
		if err != nil || !m.reload {
			return err
		}

		cfg, err = readConm()
		if err != nil {
			return err
		}
		active = m.active
	}
}

func runConnections(cfg config.Conm, active config.ConnType) (Model, error) {
	ws, err := repository.NewWorkspace(cfg)
	if err != nil {
		return Model{}, err
	}
	defer ws.Close()

	conns := view.NewConnections(ws.Connections()...)
	conns.SetActive(active)

	m, err := tea.NewProgram(New(conns, view.NewSecrets(ws.Keyring), view.NewTransports(ws.Transports))).Run()
	if err != nil {
		return Model{}, err
	}

	cm, ok := m.(Model)
	if !ok {
		return Model{}, fmt.Errorf("connections screen returned an unexpected model %T", m)
	}

	return cm, nil
}

func enabledConnections(cfg config.Conm) int {
	n := 0
	for _, conn := range cfg.Connections {
		if conn.Enabled {
			n++
		}
	}
	return n
}

func Connect(ctx context.Context, conm config.Conm, t config.ConnType, name string) error {
	ws, conn, err := openNamedConnection(conm, t, name)
	if err != nil {
		return err
	}
	defer ws.Close()

	return conn.Run(ctx)
}

func Ping(ctx context.Context, conm config.Conm, t config.ConnType, name string) (time.Duration, error) {
	ws, conn, err := openNamedConnection(conm, t, name)
	if err != nil {
		return 0, err
	}
	defer ws.Close()

	result, err := conn.Ping(ctx)
	if err != nil {
		return 0, err
	}

	return result.PingTime, nil
}

func openNamedConnection(
	conm config.Conm,
	t config.ConnType,
	name string,
) (*repository.Workspace, network.Connection, error) {
	ws, repo, err := openConnections(conm, t)
	if err != nil {
		return nil, nil, err
	}

	i, _, err := namedConnection(repo, name)
	if err != nil {
		_ = ws.Close()
		return nil, nil, err
	}

	conn, ok := repo.ConnectionAt(i)
	if !ok {
		_ = ws.Close()
		return nil, nil, fmt.Errorf("no %s connection named %q", t, name)
	}

	return ws, conn, nil
}

func RemoveConnection(conm config.Conm, t config.ConnType, name string) error {
	ws, repo, err := openConnections(conm, t)
	if err != nil {
		return err
	}
	defer ws.Close()

	i, _, err := namedConnection(repo, name)
	if err != nil {
		return err
	}

	return repo.Remove(i)
}

func openConnections(conm config.Conm, t config.ConnType) (*repository.Workspace, repository.Connections, error) {
	ws, err := repository.NewWorkspace(conm)
	if err != nil {
		return nil, nil, err
	}

	repo, ok := ws.ConnectionsOf(t)
	if !ok {
		_ = ws.Close()
		return nil, nil, fmt.Errorf("%q is not enabled\nrun 'conm add' to enable it", t)
	}

	return ws, repo, nil
}

func namedConnection(repo repository.Connections, name string) (int, config.Connection, error) {
	for i := range repo.Len() {
		if cfg, ok := repo.ConfigAt(i); ok && cfg.Meta().Name == name {
			return i, cfg, nil
		}
	}

	return 0, nil, fmt.Errorf("no %s connection named %q", repo.Kind(), name)
}
