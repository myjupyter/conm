package conn

import (
	"context"
	"database/sql"
	"os"
	"os/exec"

	"github.com/myjupyter/conm/internal/config"

	_ "github.com/jackc/pgx/stdlib"
)

type PGClient struct {
	conmCfg config.Conm
	cfg     config.Postgres

	db *sql.DB
}

func NewPGClient(
	conmConfig config.Conm,
	cfg config.Postgres,
) (*PGClient, error) {
	db, err := sql.Open("pgx", cfg.URL())
	if err != nil {
		return nil, err
	}

	return &PGClient{
		conmCfg: conmConfig,
		cfg:     cfg,
		db:      db,
	}, nil
}

func (p *PGClient) Name() string {
	return p.cfg.Name()
}

func (p *PGClient) Description() string {
	return p.cfg.Description()
}

func (p *PGClient) Tags() []string {
	return p.cfg.Tags()
}

func (p *PGClient) Host() string {
	return p.cfg.Host()
}

func (p *PGClient) Username() string {
	return p.cfg.Username()
}

func (p *PGClient) Port() int {
	return p.cfg.Port()
}

func (p *PGClient) Database() string {
	return p.cfg.Database()
}

func (p *PGClient) Ping(ctx context.Context) error {
	db, err := sql.Open("pgx", p.cfg.URL())
	if err != nil {
		return err
	}
	defer db.Close()

	return db.PingContext(ctx)
}

func (p *PGClient) Run(ctx context.Context) error {

	executor := exec.CommandContext(ctx, p.conmCfg.PostgresCli, p.cfg.URL())

	executor.Stdin = os.Stdin
	executor.Stdout = os.Stdout
	executor.Stderr = os.Stderr

	if err := executor.Run(); err != nil {
		return err
	}

	return nil
}

func (p *PGClient) Close() error {
	return p.db.Close()
}
