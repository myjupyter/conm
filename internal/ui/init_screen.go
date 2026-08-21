package ui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/myjupyter/conm/internal/config"
)

type InitPostgresScreenConfig struct {
	UsePGPassImportValue bool
}

type InitScreenConfig struct {
	ConnType       config.ConnType
	CLIFlagValue   string
	PostgresConfig *InitPostgresScreenConfig
}

func RunInitScreen(cfg InitScreenConfig) error {
	if err := config.CreateConmConfigPath(); err != nil {
		return err
	}

	if cfg.CLIFlagValue != "" {
		if err := config.ValidateCLI(cfg.ConnType, cfg.CLIFlagValue); err != nil {
			return err
		}
	}

	clies := config.DetectCLI(cfg.ConnType)

	var conmConfig config.Conm
	switch {
	case cfg.CLIFlagValue != "":
		finalCLI, found := config.FindCLI(clies, cfg.CLIFlagValue)
		if !found {
			cliNames := make([]string, 0, len(clies))
			for _, cli := range clies {
				cliNames = append(cliNames, cli.Name)
			}
			return fmt.Errorf("unknown %s client %q; choose one of: %s", cfg.ConnType, cfg.CLIFlagValue, strings.Join(cliNames, ", "))
		}
		if !finalCLI.Exists {
			return fmt.Errorf("%s client %q is not installed", cfg.ConnType, cfg.CLIFlagValue)
		}
		conmConfig.SetCLI(cfg.ConnType, cfg.CLIFlagValue)
	default:
		hasClient := false
		for _, cli := range clies {
			if cli.Exists {
				hasClient = true
				break
			}
		}
		if !hasClient {
			return fmt.Errorf("no known postgres client found; install one of: %s", strings.Join(config.PGClients, ", "))
		}

		cli, chosen, err := RunSelectCLIForm(clies)
		if err != nil {
			return err
		}
		if !chosen {
			return errors.New("no postgres client selected")
		}

		conmConfig.SetCLI(cfg.ConnType, cli.Name)
	}

	if err := runConnTypeSpecificInit(cfg); err != nil {
		return err
	}

	c, err := config.OpenConfig[*config.ConmConfigWrapper](config.ConmPath())
	if err != nil {
		return fmt.Errorf("couldn't open conm config file: %w", err)
	}

	defer c.Close()

	// conm.toml holds a single [conm] section, not a list.
	c.Put(0, conmConfig)

	if err := c.Save(); err != nil {
		return fmt.Errorf("couldn't save conm config file: %w", err)
	}

	return nil
}

func runConnTypeSpecificInit(cfg InitScreenConfig) error {
	switch cfg.ConnType {
	case config.PostgresConnType:
		return runPostgresInit(cfg.PostgresConfig)
	default:
		return fmt.Errorf("unknown connection type %q", cfg.ConnType)
	}
}

func runPostgresInit(cfg *InitPostgresScreenConfig) error {
	var connCfgs []config.Postgres
	switch cfg.UsePGPassImportValue {
	case true:
		pgPassCreds, err := config.ImportFromPGPass()
		if err != nil {
			return fmt.Errorf("couldn't import from .pgpass: %w", err)
		}

		connCfgs = pgPassCreds
	default:
		opt, chosen, err := RunSelectImport()
		if err != nil {
			return err
		}
		if !chosen {
			return errors.New("init cancelled")
		}

		switch opt {
		case importSelectPGPass:
			pgPassCreds, err := config.ImportFromPGPass()
			if err != nil {
				return fmt.Errorf("couldn't import from .pgpass: %w", err)
			}

			connCfgs = pgPassCreds
		case importSelectSkip:
			// Nothing to import; the config file is created empty.
		case importSelectAdd:
			conf, added, err := runAddForm(config.PostgresConnType, nil)
			if err != nil {
				return err
			}
			if added {
				pg, ok := conf.(config.Postgres)
				if !ok {
					return fmt.Errorf("unexpected connection type %T for postgres add form", conf)
				}
				connCfgs = append(connCfgs, pg)
			}
		}
	}

	c, err := config.OpenConfig[*config.PostgresConfigWrapper](config.PostgresPath())
	if err != nil {
		return fmt.Errorf("couldn't open postgres config file: %w", err)
	}

	defer c.Close()

	for i := range connCfgs {
		c.Add(connCfgs[i])
	}

	if err := c.Save(); err != nil {
		return fmt.Errorf("couldn't save postgres config file: %w", err)
	}

	return nil
}
