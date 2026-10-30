package ui

import (
	"context"
	"fmt"
	"strings"
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
		keyMap = newKeymap(cfg.Keybindings)

		m, err := runConnections(cfg, active)
		if err != nil || !m.reload {
			return err
		}

		cfg, err = config.ReadConm(config.ConmPath())
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
	secrets, transports := view.NewSecrets(ws.Keyring), view.NewTransports(ws.Transports)
	if conns.Total() == 0 {
		first, saved, err := runFirstConnection(connectionFormContext(conns, secrets, transports), conns.Add)
		if err != nil || !saved {
			return Model{}, err
		}
		active = first
	}
	conns.SetActive(active)

	table := New(conns, secrets, transports)
	if err := cfg.Keybindings.Validate(); err != nil {
		table.status, table.statusKind = "keybindings in conm.toml ignored, defaults in use · "+firstLine(err), kindWarn
	}

	m, err := tea.NewProgram(table).Run()
	if err != nil {
		return Model{}, err
	}

	cm, ok := m.(Model)
	if !ok {
		return Model{}, fmt.Errorf("connections screen returned an unexpected model %T", m)
	}

	return cm, nil
}

func runFirstConnection(ctx formContext, save func(config.Connection) error) (config.ConnType, bool, error) {
	for {
		t, picked, err := runTypePicker(newInitModel())
		if err != nil || !picked {
			return 0, false, err
		}

		saved, err := runAddConnection(t, ctx, save)
		if err != nil || saved {
			return t, saved, err
		}
	}
}

func Connect(ctx context.Context, conm config.Conm, t config.ConnType, name string, query ...string) error {
	ws, conn, err := openNamedConnection(conm, t, name)
	if err != nil {
		return err
	}
	defer ws.Close()

	return conn.Run(ctx, query...)
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

	i, _ := repo.IndexOf(name)
	conn, ok := repo.ConnectionAt(i)
	if !ok {
		_ = ws.Close()
		return nil, nil, noConnectionNamed(repo, name)
	}

	return ws, conn, nil
}

func Connections(conm config.Conm, t config.ConnType) ([]config.Connection, error) {
	ws, repo, err := openConnections(conm, t)
	if err != nil {
		return nil, err
	}
	defer ws.Close()

	conns := make([]config.Connection, 0, repo.Len())
	for i := range repo.Len() {
		if cfg, ok := repo.ConfigAt(i); ok {
			conns = append(conns, cfg)
		}
	}

	return conns, nil
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
		return nil, nil, fmt.Errorf("unknown connection type %q", t)
	}

	return ws, repo, nil
}

func namedConnection(repo repository.Connections, name string) (int, config.Connection, error) {
	i, _ := repo.IndexOf(name)
	cfg, ok := repo.ConfigAt(i)
	if !ok {
		return 0, nil, noConnectionNamed(repo, name)
	}

	return i, cfg, nil
}

func noConnectionNamed(repo repository.Connections, name string) error {
	return fmt.Errorf("no %s connection named %q", repo.Kind(), name)
}

func firstLine(err error) string {
	line, _, _ := strings.Cut(err.Error(), "\n")
	return line
}
